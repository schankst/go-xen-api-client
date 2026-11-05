# Pending updates and patches

XenAPI has no single "pending patches" call. Use one of three approaches:

| Question | Approach |
| --- | --- |
| Uploaded updates not yet applied? | `PoolUpdate` records and their `Hosts` |
| Updates available in the repository? | `Pool.SyncUpdates` and the `/updates` HTTP handler |
| Actions still required after updating? | `Host.GetPendingGuidances` |

## Uploaded updates (`PoolUpdate`)

`PoolUpdateRecord.Hosts` lists the hosts the update is applied to; it is
pending on all others (error handling omitted):

```go
updates, err := client.PoolUpdate.GetAllRecords(ctx, session)
hosts, err := client.Host.GetAll(ctx, session)
for _, u := range updates {
    for _, h := range hosts {
        if !slices.Contains(u.Hosts, h) {
            // u.NameLabel, u.Version pending on host h
        }
    }
}
```

- `PoolUpdate.Precheck` returns a `LivepatchStatus` (can it apply live?).
- `AfterApplyGuidance`: `restartHost`, `restartXAPI`, `restartHVM`, `restartPV`.
- `PoolUpdate.Apply` applies to one host, `PoolUpdate.PoolApply` to all. On
  XenServer 6.x, use the deprecated `Host.GetPatches` and `host_patch.applied`
  instead.

## Repository updates (XenServer 8, XCP-ng)

No RPC lists the available packages; XML-RPC only syncs and applies:

```go
hash, err := client.Pool.SyncUpdates(ctx, session, pool, false, "", "", "", "")
ready, err := client.Pool.CheckUpdateReadiness(ctx, session, pool, true)
applied, err := client.Host.ApplyUpdates(ctx, session, host, hash)
```

The list comes from `GET https://<master>/updates?session_id=<ref>` (JSON,
not wrapped by this package).

- State: `Repository.GetHash`, `Pool.GetLastUpdateSync`,
  `Pool.GetRepositories`.
- Scheduled sync: `Pool.SetUpdateSyncEnabled`, `Pool.ConfigureUpdateSync`
  (`daily`/`weekly`, day) and the `GetUpdateSync*` getters.

## Outstanding actions per host

- `Host.GetLatestSyncedUpdatesApplied` is the quickest check: `yes`, `no`
  (synced updates are missing on this host) or `unknown`.
- Details: `Host.GetPendingGuidances`, `GetPendingGuidancesRecommended`,
  `GetPendingGuidancesFull`, `GetUpdatesRequiringReboot`,
  `GetLastSoftwareUpdate`.
- `UpdateGuidances`: `reboot_host`, `reboot_host_on_[kernel_|xen_]livepatch_failure`,
  `restart_toolstack`, `restart_device_model`, `restart_vm`.
