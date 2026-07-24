# Good to know

Real XenAPI behavior this fork's development turned up - each one
verified against a live XCP-ng host and/or the XAPI source, not assumed.
Meant to save you from re-discovering the same surprises.

## Two unrelated "version" numbers

`Host.APIVersionMajor`/`APIVersionMinor` (e.g. `2.21`) and the XAPI
release your host is actually running (e.g. `26.16.0`) are **completely
independent numbers from different code paths** - don't compare them or
expect them to move together.

- `Host.APIVersionMajor`/`Minor` comes from `Datamodel_common.api_version_major`/
  `minor`, which is just `Api_version.api_version_major`/`minor` - a
  **compile-time flag** (`--xapi_api_version_major`/`--xapi_api_version_minor`),
  defaulting to `2`/`23` upstream. Each distributor sets its own value at
  build time; XCP-ng hardcodes `2`/`21` in its own packaging.
  - Source: [`configure.ml`](https://github.com/xapi-project/xen-api/blob/master/configure.ml)
    (upstream default), [`ocaml/idl/datamodel_common.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/idl/datamodel_common.ml)
    (aliases `Api_version`), [`xcp-ng-rpms/xapi`'s `SPECS/xapi.spec`](https://github.com/xcp-ng-rpms/xapi/blob/master/SPECS/xapi.spec)
    lines ~23-24 and ~565 (XCP-ng's own override).
- The `26.16.0`-style number is the actual `xapi` package/build version,
  parsed from the build's own version string (e.g. a git tag) via OCaml's
  `Build_info.V1.version()` - a completely separate mechanism with no
  code relationship to the field above.
  - Source: [`ocaml/util/xapi_version.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/util/xapi_version.ml).
- **This is also what `xenapi.SchemaXAPIRelease` in this fork tracks** -
  it's the release-name/version tag from `xenapi.json`'s own lifecycle
  data (see `MAINTAINING.md`), i.e. the same numbering family as
  `26.16.0`/`26.16.1-next`, not the `2.x` field.

## The `2.x` API version gates pool membership, not general compatibility

A host can only join a pool if its `API_version_major`/`minor` **matches
the pool master's exactly** - not "close enough", not "newer is fine".
A mismatch is rejected outright with `ERR_POOL_JOINING_HOST_MUST_HAVE_SAME_API_VERSION`
(one of the constants in `error.go`).

- Source: `assert_api_version_matches` in
  [`ocaml/xapi/xapi_pool.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi/xapi_pool.ml).

Rolling Pool Upgrade (upgrading hosts in a pool one at a time without
downtime) is the mechanism that navigates around this: the master is
upgraded first, and "a new xapi version can always talk to an old xapi,
but an old xapi may not be able to talk to a new one" during the
transition, until every host is back on a matching version.

- Source: [`features/RPU/RPU.md`](https://github.com/xapi-project/xapi-project.github.io/blob/master/features/RPU/RPU.md).

## Null/absent object references are a literal string, not an empty one

XenAPI represents "no object" as the literal string `"OpaqueRef:NULL"`,
not `""`. Check for both (an empty string can still show up in some
fields) - see `isNullRef`-style helpers in consumers of this library
(e.g. the `xen` CLI) rather than relying on Go's normal zero-value check.

## `VM.get_all_records` returns more than "VMs"

It also returns every OS template, the dom0 control domain, and (per the
schema) snapshots - all represented as `VM` objects. Filter on
`IsATemplate`, `IsControlDomain`, and `IsASnapshot` if you only want
actual guest VMs; otherwise expect several times more "VMs" than you
actually have.

## Some numeric fields use out-of-band sentinel values

Certain SR backends (observed: ISO SRs) report `-1` for
`physical_utilisation`/`physical_size` to mean "not applicable" rather
than a real byte count. A naive byte→GiB conversion renders this as a
nonsensical `-0.0`; check for negative values and show something like
`n/a` instead.

## Object references can go stale/dangling

Destroying an object does not clear other objects' fields that still
reference it - XAPI doesn't cascade-clear on delete. Concretely observed:
a pool's `default_sr` can point at an SR that was since destroyed. Trying
to resolve it raises `HANDLE_INVALID`, not a null/empty result - treat
that as "stale reference, no live default" rather than a fatal error.

## `Event.from`'s payload and `Event.snapshot` are polymorphic

The event-batching API returns a struct whose shape isn't a normal XenAPI
record, and the per-event `snapshot` field's concrete type depends on
that event's class at runtime. This fork represents both as opaque
`xmlrpc.Struct` rather than a typed Go struct - see `xmlrpc/doc.go` and
`MAINTAINING.md` for why, if you need to consume events yourself.

## `Event`'s deprecated `timestamp` field doesn't arrive as a real `datetime`

Every other `datetime`-typed field in the schema turns up on the wire as a
properly tagged `<dateTime.iso8601>` value, which the `xmlrpc` package
parses straight into a Go `time.Time`. `Event`'s `timestamp` field - the
one on `EventRecord` from `event.next`, not `event.from`'s opaque batch -
is the one exception observed so far: a live XCP-ng host sent it as a
plain `<string>` containing an OCaml-style float, e.g. `"1784931839."`
(the trailing dot is `string_of_float`'s own rendering of a whole-number
float, not a typo).

`convertTimeToGo` tolerates this - falling back to `strconv.ParseFloat`
and `time.Unix` when the value isn't already a `time.Time` - instead of
failing the whole event. Worth knowing if you write your own XML-RPC
client against this field: don't assume every `datetime`-typed field is
wire-compatible with the others just because the schema says so alike.
The field is marked `Deprecated_s` in `xenapi.json`, which likely explains
why it isn't held to the same wire-format discipline as everything else.

## A host's boot time isn't a schema field - it's an `other_config` convention

There's no `Host.boot_time`/`uptime` field in the schema at all. XCP-ng
(and XenServer generally) instead stashes it as a plain string value in
`Host.other_config["boot_time"]` - an informal convention, not part of
the typed datamodel, so it's just whatever string happens to be in that
generic `map[string]string`. `other_config["agent_start_time"]` is the
same idea for when the `xapi` toolstack itself last (re)started, which
can be more recent than the actual host boot if only the toolstack was
restarted.

Observed on a live XCP-ng host, both values are the same kind of
OCaml-float-as-string this fork already works around for
`Event.timestamp` above - e.g. `"1782639432."`, trailing dot included -
not a plain integer. Parse with `strconv.ParseFloat` + `time.Unix`, not
`strconv.Atoi`/`ParseInt`, the same way. See `parseOtherConfigTime` in
the [`xen`](https://github.com/schankst/xen) CLI's `hosts.go` for a
consumer doing exactly that to show boot time/uptime.
