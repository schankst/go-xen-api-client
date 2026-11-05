# Maintaining this fork

How the generated bindings and `error.go` are produced, and how to refresh
them by hand if you ever need to. See [CONTRIBUTING.md](CONTRIBUTING.md)
for the PR/issue process, and [LICENSING.md](LICENSING.md) for why the
upstream data sources referenced below are safe to build against.

## Implementation notes

Most of the code in this library is generated from a description of the XenAPI.
This description is the file `xenapi.json`, the source of which is the XenAPI
documentation at http://xapi-project.github.io/:

- https://github.com/xapi-project/xapi-project.github.io/tree/master/_data

The JSON file contains the lifecycle of published classes, fields and messages.
Each of the release names can be mapped back to a version listed here:

- https://xapi-project.github.io/xen-api/releases.html

The list of error code constants in `error.go` is derived from the OCaml
source of truth (note: unlike the `*_gen.go` files, it is **not** covered by
`xenapi.json` / `go generate` at all, so it needs a separate refresh
whenever it goes stale):

- https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi-consts/api_errors.ml
  (moved from `ocaml/idl/api_errors.ml` at some point upstream)

Format is `let name = add_error "VALUE"`, with a handful of entries built by
string concatenation of two previously-defined names/literals (e.g.
`add_error $ auth_enable_failed ^ auth_suffix_invalid_ou`) instead of a
plain literal — resolve those the same way if refreshing this file again.
Go constant names are mechanically derived as `Err` + the CamelCase OCaml
identifier (`naming.SnakeToCamelXenAPI`), not hand-picked, so name and value
can't drift out of sync. Each constant has type `ErrorCode`.

### Constant names follow the OCaml identifier, not always the wire value

The XML-RPC value a failed call actually carries in
`ErrorDescription[0]` — and therefore `Error.Code` —
is the `add_error` literal, *not* the OCaml identifier. They usually differ
only in case (`ErrVMBadPowerState` = `"VM_BAD_POWER_STATE"`), but about
two dozen entries upstream don't follow that pattern:

| Name | Wire value | Why |
| --- | --- | --- |
| `ErrSRNoPBDs` | `SR_HAS_NO_PBDS` | abbreviated identifier |
| `ErrVMToomanyVCPUs` | `VM_TOO_MANY_VCPUS` | missing underscore |
| `ErrPoolJoiningHostCannot*` | `JOINING_HOST_CANNOT_*` | prefix only on the identifier |
| `ErrVMSnapshotWithQuiescePluginDoesNotRespond` | `..._PLUGIN_DEOS_NOT_RESPOND` | typo in the **value** |

`internal/cmd/generrors` emits a `// The wire value is "..."` comment above exactly
those constants, and prints how many it found, so the exception is visible
at the point of use instead of having to be remembered. Two rules follow
from that, for consumers and for anyone editing this file by hand:

- Match against the constant (`errors.Is(err, xenapi.ErrSRNoPBDs)`), never
  against a hand-typed string. A string built from the constant's own name
  matches only the majority case.
- Don't "fix" a value to match its name. The value is the wire contract: the
  server sends these bytes, typos included.

## Automated regeneration

`.github/workflows/regenerate.yml` runs weekly, or manually; the `force`
input regenerates even when `xenapi.json` is unchanged, e.g. after a
generator change. It:

1. Fetches the current `xenapi.json` and stops if it is unchanged.
2. Regenerates the bindings (`go generate ./...`), `error.go`
   (`go run ./internal/cmd/generrors`) and `SchemaXAPIRelease`
   (`go run ./internal/cmd/schemarelease`).
3. Commits the result to a `regenerate/<release>-<run>` branch and opens a PR.
4. Runs `test.yml` on that branch (PRs opened by the workflow token don't
   trigger it on their own).
5. If the tests pass: merges the PR, tags the merge commit with the next minor
   version (`v0.10.0` -> `v0.11.0`) and creates a GitHub release.

If any step fails, the workflow opens an issue linking the run and the PR, if
there is one. A schema construct the generator has never seen (as `"an event
batch"` and `"X option"` once were) makes step 2 fail with the type's name;
mapping it needs a human decision.

All three commands also run locally from the repo root.

## Regenerating API after xenapi.json update (manual / what the tools above do)
If XenAPI was updated, it is required to regenerate all of files with a new API description. In order to do that one needs to follow these steps:
- Get newest `xenapi.json` from the link above (e.g.
  `https://raw.githubusercontent.com/xapi-project/xapi-project.github.io/master/_data/xenapi.json`).
- Delete old generated APIs using `rm *_gen.go`
- Generate new API with `go generate ./...`.
- If the schema introduced a new construct the generator doesn't know how to
  map yet, generation fails with a message like `Unsupported XenAPI type: ...`
  naming the offending type string — add a case for it in `goTypeForXenType`,
  `funcPartialForXenType`, and `buildConverterFunc` (see
  `internal/cmd/genapi`'s package comment and the existing cases for examples).

## Lifecycle

`xenapi.json` records when each class, field and message was published,
deprecated or removed. The generator keeps removed items, since older hosts
still support them, and marks removed and deprecated ones with a
`// Deprecated:` paragraph naming the XAPI release. Prototyped items get an
`Experimental:` note. Items introduced only in an unreleased `-next` version
are skipped, and `SchemaXAPIRelease` ignores `-next` versions.

## Generator tests

`go test ./internal/cmd/genapi` covers the generator:

- `TestGolden` generates from `testdata/schema.json`, a small schema that uses
  every supported type shape, and compares the result with `testdata/golden`.
  After an intended template change, run it with `-update` and review the
  golden diff.
- `TestCommittedBindingsUpToDate` fails if the committed `*_gen.go` files
  don't match what the generator produces for the committed `xenapi.json`.

## Formatting of generated code
Both generators render into memory and run the result through `go/format`
before writing it, so `*_gen.go` and `error.go` are gofmt-clean by
construction and never need hand-formatting — `test.yml` enforces that with
a `gofmt -l .` check over the whole repository.

gofmt's rules change between Go releases, so the toolchain you generate with
decides what the committed files look like. Regenerate with the current stable
Go, as both CI workflows do; an older toolchain shows up as a large
whitespace-only diff. The `go` line in `go.mod` is only the minimum for
consumers, verified by `test.yml`'s version matrix.
