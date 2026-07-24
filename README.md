# Go XenAPI client library (fork)

This is a private fork of [terra-farm/go-xen-api-client](https://github.com/terra-farm/go-xen-api-client),
originally forked from [v0.0.2](https://github.com/terra-farm/go-xen-api-client/releases/tag/v0.0.2)
(a schema snapshot only current through XenServer 7.3 "inverness", ~2017).

Module path is `github.com/schankst/go-xen-api-client` so it can be pulled
in directly via `go get`/`require` without a `replace` directive.

## What's different from upstream

1. **Bindings regenerated from the current XenAPI schema** (as of the last
   regeneration, covering releases up to `26.16.1-next`), instead of the
   original v0.0.2 snapshot. This is what makes the library work at all
   against a modern XCP-ng host — v0.0.2 predates entire classes and fields
   that current hosts report.

   Regenerating from a schema this much newer than the generator (`xenapi.go`)
   itself required teaching the generator four new schema constructs it
   didn't understand yet:

   - `lifecycle` changed from a bare array to an object (`{state, transitions}`).
   - New opaque result type `"an event batch"` (event batching, used by
     `Event.from`) — mapped to `xmlrpc.Struct` since it isn't a proper record
     and this fork doesn't need it typed.
   - New polymorphic field type `` `<class> record` `` (`Event.snapshot` — the
     concrete type depends on the event's class at runtime) — likewise mapped
     to `xmlrpc.Struct`.
   - New `` `X option` `` type pattern (optional values) — added as a generic
     nil-tolerant wrapper around the inner type's own converter.

   Along the way, some enums (e.g. `CertificatePurpose`, `UpdateGuidances`)
   turned out to be declared under more than one class in the newer schema;
   the generator now tracks already-emitted enum names so it doesn't emit
   the same Go type twice.

2. **Tolerate unknown enum values from newer XAPI versions**, as a hedge
   against anything even newer than the schema above. The generated enum
   parsers otherwise hard-error when the server returns a value that isn't
   in the schema the client was generated from (this is what originally broke
   on the VM operation `"sysprep"` against v0.0.2). All generated `*ToGo`
   enum converters in `convert_gen.go` (72 of them, one per enum type) pass
   unknown values through as-is instead of failing the whole record parse.

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

The most important missing pieces before this library is production-ready are:

- A strategy how to handle the differences in the XenAPI versions.
- Tests, at least for the various data type conversions.
- Embed XenAPI documentation as GoDoc in the generated code.
- Better error messages.
- Usage examples.

Contributions welcome!

Please note that I want to keep this library lean. I envision it to merely
provide a one-to-one mapping of XenAPI functions to Go functions. Because of
this, I will likely not accept pull requests that implement higher level
functionality.

## Implementation notes

Most of the code in this library is generated from a description of the XenAPI.
This description is the file `xenapi.json`, the source of which is the XenAPI
documentation at http://xapi-project.github.io/:

- https://github.com/xapi-project/xapi-project.github.io/tree/master/_data

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

The JSON file contains the lifecycle of published classes, fields and messages.
Each of the release names can be mapped back to a version listed here:

- https://xapi-project.github.io/xen-api/releases.html

## Regenerating API after xenapi.json update
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
  generated file. It's a mechanical find/replace over every
  `default: err = fmt.Errorf("... but this is not any of the known values" ...)`
  case in that file, turning it into `value = <EnumType>(strValue)`.
- If the schema introduced a new construct `xenapi.go` doesn't know how to
  map yet, generation panics with a message like `Unsupported XenAPI type: ...`
  naming the offending type string — add a case for it in `goTypeForXenType`,
  `funcPartialForXenType`, and `buildConverterFunc` (see the four cases added
  for this fork's last regeneration, above, as examples).
