# Go XenAPI client library (fork)

[![Regenerate bindings](https://github.com/schankst/go-xen-api-client/actions/workflows/regenerate.yml/badge.svg)](https://github.com/schankst/go-xen-api-client/actions/workflows/regenerate.yml)

A Go client covering the entire [XenAPI](https://xapi-project.github.io/xen-api/)
(XCP-ng / XenServer / Citrix Hypervisor) — one Go type per class/enum, one Go
method per RPC message, generated straight from XenAPI's own schema.

This is a fork of [terra-farm/go-xen-api-client](https://github.com/terra-farm/go-xen-api-client),
originally forked from [v0.0.2](https://github.com/terra-farm/go-xen-api-client/releases/tag/v0.0.2)
(a schema snapshot only current through XenServer 7.3 "inverness", ~2017).

Module path is `github.com/schankst/go-xen-api-client` so it can be pulled
in directly via `go get`/`require` without a `replace` directive.

## Usage example

The following example demonstrates how to instruct XenServer to start a VM with
a given name label. `XEN_API_HOST` is just the host - a bare IP/hostname, not a
URL - and `XEN_API_INSECURE` controls TLS certificate verification, off by
default here since XCP-ng/XenServer hosts commonly run with a self-signed
certificate:

```go
package main

import (
    "crypto/tls"
    "fmt"
    "net/http"

    "github.com/schankst/go-xen-api-client"
)

const XEN_API_HOST string = "10.0.0.10"
const XEN_API_INSECURE bool = true // false to verify the server's TLS certificate
const XEN_API_USERNAME string = "USERNAME"
const XEN_API_PASSWORD string = "PASSWORD"
const VM_NAME_LABEL = "VM NAME LABEL"

func main() {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: XEN_API_INSECURE},
    }

    xapi, err := xenapi.NewClient(fmt.Sprintf("https://%s/", XEN_API_HOST), transport)
    if err != nil {
        panic(err)
    }

    session, err := xapi.Session.LoginWithPassword(XEN_API_USERNAME, XEN_API_PASSWORD, "1.0", "example")
    if err != nil {
        panic(err)
    }

    vms, err := xapi.VM.GetByNameLabel(session, VM_NAME_LABEL)
    if err != nil {
        panic(err)
    }

    if len(vms) == 0 {
        panic(fmt.Errorf("No VM template with name label %q has been found", VM_NAME_LABEL))
    }

    if len(vms) > 1 {
        panic(fmt.Errorf("More than one VM with name label %q has been found", VM_NAME_LABEL))
    }

    vm := vms[0]

    err = xapi.VM.Start(session, vm, false, false)
    if err != nil {
        panic(err)
    }

    err = xapi.Session.Logout(session)
    if err != nil {
        panic(err)
    }
}
```

## Status

- **Actively maintained.** Dependency security alerts, `go vet` findings,
  documentation gaps, and licensing issues in what it depends on get
  triaged and fixed as they come up.
- **Current.** Regenerated weekly against the live XenAPI schema (see
  "Versioning" below), not a one-time snapshot.
- **Tested.** `xmlrpc` has both unit and `httptest.Server`-based
  integration tests; a handful of representative generated converters are
  covered too — not exhaustive across all 74 generated files, but covers
  what has actually broken so far.
- **Documented.** `go doc` is useful throughout: generated code carries
  XenAPI's own descriptions, and the hand-written scaffolding around it
  has its own doc comments.
- **Zero external Go module dependencies.** Including `xmlrpc` itself,
  which used to be a vendored copy of an unlicensed library — see
  [LICENSING.md](LICENSING.md).

This entire fork - the schema regeneration, the enum-tolerance behavior,
the `xmlrpc` rewrite, the documentation and test suite, and the CI
automation - was built end to end by [Claude Code](https://claude.com/claude-code)
(Anthropic's coding agent) working with the maintainer in a single ongoing
conversation, starting from a client that panicked against a modern
XCP-ng host because upstream's schema predated it by years.

See [Releases](https://github.com/schankst/go-xen-api-client/releases) for
what changed release to release.

## Versioning

Git tags on this repo (`v0.1.0`, ...) are this module's own SemVer, tracked
independently of XenAPI's own versioning — deliberately, since tagging this
module to match an XAPI version number directly (e.g. `v26.16.1`) would make
Go's module tooling treat it as a `v26` major version and require a `/v26`
module path suffix, which has nothing to do with XAPI at all.

To keep it traceable which live XAPI version a given tag was verified
against, that's tracked separately:

- `xenapi.SchemaXAPIRelease` (in `client.go`) — the newest XAPI release
  name/version present in the `xenapi.json` schema the bindings were
  generated from, queryable at runtime.
- Each tag's annotation message states the same.

## What's different from upstream

- Bindings regenerated against the current XenAPI schema (see
  `xenapi.SchemaXAPIRelease` for exactly which release), instead of the
  original v0.0.2 snapshot, which only covered XenServer 7.3-era XenAPI.
- Enum parsing tolerates unknown values instead of hard-erroring, as a hedge
  against anything newer than that schema.
- `error.go` is kept in sync with the current upstream error definitions
  instead of a one-time, now-stale copy (see [MAINTAINING.md](MAINTAINING.md)).
- Automated weekly regeneration (see [MAINTAINING.md](MAINTAINING.md)),
  instead of a manual, one-off process.
- Module path changed to `github.com/schankst/go-xen-api-client` so it can
  be pulled in directly.
- `xmlrpc` is this fork's own implementation instead of a dependency on
  `amfranz/go-xmlrpc-client` — that library had no license (see
  [LICENSING.md](LICENSING.md)), and the rewrite along the way fixed its
  remaining bugs and made request encoding measurably faster.

## More

- [GOOD_TO_KNOW.md](GOOD_TO_KNOW.md) — real XenAPI behavior/gotchas this
  fork's development turned up, with sources.
- [PERFORMANCE.md](PERFORMANCE.md) — fetching at scale: why per-object RPC
  calls don't scale to real environments, and the batch-fetch pattern to
  use instead.
- [UPDATES_AND_PATCHES.md](UPDATES_AND_PATCHES.md) — finding pending patches:
  unapplied `pool_update` objects, repository-based updates, and outstanding
  post-update guidances.
- [MAINTAINING.md](MAINTAINING.md) — how the generated bindings and
  `error.go` are produced, and how to refresh them by hand.
- [LICENSING.md](LICENSING.md) — the licensing situation of everything
  this fork is built from or against.
- [CONTRIBUTING.md](CONTRIBUTING.md) — issue/PR process.
- [SECURITY.md](SECURITY.md) — reporting a vulnerability.
