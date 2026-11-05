# Importing and exporting disk data

No XML-RPC method carries disk bytes. Use XML-RPC (this package) to create
objects and tasks, and plain HTTP requests for the data. The HTTP endpoints
are not wrapped by this package; authenticate them with the session
reference from `Session.LoginWithPassword` as `session_id`. Uploads stream
from your process; the file need not be on the host. `VM.Import` (XVA) and
`VM.ImportConvert` make the host pull from a URL instead. Endpoint facts
were verified on XCP-ng 8.3.0 (xapi 26.1.16).

## VDI endpoints

`PUT /import_raw_vdi` and `GET /export_raw_vdi`. Query parameters:

| Parameter | Meaning |
| --- | --- |
| `session_id` | Session ref. Basic auth and client certificates also work. |
| `task_id` | Optional. Attach the work to an existing task. |
| `vdi` | Target or source VDI, as ref or UUID. |
| `sr_id` / `sr_uuid` | Target SR, used when `vdi` is absent. |
| `format` | `raw` (default), `vhd`, `tar` or `qcow2`. |
| `chunked` | Presence only. Chunked transfer encoding; `raw` only. |

## Wire formats

| `format` | Content-Type | Auto-create VDI | Chunked |
| --- | --- | --- | --- |
| `raw` | `application/octet-stream` | yes, size = `Content-Length` | yes |
| `vhd` | `application/vhd` | no | no |
| `tar` | `application/x-tar` | no | no |
| `qcow2` | `application/x-qemu-disk` | no | no |

- Only `raw` with `sr_id`/`sr_uuid` and `Content-Length` creates the VDI.
- `tar` is XAPI's checksummed VDI stream, not an OVA. There is no `ova`,
  `vmdk` or `qcow` value.
- Unknown `format`: HTTP 404. Known format in an unsupported context: HTTP
  400. This can be used to probe support.

## Task handling

The handler returns a `task-id` header, but that task cannot be queried
(`HANDLE_INVALID`). Create your own task with `Task.Create`, pass it as
`task_id`, and poll `Task.GetStatus` until `success` or `failure`
(details in `Task.GetErrorInfo`). HTTP 200 alone does not mean the data
was written. To verify content, hash the `GET /export_raw_vdi` output and
compare with the source's virtual content (without any VHD footer).

## Fixed vs. dynamic VHD

`format=vhd` accepts only dynamic VHDs. A fixed VHD fails with a task in
`failure` and no useful message (details only in host logs); this also
affects `xe vdi-import` and Xen Orchestra. MikroTik CHR images are fixed.

A fixed VHD is the raw disk plus a 512-byte footer. In the footer, `conectix`
is at offset 0, the virtual size (uint64 BE) at 48, and the disk type
(uint32 BE: `2` fixed, `3` dynamic, `4` differencing) at 60. Either stream
the first virtual-size bytes with `format=raw`, or convert first:

```sh
qemu-img convert -O vpc -o subformat=dynamic in.vhd out.vhd  # format=vhd
qemu-img convert -O raw in.vhd out.raw                       # format=raw
```

## QCOW2

