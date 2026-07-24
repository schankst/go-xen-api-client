# Performance: fetching at scale

XenAPI is XML-RPC: every call is an HTTP round trip to the pool master.
That's easy to forget when writing straight-line code that walks an
object graph - and it's exactly the kind of code that quietly turns into
a performance problem once it runs against a real environment with many
VMs, disks, and NICs instead of the two or three you tested with.

## The anti-pattern: one RPC per object

The natural way to write "for each VM, show its disks" is to fetch the
VMs, then loop and fetch each VM's VBDs, then loop those and fetch each
VDI, then its SR, and so on:

```go
vmMap, _ := client.VM.GetAllRecords(session)

for vmRef, vmRec := range vmMap {
    for _, vbdRef := range vmRec.VBDs {
        vbdRec, _ := client.VBD.GetRecord(session, vbdRef)   // 1 RPC
        vdiRec, _ := client.VDI.GetRecord(session, vbdRec.VDI) // 1 RPC
        srRec, _ := client.SR.GetRecord(session, vdiRec.SR)    // 1 RPC
        // ...
    }
}
```

This is correct and reads fine, but it's classic N+1: the number of RPCs
grows with the number of VMs *times* the number of disks/NICs/snapshots
each one has. A pool with a few dozen VMs easily produces several hundred
sequential round trips for something that should be near-instant - and
it's visibly slow, not just "a bit slower than ideal".

## The fix: fetch each class once, join in memory

XenAPI classes support `GetAllRecords`, which returns every object of
that class in a single call. Fetch each related class once up front,
keyed by ref, and then treat the rest of the work as ordinary in-memory
map lookups instead of network calls:

```go
vmMap, _ := client.VM.GetAllRecords(session)
vbdMap, _ := client.VBD.GetAllRecords(session)
vdiMap, _ := client.VDI.GetAllRecords(session)
srMap, _ := client.SR.GetAllRecords(session)

for _, vmRec := range vmMap {
    for _, vbdRef := range vmRec.VBDs {
        vbdRec, ok := vbdMap[vbdRef]   // map lookup, no RPC
        if !ok {
            continue
        }
        vdiRec, ok := vdiMap[vbdRec.VDI]
        if !ok {
            continue
        }
        srRec := srMap[vdiRec.SR]
        // ...
    }
}
```

This turns "hundreds of round trips, growing with environment size" into
a fixed, small number of round trips - one per class involved - regardless
of how many VMs, disks, or NICs the pool actually has.

A few related things worth knowing when you do this:

- **Check `ok`, not just `err`.** A ref that doesn't resolve in the map
  usually means the object was deleted after the VM's record was fetched
  (see the stale-reference note in [GOOD_TO_KNOW.md](GOOD_TO_KNOW.md)),
  not a bug - skip it rather than treating it as fatal.
- **Records can already contain what you'd otherwise re-fetch.** Some
  reverse relationships are already inline on the record you have - e.g.
  `VMRecord.Snapshots` is a `[]VMRef`, and a snapshot is itself a `VM`
  object present in the same `VM.GetAllRecords` map. There's no need for
  a separate `VM.GetSnapshots` call plus a `VM.GetRecord` per snapshot if
  you already fetched the VM map.
- **This is a real example, not a hypothetical.** The [`xen`](https://github.com/schankst/xen)
  CLI that exercises this library did exactly the naive version first,
  fetching VBD/VDI/SR/VIF/Network/VMGuestMetrics records one object at a
  time per VM - and it was visibly slow to watch run. Switching its `vms`
  command to fetch each of those six classes once via `GetAllRecords` and
  joining in memory (see `vms.go` there) took it down to well under two
  seconds against a pool of over 20 VMs with their disks and NICs.

## The trade-off

`GetAllRecords` isn't free - it returns the *entire* record for every
object of that class, including fields you don't need, in one response.
For a handful of classes against a normal-sized pool that's a clear win
over per-object calls. If you're dealing with a very large pool and only
need one or two fields off of a huge class, weigh that against:

- Only fetching the classes you actually need to join against - don't
  reach for `GetAllRecords` on every class in the schema "just in case".
- For long-running processes that need to stay current rather than doing
  one-shot reads, `Event.From`/`Event.Register` exist as a lower round-trip
  alternative to re-polling `GetAllRecords` on a timer - but this fork only
  implements the RPC surface, not a typed consumer: the returned batch and
  each event's `snapshot` are opaque `xmlrpc.Struct` values (see
  `GOOD_TO_KNOW.md`), since `snapshot`'s concrete shape depends on the
  event's class at runtime. Using it means parsing and dispatching those
  structs yourself; there's no ready-made typed helper for it here.

As always: measure against your actual environment size before assuming
either approach is the right call for it.
