# Maintaining

## Sources

| Output | Generator | Source |
| --- | --- | --- |
| `*_gen.go`, `docs/lifecycle.md` | `internal/cmd/genapi` (`go generate ./...`) | [`xenapi.json`](https://github.com/xapi-project/xapi-project.github.io/tree/master/_data) |
| `error.go` | `internal/cmd/generrors` | [`api_errors.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi-consts/api_errors.ml) |
| `SchemaXAPIRelease` in `client.go` | `internal/cmd/schemarelease` | `xenapi.json` |

Run all three from the repo root. Generate with the current stable Go: the
output goes through `go/format`, whose rules change between releases. The
`go` line in `go.mod` is only the minimum for users.

If the schema uses a type construct the generator doesn't know, generation
fails naming the type. Add a case to `goTypeForXenType`,
`funcPartialForXenType` and `buildConverterFunc` in `internal/cmd/genapi`.

## Lifecycle

`xenapi.json` records when each class, field and message was published,
deprecated or removed (release names: see
[releases](https://xapi-project.github.io/xen-api/releases.html)). Removed
items are still generated, since older hosts support them, and are marked
`Deprecated:` like deprecated ones. Prototyped items get an `Experimental:`
note. Items that exist only in an unreleased `-next` version are skipped,
and `SchemaXAPIRelease` ignores `-next` versions.

## Error codes

Constants are named `Err` plus the CamelCase OCaml identifier; the value is
the `add_error` literal sent on the wire. For 27 entries the two differ
beyond case (e.g. `ErrSRNoPBDs` = `"SR_HAS_NO_PBDS"`, and
`ErrVMSnapshotWithQuiescePluginDoesNotRespond` = `"..._DEOS_NOT_RESPOND"`, a
typo upstream); those carry a `// The wire value is "..."` comment. Never
"fix" a value: it is what the server sends.

## Tests

- `TestGolden` (`internal/cmd/genapi`) generates from `testdata/schema.json`,
  which uses every supported type shape and lifecycle state, and compares the
  output with `testdata/golden`. Refresh with `go test ./internal/cmd/genapi
  -update` and review the diff.
- `TestCommittedBindingsUpToDate` fails if the committed `*_gen.go` files
  don't match the committed `xenapi.json`.

## Automated regeneration

`.github/workflows/regenerate.yml` runs weekly, or manually. It fetches both
sources, regenerates everything and stops if nothing changed. Otherwise it
commits to a `regenerate/<release>-<run>` branch, opens a PR (closing older
open ones), runs `test.yml` on it, and only if that passes merges the PR,
tags the next minor version and creates a release. On failure it opens an
issue.
