# Importing and exporting VDI data

How to get disk data in and out of an XCP-ng/XenServer host through XenAPI -
what the API actually exposes, which formats it accepts, and the sharp edges
this fork's development hit against a live XCP-ng 8.3 host. Each claim is
checked against the XAPI source and/or a live host, noted inline.

## TL;DR

- There is **no XML-RPC method that carries disk bytes**. Bulky data always
  moves over HTTP; XML-RPC is only used to set up the objects around it.
- For VDI data, the endpoints are `PUT /import_raw_vdi` and
  `GET /export_raw_vdi`.
- Accepted `format` values are **`raw`, `vhd`, `tar`, `qcow2`** (`raw` is the
  default). The qcow2 value is spelled `qcow2`, not `qcow`.
- The file being imported does **not** have to live on the host: the client can
  stream it from anywhere (push). Whole-VM XVA import can also be pulled by
  the host from a URL.
- Only `raw` can auto-create a VDI, and only `raw` tolerates chunked transfer.
  `vhd`/`qcow2` need a pre-created VDI.
- A **fixed** VHD is rejected by the import path; strip its 512-byte footer and
  import as `raw`, or convert it with `qemu-img` first. Details in
  [Fixed vs. dynamic VHD](#fixed-vs-dynamic-vhd).
- qcow2 is available on XCP-ng 8.3 (and accepted by xapi as a wire format), but
  the on-SR *image format* is a separate, older concept - see
  [QCOW2](#qcow2-image-format-vs-wire-format).
- Whole-VM archives are a separate topic: **neither OVA nor VMDK is a VDI wire
  format**, and the `tar` value is not the OVF container. XAPI itself parses
  only **XVA**, which is imported via `PUT /import` (or `VM.Import` pull) - see
  [Whole-VM archives: OVA and XVA](#whole-vm-archives-ova-and-xva).

## Where the data actually flows

Two independent layers are easy to conflate:

| Layer | What it is | Method |
| --- | --- | --- |
| XML-RPC | object setup: `VDI.create`, `VBD.create`, `Task.create`, ... | typed methods in this package |
| HTTP | the disk bytes themselves | `/import_raw_vdi`, `/export_raw_vdi` |

This package binds the XML-RPC layer completely. It deliberately has **no**
binding for the HTTP endpoints - those are plain HTTP requests you make
yourself, reusing the session reference obtained from
`Session.LoginWithPassword` as the `session_id` query parameter.

The one typed bulk-transfer method in the package is
`VM.Import(url, sr, fullRestore, force)` - "Import an XVA from a URI" - where
the **host** fetches the archive from a URL it can reach. That is a *pull*, and
is distinct from the VDI upload endpoints in this document.

So, on direction:

- **Push** - your process streams the file to the host:
  `PUT /import_raw_vdi`.
- **Pull** - the host fetches from a URL you give it: `VM.Import` (XVA),
  `VM.ImportConvert` (converts a remote disk into a VM).

Neither requires the file to be on the host beforehand. The `xe vdi-import`
command only looks local because the `xe` CLI runs *on* the host and reads a
path there - that is a property of the CLI, not of the API.

## The VDI endpoints

Constants: `import_raw_vdi_uri` = `"/import_raw_vdi"`,
`export_raw_vdi_uri` = `"/export_raw_vdi"`
([`ocaml/xapi-consts/constants.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi-consts/constants.ml)).

Query parameters read by the handlers:

| Parameter | Meaning |
| --- | --- |
| `session_id` | XenAPI session ref; the same value returned by `Session.LoginWithPassword`. Basic auth and client-cert auth are also accepted. |
| `task_id` | Optional. If given, the work is attached to that existing task. |
| `vdi` | Target/source VDI, as an opaque ref **or** a UUID. |
| `sr_id` / `sr_uuid` | Target SR, as an opaque ref or UUID. Used when `vdi` is absent. |
| `format` | `raw` (default), `vhd`, `tar` or `qcow2`. |
| `chunked` | Presence only; requests chunked transfer encoding. Raw only. |

Handlers: [`ocaml/xapi/import_raw_vdi.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi/import_raw_vdi.ml),
[`ocaml/xapi/export_raw_vdi.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi/export_raw_vdi.ml).
The URI/argument declaration is in
[`ocaml/idl/datamodel.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/idl/datamodel.ml)
(`put_import_raw_vdi`, `get_export_raw_vdi`), and session/task handling in
[`ocaml/xapi/xapi_http.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi/xapi_http.ml)
(`assert_credentials_ok`, `with_context`).

## Wire formats

Defined by `Importexport.Format` in
[`ocaml/xapi/importexport.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi/importexport.ml):

| `format` | Content-Type | Auto-create VDI from `sr_uuid`? | Chunked? | Notes |
| --- | --- | --- | --- | --- |
| `raw` | `application/octet-stream` | **yes** (virtual size = `Content-Length`) | yes | the default |
| `vhd` | `application/vhd` | no | no | needs an existing VDI |
| `tar` | `application/x-tar` | no | no | VDI as a tar stream with inline checksums |
| `qcow2` | `application/x-qemu-disk` | no | no | needs a pre-created VDI on a qcow2-capable SR |

Constraints, straight from `import_raw_vdi.ml`:

- Auto-creation happens only for `raw` with a `sr_uuid`/`sr_id` and a
  `Content-Length`; anything else without a `vdi` is rejected with
  `"Importing a VHD/QCOW2 directly into an SR not yet supported"` or
  `"Not enough info supplied to import"`.
- Chunked encoding is rejected for every non-`raw` format
  (`"Cannot handle chunked VHD"`).
- An unknown `format` string yields `"Unknown format <x>"` and HTTP 404,
  whereas a *known* format in an unsupported context yields HTTP 400. That
  difference is a reliable capability probe (see below).

A capability probe against a live XCP-ng 8.3.0 host (`xapi 26.1.16`), sending a
VDI-less request per format string:

```
format=raw    -> HTTP 200 (auto-create path; empty body fails later)
format=vhd    -> HTTP 400 (known)
format=tar    -> HTTP 400 (known)
format=qcow2  -> HTTP 400 (known)
format=qcow   -> HTTP 404 (unknown)
format=bogus  -> HTTP 404 (unknown)
```

## Task handling

`import_raw_vdi.ml` sends a `task-id` HTTP header referring to the task created
by `xapi_http.with_context` for the request. On a live XCP-ng 8.3 host that
header value is **not retrievable** via `Task.get_status` - it comes back as
`HANDLE_INVALID`.

To actually observe the outcome, create your own task first and pass its ref as
`task_id`:

```go
task, err := xapi.Task.Create(session, "import my.vhd", "")
// ... PUT .../import_raw_vdi?...&task_id=<task>
status, err := xapi.Task.GetStatus(session, task) // poll until success/failure
```

The import handler completes that forwarded task, so polling it is the reliable
way to distinguish "uploaded and written" from "stream failed".

## Fixed vs. dynamic VHD

`format=vhd` is handled by `vhd-tool` (`--source-format vhd --destination-format
raw`, see
[`ocaml/xapi/vhd_tool_wrapper.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi/vhd_tool_wrapper.ml)),
which is built on the `vhd-format` library. That library reads only **dynamic**
VHDs and rejects fixed ones outright:

```ocaml
(* vhd-format: get_sector_location / read_sector *)
| Disk_type.Fixed_hard_disk,_ -> fail (Failure "Fixed disks are not supported")
```

([`vhd_format/f.ml`](https://github.com/xapi-project/ocaml-vhd/blob/master/vhd_format/f.ml)).

MikroTik's CHR images (`chr-*.vhd`) are **fixed** VHDs: the raw disk starting at
offset 0, with only a 512-byte `conectix` footer appended at the end. Uploading
one as `format=vhd` therefore fails server-side with a task in `failure` and no
useful message (the exception text only reaches the host's logs). The same
happens for `xe vdi-import format=vhd` and Xen Orchestra's disk import - they
all go through the same `vhd-tool`.

Two ways out:

1. **Strip the footer and import as `raw`** (no external tools needed). A fixed
   VHD's payload is the first `current_size` bytes of the file, so stream
   `file[0:current_size]` with `format=raw`. Detect it from the footer:

   | Footer field | Offset (from end-512) | Value |
   | --- | --- | --- |
   | cookie | 0 | `conectix` |
   | current size | 48 (uint64 BE) | virtual disk size in bytes |
   | disk type | 60 (uint32 BE) | `2` fixed, `3` dynamic, `4` differencing |

2. **Convert it first**, then use the format you like:

   ```sh
   qemu-img convert -O vpc -o subformat=dynamic chr.vhd chr-dynamic.vhd  # -> format=vhd
   qemu-img convert -O raw chr.vhd chr.raw                               # -> format=raw
   ```

A dynamic VHD is the intended input for `format=vhd` (footer copy at offset 0,
a `cxsparse` dynamic header at 512, and a block-allocation table).

## QCOW2: image format vs. wire format

Two different things share the name:

- **On-SR image format** - how a VDI is stored on the SR. Selected by the SR's
  `preferred-image-formats` (a PBD `device-config` value, default `vhd,qcow2`)
  or per-VDI via `sm-config:image-format`. Read it back with the VDI's
  `sm-config`. Supported on `ext`, `nfs`, `lvm`, `lvmohba`, `lvmofcoe`,
  `lvmoiscsi`, ... but **not** on LINSTOR/XOSTOR or SMB. This layer also raises
  the per-disk limit beyond VHD's 2 TiB (currently 16,381 GiB).
- **Wire format** - the `format=qcow2` query parameter on
  `/import_raw_vdi`/`/export_raw_vdi`, i.e. the encoding of the bytes on the
  socket.

They are related but independently implemented. The XCP-ng storage-side feature
is documented in the
[QCOW2 FAQ](https://docs.xcp-ng.org/storage/qcow2_faq/); note that page does not
mention `vdi-import`/`vdi-export` at all - it is about the on-SR image format.

On the wire side, `Importexport.Format` maps `"qcow2"` to `Qcow`, and the
handlers call `Qcow_tool_wrapper.receive`/`send`
([`ocaml/xapi/qcow_tool_wrapper.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi/qcow_tool_wrapper.ml)).
This is present on the XCP-ng `8.3` branch and confirmed on a live host as
described above. To import a `.qcow2` file:

1. make sure the target SR can hold qcow2 VDIs,
2. create the VDI first (ideally `sm-config:image-format=qcow2`, size ≥ the
   qcow2 virtual size),
3. `PUT /import_raw_vdi?...&vdi=<ref>&format=qcow2` - **not** chunked,
4. poll your task for the result.

Exporting qcow2 from a VHD-backed VDI just falls back to raw (see the note in
`qcow_tool_wrapper.ml`), so a real qcow2 export needs a qcow2 VDI.

## Verifying an import

`/export_raw_vdi` returns the VDI's raw content, which makes a cheap end-to-end
proof: export the freshly imported VDI and compare a hash against the source
file's virtual content (skipping a VHD footer if there is one). This catches
truncated or mis-routed uploads that still returned HTTP 200.

## Minimal example

The XML-RPC half comes from this package; the upload is a plain HTTP `PUT`:

```go
transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
xapi, _ := xenapi.NewClient("https://10.0.0.10/", transport)
session, _ := xapi.Session.LoginWithPassword(user, pass, "1.0", "myimport")
defer xapi.Session.Logout(session)

vdi, _ := xapi.VDI.Create(session, xenapi.VDIRecord{
    NameLabel:   "imported",
    SR:          srRef,
    VirtualSize: 128 << 20,
    Type:        xenapi.VdiTypeUser,
})
task, _ := xapi.Task.Create(session, "import", "")

f, _ := os.Open("disk.raw")
defer f.Close()
fi, _ := f.Stat()

url := fmt.Sprintf("https://10.0.0.10/import_raw_vdi?session_id=%s&vdi=%s&format=raw&task_id=%s",
    session, url.QueryEscape(string(vdi)), url.QueryEscape(string(task)))
req, _ := http.NewRequest(http.MethodPut, url, f)
req.ContentLength = fi.Size() // required: raw auto-create also uses it
req.Header.Set("Content-Type", "application/octet-stream")
resp, _ := (&http.Client{Transport: transport}).Do(req) // no timeout for large images
resp.Body.Close()

for {
    st, _ := xapi.Task.GetStatus(session, task)
    if st == xenapi.TaskStatusTypeSuccess { break }
    if st == xenapi.TaskStatusTypeFailure { panic("import failed") }
    time.Sleep(2 * time.Second)
}
```

## Whole-VM archives: OVA and XVA

The endpoints above move a single disk. A whole VM travels as an *archive*, and
the only archive format XAPI itself understands is **XVA** - it is not a
neutral interchange format but a dump of XAPI's own object graph. **OVA** (and
bare VMDK/disk images) is handled entirely *client-side*: XenCenter and Xen
Orchestra unpack the archive, parse the OVF, convert the VMDK, and then drive
the same HTTP endpoints the CLI uses. The two formats share the "tar" framing
and little else.

### Neither OVA nor VMDK is a VDI wire format

`Importexport.Format` has no member for either, so `/import_raw_vdi` cannot take
an OVA or a VMDK: there is no `format=ova`/`format=vmdk`, and the only accepted
strings are `raw`, `vhd`, `tar`, `qcow2`. The `tar` member is not the OVF tar
container either - it is XAPI's own `Stream_vdi` tar (a VDI streamed as
checksummed chunks, see below), consumed by `Stream_vdi.recv_all_vdi`, and an
OVA tar does not match its layout. The XAPI tree contains no `ovf`/`ova`/`vmdk`
handling at all: OVA import is client-side work in XenCenter/Xen Orchestra,
which unpack the archive, parse the OVF, and feed the disk to the host.

### Why OVA import is client-side

This is a deliberate boundary, not a missing feature:

- **XVA is a XAPI serialisation, not an interchange format.** Its `ova.xml` is
  the XAPI database rows for one VM; import is largely "replay my own records".
  The OVF/CIM model has no XAPI refs, so importing it means inventing a lossy
  mapping (which SR, which network, MACs, `platform`, SM config, VBD device
  numbers, ...).
- **Disk readers.** XAPI's import pipeline wires up `vhd-tool` and `qcow-tool`
  only; there is no VMDK reader in the server, and an OVA's disk is usually a
  possibly-compressed `streamOptimized` VMDK. VMware imports exist, but they go
  through `VM.import_convert` (`xe vm-import type=<vendor>`), which drives an
  external conversion service - still no server-side OVA parsing.
- **Where conversion belongs.** XenServer/XCP-ng keep XAPI thin and put
  foreign-format handling in clients/helpers: XenCenter (.NET) and Xen
  Orchestra (Node) both parse the OVF and convert the VMDK themselves, then
  reuse the same `/import_raw_vdi` and `/import` endpoints. The "XVA, OVA, ..."
  menu in XenCenter reflects the *client's* capabilities, not the server's.
- **Attack surface.** Parsing untrusted OVF/VMDK inside dom0 would enlarge the
  control plane's attack surface and add an open-ended maintenance burden for a
  spec with many optional vendor extensions.

### XVA is a tar too, but its contents are XAPI's

A whole-VM archive moves through a different endpoint, `PUT /import`
([`ocaml/xapi/import.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi/import.ml)),
or through `VM.Import(url, sr, fullRestore, force)`, which has the **host** pull
an XVA from a URL. Export is `GET /export`
([`ocaml/xapi/export.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi/export.ml)).
Inspecting a real XVA exported from a live host shows how little it shares with
an OVA:

| OVA | XVA |
| --- | --- |
| tar container | tar container |
| `*.ovf` (DMTF/CIM hardware description) + `*.mf` | `ova.xml` (XAPI object graph: `VM`, `VDI`, `VBD`, `VIF`, `network`, `SR`, `VM_metrics`, ... in XML-RPC `<value>` encoding) |
| one disk, usually a `streamOptimized` VMDK | raw disk split into **1 MiB chunks** `Ref:<vdi-ref>/<8-digit index>`, each with a 16-hex `.xxhash` sibling |
| no XAPI references | internal `Ref:NNNN` ids, UUIDs, MACs, `platform`, SM config |

So the two share only the tar framing. Converting between them is a real format
translation: map the OVF hardware model onto XAPI records, decode the VMDK
(usually inflating a compressed `streamOptimized` grain table), split the raw
stream into 1 MiB chunks with xxhash checksums, and wrap it all in an
`ova.xml`. There is no upstream OVA-to-XVA converter; Xen Orchestra offers both
because it performs exactly these intermediate steps (or, more simply, imports
the disk and lets the host produce the XVA).

### The pragmatic OVA -> XVA route

Rather than hand-building `ova.xml`, import the OVA's disk into a VM first and
then export *that* VM as XVA - the host generates the correct metadata and
checksums itself:

```
OVA --(untar)--> VMDK --(qemu-img convert -f vmdk -O raw)--> raw
     --(VDI.create + PUT /import_raw_vdi?format=raw)--> VM
     --> GET /export?uuid=<vm-uuid> --> XVA
```

Verified on a live XCP-ng 8.3 host with MikroTik `chr-7.25rc1.ova`:

- `qemu-img` reports the OVF's disk as `streamOptimized`, compressed, virtual
  size 128 MiB; after conversion the raw file is a GPT image (`EFI PART` at
  offset 512).
- The exported XVA (134,436,864 bytes) contained a 20,873-byte `ova.xml`, 128
  data chunks + 128 `.xxhash` files, and `virtual_size` 134,217,728.
- Re-importing that XVA with `PUT /import` (`sr_uuid=...`, `task_id=...`)
  succeeded (HTTP 200, task `success`) and produced a **fresh** VM UUID and
  **fresh** VIF MAC: `restore=false` regenerates identity, so it does not
  collide with the source VM. The disk came back at its full 128 MiB virtual
  size.

### `/import` endpoint facts

- The method is **PUT** (the URI is declared as `put_import`).
- `Content-Length` is mandatory: chunked transfer encoding is explicitly
  rejected (`Cannot_handle_chunked`, HTTP 403). Set `req.ContentLength` - Go
  does not infer it from an `*os.File` body.
- Query parameters: `session_id` (also used to keep the session alive during a
  long upload), `sr_id`/`sr_uuid` (target SR; falls back to the default SR when
  omitted), `task_id` (forward to a task you created, as with VDI import),
  `restore` (full restore: preserve UUIDs and replace an existing VM) and
  `force` (ignore checksum failures).
- Without `restore`, the VM is imported as new: fresh UUIDs and MACs, and the
  `name_label` from `ova.xml` is kept - XAPI accepts no name override - so
  rename afterwards if you need a different one.

## Version compatibility

Generated bindings in this fork are regenerated against the current XenAPI
schema (see `xenapi.SchemaXAPIRelease`). At the time of writing that is XAPI
release `26.17.0`, while the endpoint facts above were verified against a live
**XCP-ng 8.3.0 / xapi 26.1.16** host. The schema can be ahead of a given host:
methods or enum values the host does not know fail with
`MESSAGE_METHOD_UNKNOWN`/`UNIMPLEMENTED`, and enum parsing here tolerates
unknown values rather than aborting (see `convert_gen.go`). None of the
import/export machinery above is version-gated beyond what is stated in the
format table.

## Sources

- XAPI upstream: [`xapi-project/xen-api`](https://github.com/xapi-project/xen-api)
  - `ocaml/xapi/importexport.ml`, `import_raw_vdi.ml`, `export_raw_vdi.ml`
  - `ocaml/xapi/import.ml`, `ocaml/xapi/export.ml` (whole-VM XVA import/export,
    `PUT /import` and `GET /export`)
  - `ocaml/xapi/vhd_tool_wrapper.ml`, `qcow_tool_wrapper.ml`
  - `ocaml/xapi/xapi_http.ml`, `ocaml/xapi-consts/constants.ml`,
    `ocaml/idl/datamodel.ml`
- VHD reader: [`xapi-project/ocaml-vhd`](https://github.com/xapi-project/ocaml-vhd)
  - `vhd_format/f.ml` ("Fixed disks are not supported")
- XCP-ng fork (the branch XCP-ng 8.3 runs):
  [`xcp-ng/xen-api`, branch `8.3`](https://github.com/xcp-ng/xen-api/tree/8.3)
- XCP-ng documentation:
  [QCOW2 FAQ](https://docs.xcp-ng.org/storage/qcow2_faq/)
