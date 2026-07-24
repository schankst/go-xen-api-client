# Go XenAPI client library (fork)

[![Regenerate bindings](https://github.com/schankst/go-xen-api-client/actions/workflows/regenerate.yml/badge.svg)](https://github.com/schankst/go-xen-api-client/actions/workflows/regenerate.yml)

This is a fork of [terra-farm/go-xen-api-client](https://github.com/terra-farm/go-xen-api-client),
originally forked from [v0.0.2](https://github.com/terra-farm/go-xen-api-client/releases/tag/v0.0.2)
(a schema snapshot only current through XenServer 7.3 "inverness", ~2017).

Module path is `github.com/schankst/go-xen-api-client` so it can be pulled
in directly via `go get`/`require` without a `replace` directive.

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
  instead of a one-time, now-stale copy (see "Implementation notes" below).
- Automated weekly regeneration (see below), instead of a manual, one-off
  process.
- Module path changed to `github.com/schankst/go-xen-api-client` so it can
  be pulled in directly.
- `xmlrpc` is this fork's own implementation (see `xmlrpc/doc.go`) instead
  of a dependency on `amfranz/go-xmlrpc-client` — that library had no
  license, and the rewrite along the way fixed its remaining bugs and
  made request encoding measurably faster (see the v0.2.0 release notes
  for benchmarks). No external Go module dependencies remain at all.

For what actually changed release to release, see
[Releases](https://github.com/schankst/go-xen-api-client/releases) rather
than this file — this README describes the current state, not the history
of how it got there.

---

# Go XenAPI client library

This is a client library for the Xapi toolstack
(http://xapi-project.github.io/).

This library covers the entire [XenAPI](https://xapi-project.github.io/xen-api/)
and I have successfully used it to implement a Terraform plugin that interfaces
Citrix XenServer. That being said, this library is not production-ready yet.
Use it at your own risk, and don't expect everything in this library to work
out of the box.

## Usage example

The following example demonstrates how to instruct XenServer to start a VM with
a given name label:

```go
package main

import (
    "fmt"
    "github.com/schankst/go-xen-api-client"
)

const XEN_API_URL string = "https://IP.OF.XEN.SERVER"
const XEN_API_USERNAME string = "USERNAME"
const XEN_API_PASSWORD string = "PASSWORD"
const VM_NAME_LABEL = "VM NAME LABEL"

func main() {
    xapi, err := xenapi.NewClient(XEN_API_URL, nil)
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

    xapi.VM.Start(session, vm, false, false)
    if err != nil {
        panic(err)
    }

    err = xapi.Session.Logout(session)
    if err != nil {
        panic(err)
    }
}
```

## Project status

This is upstream's original TODO list for the project to be
production-ready. Status in this fork:

- ~~A strategy how to handle the differences in the XenAPI versions.~~
  **Done** — see `SchemaXAPIRelease`, the enum-tolerance patch, and the
  automated weekly regeneration workflow above.
- Tests, at least for the various data type conversions. **Partially
  done** — `xmlrpc` has both unit-level tests (request/response round
  trip, including the enum-tolerance behavior specifically, and malformed
  input) and `httptest.Server`-based integration tests for `Client.Call`
  (success, XML-RPC fault, pool-affinity cookie capture/replay,
  concurrent calls); a handful of representative `convert_gen.go`
  converters are covered too. Not exhaustive across all 74 generated
  files, but covers what has actually broken so far.
- ~~Embed XenAPI documentation as GoDoc in the generated code.~~ **Already
  true** — inherited from the original generator (class/method/field
  descriptions come straight from the XenAPI schema); topped up the
  hand-written scaffolding around it (`doc.go`, `NewClient`, `Client`,
  `Error`, ...) so `go doc` is useful throughout, not just in the
  generated 95% of the package.
- Better error messages. **Partially addressed** — `Error.Error()` no
  longer prints blank fields for error codes that don't carry an object
  type/UUID.
- ~~Usage examples.~~ One exists below; good enough for this fork's scope.

This fork is **actively maintained**, not a one-time patch: dependency
security alerts, `go vet` findings, documentation gaps, and (as of the
`xmlrpc` rewrite) even licensing issues in what it depends on get triaged
and fixed as they come up, not just left in a TODO list. Concretely, this
entire fork - the schema regeneration, the enum-tolerance patch, the
`xmlrpc` package rewrite (originally vendored, then rewritten from scratch
once that turned out to have no license, along the way fixing its last
remaining bugs and making it measurably faster - see `xmlrpc/doc.go`), the
documentation pass, this test suite, and the CI automation - was built end
to end by [Claude Code](https://claude.com/claude-code) (Anthropic's
coding agent) working with the maintainer in a single ongoing conversation
- starting from a client that panicked against a modern XCP-ng host
because upstream's schema predated it by years, ending at a documented,
tested, self-updating private fork with no external dependencies at all.

## Implementation notes

Most of the code in this library is generated from a description of the XenAPI.
This description is the file `xenapi.json`, the source of which is the XenAPI
documentation at http://xapi-project.github.io/:

- https://github.com/xapi-project/xapi-project.github.io/tree/master/_data

**Licensing note**: `xapi-project.github.io` carries no LICENSE file. Unlike
the `xmlrpc` package (see above), this isn't vendored/copied source code -
it's the XenAPI's own published interface description (class names, field
names, types, and documentation strings), whose explicit purpose is to let
third parties generate client libraries against it, the same role an
OpenAPI/Swagger spec or `.proto` file plays for other APIs. That doesn't
make the absence of a license a non-issue, just a materially different and
lower-risk one than redistributing someone's general-purpose source code
with no license at all.

The list of error code constants in `error.go` is derived from the OCaml
source of truth (note: unlike the `*_gen.go` files, it is **not** covered by
`xenapi.json` / `go generate` at all, so it needs a separate refresh
whenever it goes stale):

- https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi-consts/api_errors.ml
  (moved from `ocaml/idl/api_errors.ml` at some point upstream)

`xen-api` is LGPL-2.1. What's extracted here is a mechanical list of bare
`NAME -> "STRING"` pairs (arguably not creative expression at all, more
like extracting a list of HTTP status codes) - not the surrounding OCaml
code, comments, or structure - so this is a much lower-risk case than the
`xmlrpc` situation above, but noted here for the same reason.

Format is `let name = add_error "VALUE"`, with a handful of entries built by
string concatenation of two previously-defined names/literals (e.g.
`add_error $ auth_enable_failed ^ auth_suffix_invalid_ou`) instead of a
plain literal — resolve those the same way if refreshing this file again.
Go constant names are mechanically derived as `ERR_` + the uppercased OCaml
identifier, not hand-picked, so name and value can't drift out of sync.

The JSON file contains the lifecycle of published classes, fields and messages.
Each of the release names can be mapped back to a version listed here:

- https://xapi-project.github.io/xen-api/releases.html

## Automated regeneration

`.github/workflows/regenerate.yml` runs weekly (and on manual dispatch) and
does the full regeneration below by itself:

1. Fetches the current `xenapi.json`; exits immediately if it's unchanged.
2. Regenerates all `*_gen.go` files (`go run xenapi.go`).
3. Re-applies the enum-tolerance patch (`go run patch_enums.go`).
4. Regenerates `error.go` from the upstream OCaml error definitions
   (`go run gen_errors.go`) — this source isn't part of `xenapi.json` at all
   and would otherwise silently keep drifting (see "Implementation notes"
   below).
5. Updates the `SchemaXAPIRelease` constant (`go run update_schema_release.go`).
6. Runs `go build`/`go vet`; if either fails, or if step 2 panics on an
   unrecognized schema construct, **the workflow fails loudly and opens a
   GitHub issue** rather than pushing anything broken.
7. On success: commits everything, bumps the minor version (e.g. `v0.1.0` ->
   `v0.2.0`), tags, and pushes.

The one thing this **can't** automate: if `xenapi.json` introduces a schema
construct `xenapi.go` has never seen (the way `"an event batch"` or
`"X option"` showed up in this fork's last manual regeneration), generation
panics and step 6 catches it — but teaching the generator that new construct
still needs a human. That's an inherent limit, not a workflow bug: whether a
brand-new type pattern is safe to map to `xmlrpc.Struct`, needs a real Go
type, or something else entirely, is a judgment call the generator can't
make for itself.

Each of the three `go run <file>.go` tools above is also meant to be run
locally the same way, independent of CI, if you're regenerating by hand.

## Regenerating API after xenapi.json update (manual / what the tools above do)
If XenAPI was updated, it is required to regenerate all of files with a new API description. In order to do that one needs to follow these steps:
- Get newest `xenapi.json` from the link above (e.g.
  `https://raw.githubusercontent.com/xapi-project/xapi-project.github.io/master/_data/xenapi.json`).
- Delete old generated APIs using `rm *_gen.go`
- The generator (`xenapi.go`) is itself tagged `// +build ignore`, so `go.mod`
  doesn't carry its own dependency (`github.com/serenize/snaker`); fetch it
  once with `go get github.com/serenize/snaker` before generating, and run
  `go mod tidy` afterwards to drop it again (matches upstream's own `go.mod`).
- Generate new API with `go generate` (or `go run xenapi.go` directly)
- **Reapply the enum-tolerance patch** described above — it lives only in
  `convert_gen.go` and gets wiped out by regeneration along with every other
  generated file. `go run patch_enums.go` does this mechanically; by hand,
  it's a find/replace over every
  `default: err = fmt.Errorf("... but this is not any of the known values" ...)`
  case in that file, turning it into `value = <EnumType>(strValue)`.
- If the schema introduced a new construct `xenapi.go` doesn't know how to
  map yet, generation panics with a message like `Unsupported XenAPI type: ...`
  naming the offending type string — add a case for it in `goTypeForXenType`,
  `funcPartialForXenType`, and `buildConverterFunc` (see the four cases added
  for this fork's last regeneration, above, as examples).
