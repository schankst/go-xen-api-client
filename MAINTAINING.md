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
Go constant names are mechanically derived as `ERR_` + the uppercased OCaml
identifier, not hand-picked, so name and value can't drift out of sync.

## Automated regeneration

`.github/workflows/regenerate.yml` runs weekly (and on manual dispatch) and
does the full regeneration below by itself:

1. Fetches the current `xenapi.json`; exits immediately if it's unchanged.
2. Regenerates all `*_gen.go` files (`go run xenapi.go`) - enum parsing is
   tolerant of unknown values by construction (see "Implementation
   notes"), so this alone is enough; no separate patch step.
3. Regenerates `error.go` from the upstream OCaml error definitions
   (`go run gen_errors.go`) — this source isn't part of `xenapi.json` at all
   and would otherwise silently keep drifting (see "Implementation notes"
   above).
4. Updates the `SchemaXAPIRelease` constant (`go run update_schema_release.go`).
5. Runs `go build`/`go vet`/`go test -short`; if any fail, or if step 2
   panics on an unrecognized schema construct, **the workflow fails
   loudly and opens a GitHub issue** rather than pushing anything broken.
6. On success: commits everything, bumps the minor version (e.g. `v0.1.0` ->
   `v0.2.0`), tags, and pushes.

The one thing this **can't** automate: if `xenapi.json` introduces a schema
construct `xenapi.go` has never seen (the way `"an event batch"` or
`"X option"` showed up in this fork's last manual regeneration), generation
panics and step 5 catches it — but teaching the generator that new construct
still needs a human. That's an inherent limit, not a workflow bug: whether a
brand-new type pattern is safe to map to `xmlrpc.Struct`, needs a real Go
type, or something else entirely, is a judgment call the generator can't
make for itself.

Each of the two `go run <file>.go` tools above is also meant to be run
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
- Generate new API with `go generate` (or `go run xenapi.go` directly) -
  enum tolerance is baked into `convertEnumTypeToGoFuncTemplate`, so
  `convert_gen.go` comes out correct with no separate patch step.
- If the schema introduced a new construct `xenapi.go` doesn't know how to
  map yet, generation panics with a message like `Unsupported XenAPI type: ...`
  naming the offending type string — add a case for it in `goTypeForXenType`,
  `funcPartialForXenType`, and `buildConverterFunc` (see `xenapi.go`'s own
  file-level comment and the existing cases for examples).

## Formatting of generated code
Both generators render into memory and run the result through `go/format`
before writing it, so `*_gen.go` and `error.go` are gofmt-clean by
construction and never need hand-formatting — `test.yml` enforces that with
a `gofmt -l .` check over the whole repository.

The consequence is that the toolchain you generate with decides what the
committed files look like, because gofmt's rules change between Go releases
(Go 1.19's doc-comment reformatting, for one). Regenerate with a current Go,
not with the `go 1.16` from `go.mod` — that's the library's minimum for
*consumers*, verified separately by `test.yml`'s version matrix, and both CI
workflows pin the generators to `stable` for the same reason. Generating with
an older toolchain would reformat every generated file and show up as a large
whitespace-only diff.
