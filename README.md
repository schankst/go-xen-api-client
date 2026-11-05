# Go XenAPI client library (fork)

[![Go Reference](https://pkg.go.dev/badge/github.com/schankst/go-xen-api-client.svg)](https://pkg.go.dev/github.com/schankst/go-xen-api-client)
[![Tests](https://github.com/schankst/go-xen-api-client/actions/workflows/test.yml/badge.svg)](https://github.com/schankst/go-xen-api-client/actions/workflows/test.yml)
[![Latest release](https://img.shields.io/github/v/release/schankst/go-xen-api-client)](https://github.com/schankst/go-xen-api-client/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/schankst/go-xen-api-client)](go.mod)
[![License: MIT](https://img.shields.io/github/license/schankst/go-xen-api-client)](LICENSE)
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

Start a VM by name label:

```go
package main

import (
    "fmt"
    "log"

    xenapi "github.com/schankst/go-xen-api-client"
)

func main() {
    xapi, err := xenapi.NewClient("https://10.0.0.10")
    if err != nil {
        log.Fatal(err)
    }

    session, err := xapi.Session.LoginWithPassword("root", "password", "1.0", "example")
    if err != nil {
        log.Fatal(err)
    }
    defer xapi.Session.Logout(session)

    vms, err := xapi.VM.GetByNameLabel(session, "my-vm")
    if err != nil {
        log.Fatal(err)
    }
    if len(vms) != 1 {
        log.Fatal(fmt.Errorf("found %d VMs named my-vm, want 1", len(vms)))
    }

    if err := xapi.VM.Start(session, vms[0], false, false); err != nil {
        log.Fatal(err)
    }
}
```

The host's TLS certificate is verified. For a host with a self-signed
certificate, pass its certificate with `xenapi.WithTLSConfig`
(`&tls.Config{RootCAs: pool}`); see the `NewClient` documentation.

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
module to match an XAPI version number directly (e.g. tagging it `v26.1.2`)
would make Go's module tooling treat it as a `v26` major version and require
a `/v26` module path suffix, which has nothing to do with XAPI at all.

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
- [IMPORT_EXPORT.md](IMPORT_EXPORT.md) — moving VDI data in and out: the
  HTTP import/export endpoints, accepted formats (`raw`/`vhd`/`tar`/`qcow2`),
  push vs. pull, the fixed-VHD gotcha, and task handling.
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
