# XenAPI notes

XenAPI behaviour that is easy to get wrong. Verified against a live XCP-ng
host and the [XAPI source](https://github.com/xapi-project/xen-api).

## Two unrelated version numbers

- `Host.APIVersionMajor`/`APIVersionMinor` (e.g. `2.21`) is a build-time
  constant set by each distributor. The upstream default is `2.23`; XCP-ng
  sets `2.21`.
- The XAPI release (e.g. the `xapi` package version) is a separate number
  derived from the build's version string.
- `xenapi.SchemaXAPIRelease` uses the XAPI release numbering.

The two are independent; do not compare them.

## The API version gates pool membership

A host can join a pool only if its API version matches the master's
exactly. A mismatch fails with `xenapi.ErrPoolJoiningHostMustHaveSameAPIVersion`.
Rolling Pool Upgrade handles this by upgrading the master first; a newer
xapi can talk to an older one, but not the reverse.

## Odd values

- "No object" is the string `"OpaqueRef:NULL"`. Some fields can also be
  `""`. Treat both as null.
- `VM.GetAllRecords` also returns templates, dom0 and snapshots. Filter on
  `IsATemplate`, `IsControlDomain` and `IsASnapshot` to get real guests.
- ISO SRs report `-1` for `PhysicalUtilisation` and `PhysicalSize`, meaning
  "not applicable".

## Stale references

XAPI does not clear references when an object is destroyed. For example,
a pool's `DefaultSR` can point to a deleted SR. Resolving it fails with
`HANDLE_INVALID`; treat that as "no live object":

```go
_, err := client.SR.GetRecord(ctx, session, pool.DefaultSR)
if errors.Is(err, xenapi.ErrHandleInvalid) { /* stale: no default SR */ }
```

## Events are untyped

The `Event.From` result and each event's `snapshot` field are polymorphic.
The client returns them as `xmlrpc.Struct`; parse them yourself.

## Float timestamps as strings

`EventRecord.Timestamp` (deprecated) is typed `datetime`, but XCP-ng sends
it as a string holding an OCaml float, e.g. `"1784931839."`. The client
parses this form. It is the only `datetime` field seen doing this.

## Host boot time

There is no boot-time field. XCP-ng stores it in
`Host.OtherConfig["boot_time"]`; `OtherConfig["agent_start_time"]` holds the
last toolstack start. Both are OCaml float strings (`"1782639432."`): parse
with `strconv.ParseFloat` and `time.Unix`, not `strconv.Atoi`.

## Performance: fetch each class once

Every call is an HTTP round trip to the pool master. Calling `GetRecord` per
object (VM -> VBD -> VDI -> SR) is slow on a real pool. Fetch each class
once with `GetAllRecords` and join in memory:

```go
vms, _ := client.VM.GetAllRecords(ctx, session)
vbds, _ := client.VBD.GetAllRecords(ctx, session)
vdis, _ := client.VDI.GetAllRecords(ctx, session)

for _, vm := range vms {
    for _, vbdRef := range vm.VBDs {
        vbd, ok := vbds[vbdRef]
        if !ok {
            continue // deleted since the VM was fetched
        }
        vdi := vdis[vbd.VDI] // map lookup, no RPC
        fmt.Println(vm.NameLabel, vdi.NameLabel)
    }
}
```

- Check `ok`: a missing ref usually means the object was deleted.
- `VMRecord.Snapshots` refers to VMs already in the `VM.GetAllRecords` map.
- `GetAllRecords` returns full records; fetch only the classes you join on.
- For long-running processes, `Event.From` avoids re-polling, at the cost of
  parsing untyped results (see above).
