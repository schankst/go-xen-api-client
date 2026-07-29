# Querying pending patches and updates

The XenAPI has **no** single call along the lines of "show me all pending
patches". Depending on what you actually mean, there are three different
approaches — all of them available in this client.

| Question | Approach |
| --- | --- |
| Which uploaded updates have not been applied yet? | `pool_update` plus its `hosts` field (section 1) |
| Which updates are available in the repository? | `pool.sync_updates` plus the `/updates` HTTP handler (section 2) |
| Which actions are still outstanding after patching? | `host.pending_guidances` (section 3) |

---

## 1. Uploaded but not yet applied updates (`pool_update`)

The classic case. The `pool_update.hosts` field lists the hosts the update has
**already** been applied to, so "pending" is the difference against all hosts in
the pool.

```go
updates, err := client.PoolUpdate.GetAllRecords(session) // pool_update_gen.go:80
if err != nil {
    return err
}
hosts, err := client.Host.GetAll(session)
if err != nil {
    return err
}

for _, u := range updates {
    applied := make(map[xenapi.HostRef]bool, len(u.Hosts))
    for _, h := range u.Hosts {
        applied[h] = true
    }
    for _, h := range hosts {
        if !applied[h] {
            // u.NameLabel, u.Version, u.AfterApplyGuidance are pending for host h
        }
    }
}
```

Useful alongside this:

- `PoolUpdate.Precheck(session, self, host)` — `pool_update_gen.go:174`, tells
  you upfront whether the update can be applied live (`LivepatchStatus`).
- `u.AfterApplyGuidance` — what is required after applying (`reboot_host`,
  `restart_toolstack`, …).
- `PoolUpdate.Apply` / `PoolUpdate.PoolApply` — apply to a single host or to the
  whole pool.

The 6.x-era equivalent: `Host.GetPatches` (`host_gen.go:3488`) →
`host_patch.applied`.

---

## 2. Updates available from the repository (XS 8 / XCP-ng, yum-based)

Here XAPI provides **no** RPC listing of the available packages. The XML-RPC
surface only covers syncing and applying:

```go
hash, err := client.Pool.SyncUpdates(session, pool, false, "", "", "", "") // pool_gen.go:787
ready, err := client.Pool.CheckUpdateReadiness(session, pool, true)        // pool_gen.go:764
// apply:
applied, err := client.Host.ApplyUpdates(session, host, hash)              // host_gen.go:626
```

The actual **list** of available updates is fetched by XenCenter through an HTTP
handler on the pool master:

```
https://<pool-master>/updates?session_id=<session-ref>
```

That handler returns JSON and is **not** wrapped by this client — you need a
plain `http.Get` with the session ref.

Checking repository state:

- `Repository.GetUpToDate` — `repository_gen.go:319`
- `Repository.GetHash` — `repository_gen.go:338`
- `Pool.GetLastUpdateSync` — `pool_gen.go:2683`
- `Pool.GetRepositories` — `pool_gen.go:2987`

---

## 3. Outstanding actions after installed updates

If "pending" means "is there anything still open on this host":

```go
client.Host.GetPendingGuidances(session, host)            // host_gen.go:2766
client.Host.GetPendingGuidancesRecommended(session, host) // host_gen.go:2652
client.Host.GetPendingGuidancesFull(session, host)        // host_gen.go:2633
client.Host.GetUpdatesRequiringReboot(session, host)      // host_gen.go:2899
client.Host.GetLatestSyncedUpdatesApplied(session, host)  // host_gen.go:2690
client.Host.GetLastSoftwareUpdate(session, host)          // host_gen.go:2728
```

`GetLatestSyncedUpdatesApplied` is the quickest per-host check: the
`LatestSyncedUpdatesAppliedState` enum with `yes` / `no` / `unknown`. A `no`
means there are synced updates this host does not have yet.

To apply the recommended guidances: `Host.ApplyRecommendedGuidances`
(`host_gen.go:577`).

### `UpdateGuidances` values

Defined in `vm_gen.go:40` onwards:

| Value | Meaning |
| --- | --- |
| `reboot_host` | Host reboot required |
| `reboot_host_on_livepatch_failure` | Reboot if the livepatch fails |
| `reboot_host_on_kernel_livepatch_failure` | Reboot if the kernel livepatch fails |
| `reboot_host_on_xen_livepatch_failure` | Reboot if the Xen livepatch fails |
| `restart_toolstack` | Restart the toolstack |
| `restart_device_model` | Restart the device model |
| `restart_vm` | Restart the VM |

---

## Automatic update sync

Around the scheduled repository sync:

- `Pool.SetUpdateSyncEnabled` — `pool_gen.go:555`
- `Pool.ConfigureUpdateSync(session, self, frequency, day)` — `pool_gen.go:574`
- `Pool.GetUpdateSyncEnabled` / `GetUpdateSyncFrequency` / `GetUpdateSyncDay`