The on-SR image format is separate from the `format=qcow2` wire format.
It is set by the PBD `device-config:preferred-image-formats` (default
`vhd,qcow2`) or per VDI with `sm-config:image-format`, lifts VHD's 2 TiB
limit, and is not supported on LINSTOR/XOSTOR or SMB. See the
[XCP-ng QCOW2 FAQ](https://docs.xcp-ng.org/storage/qcow2_faq/).

To import a `.qcow2` file, create the VDI on a qcow2-capable SR (ideally
with `sm-config:image-format=qcow2`, size at least the image's virtual
size), then `PUT` with `vdi=<ref>&format=qcow2`, not chunked. Exporting
`qcow2` from a VHD-backed VDI returns raw data.

## Example

Error handling is shortened. `hc` is an `*http.Client` without an overall
timeout; pass it to `xenapi.WithHTTPClient` too, so both share TLS settings.

```go
client, err := xenapi.NewClient("https://pool-master", xenapi.WithHTTPClient(hc))
session, err := client.Session.LoginWithPassword(ctx, user, pass, "1.0", "import")
defer client.Session.Logout(ctx, session)

vdi, err := client.VDI.Create(ctx, session, xenapi.VDIRecord{
    NameLabel: "imported", SR: srRef, VirtualSize: size, Type: xenapi.VdiTypeUser,
})
task, err := client.Task.Create(ctx, session, "import", "")
f, err := os.Open("disk.raw") // exactly size bytes
defer f.Close()
q := url.Values{"session_id": {string(session)}, "vdi": {string(vdi)},
    "format": {"raw"}, "task_id": {string(task)}}
req, err := http.NewRequestWithContext(ctx, http.MethodPut,
    "https://pool-master/import_raw_vdi?"+q.Encode(), f)
req.ContentLength = int64(size) // required; Go does not infer it from *os.File
req.Header.Set("Content-Type", "application/octet-stream")
resp, err := hc.Do(req) // then resp.Body.Close()

for {
    st, err := client.Task.GetStatus(ctx, session, task)
    if err != nil || st == xenapi.TaskStatusTypeFailure { /* failed */ }
    if st == xenapi.TaskStatusTypeSuccess { break }
    time.Sleep(2 * time.Second)
}
```

## Whole VMs: XVA and OVA

XAPI parses only XVA: a tar of `ova.xml` (XAPI's own object graph) plus the
raw disk in 1 MiB chunks, each with an `.xxhash` checksum file.

Export with `GET /export?uuid=<vm>`. Import with `VM.Import` (host pulls
from a URL) or `PUT /import`:

- `Content-Length` is required; chunked uploads are rejected (HTTP 403).
- Parameters: `session_id`, `sr_id`/`sr_uuid` (default SR if omitted),
  `task_id`, `restore` (keep UUIDs, replace the existing VM), `force`
  (ignore checksum failures).
- Without `restore`, the VM gets new UUIDs and MACs and keeps the
  `name_label` from the archive. Rename it afterwards if needed.

OVA and VMDK are handled client-side (XenCenter, Xen Orchestra). To turn an
OVA into an XVA: untar, `qemu-img convert -f vmdk -O raw`, import the raw
disk into a new VM via `/import_raw_vdi`, then `GET /export` that VM.

## Backups

XAPI has no scheduler, retention or catalog. It provides the primitives.

| Goal | API |
| --- | --- |
| Crash-consistent snapshot | `VM.Snapshot` |
| Snapshot with RAM (running VM) | `VM.Checkpoint` |
| Roll back / list / delete | `VM.Revert`, `VM.GetSnapshots`, `VM.Destroy` |
| Disk snapshot / clone | `VDI.Snapshot`, `VDI.Clone`, `VDI.Copy` |
| Changed-block tracking | `VDI.EnableCbt`, `VDI.ListChangedBlocks`, `VDI.GetNbdInfo` |
| Disaster recovery | `VM.AssertCanBeRecovered`, `VM.Recover` |
| Whole VM (XVA) | `GET /export?uuid=<vm>` |
| VM metadata only | `GET /export_metadata?uuid=<vm>` |
| One disk | `GET /export_raw_vdi?vdi=<ref>` |
| Pool database (master) | `GET /pool/xmldbdump` |
| Host configuration | `GET /host_backup` |

A snapshot lives on its VM's SR; it is not a backup until exported. Pair
disk-only exports with `/export_metadata`.

Routine for a full backup:

1. `Task.Create`, then `VM.Snapshot` (quiesce databases in the guest
   first). Never export a running VM's live disk.
2. Stream `GET /export?uuid=<snapshot uuid>&task_id=<task>` to the target.
3. Poll `Task.GetStatus`/`Task.GetErrorInfo` with a timeout.
4. `VM.Destroy` the snapshot, also on failure or timeout.
5. Back up `/pool/xmldbdump` and `/host_backup` too; without them a
   restored VM may have no network or storage.

Caveats:

- One export per VM at a time; avoid concurrent snapshots on one SR.
  Parallel exports of one chain disrupt coalescing.
- Tag snapshots in `other_config` with a run ID; delete leftovers from
  aborted runs at startup.
- Verify: XVA chunks carry `.xxhash` files; disks can be hashed. A
  `success` task alone is not proof.
- Nothing prunes old backups (`VM.Destroy`, `VDI.Destroy`, `VDI.DataDestroy`).
- Test restores with `PUT /import` without `restore` (no UUID/MAC
  collisions). Use `restore=true` only to replace a VM.

Incremental: enable CBT, take two snapshots, get the changed-block bitmap
with `VDI.ListChangedBlocks(from, to)` and read those blocks over NBD
(`VDI.GetNbdInfo`). Chain management and pruning are yours to build.

## Version compatibility

The bindings follow `xenapi.SchemaXAPIRelease`, which can be newer than your
host. Unknown calls fail with `xenapi.ErrMessageMethodUnknown`; unknown
enum values from the host are tolerated. The endpoints are not
version-gated beyond the format table.
