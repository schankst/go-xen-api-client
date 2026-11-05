# go-xen-api-client

[![Go Reference](https://pkg.go.dev/badge/github.com/schankst/go-xen-api-client.svg)](https://pkg.go.dev/github.com/schankst/go-xen-api-client)
[![Tests](https://github.com/schankst/go-xen-api-client/actions/workflows/test.yml/badge.svg)](https://github.com/schankst/go-xen-api-client/actions/workflows/test.yml)

A Go client for the complete [XenAPI](https://xapi-project.github.io/xen-api/)
of XCP-ng and XenServer, generated from XenAPI's own schema. Fork of
[terra-farm/go-xen-api-client](https://github.com/terra-farm/go-xen-api-client).

```
go get github.com/schankst/go-xen-api-client
```

Requires Go 1.25 or later. No dependencies outside the standard library.

## Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    xenapi "github.com/schankst/go-xen-api-client"
)

func main() {
    ctx := context.Background()

    xapi, err := xenapi.NewClient("https://10.0.0.10")
    if err != nil {
        log.Fatal(err)
    }

    session, err := xapi.Session.LoginWithPassword(ctx, "root", "password", "1.0", "example")
    if err != nil {
        log.Fatal(err)
    }
    defer xapi.Session.Logout(ctx, session)

    vms, err := xapi.VM.GetByNameLabel(ctx, session, "my-vm")
    if err != nil {
        log.Fatal(err)
    }
    if len(vms) != 1 {
        log.Fatal(fmt.Errorf("found %d VMs named my-vm, want 1", len(vms)))
    }

    if err := xapi.VM.Start(ctx, session, vms[0], false, false); err != nil {
        log.Fatal(err)
    }
}
```

## Usage notes

- **TLS** certificates are verified. For a host with a self-signed
  certificate, trust it with `xenapi.WithTLSConfig(&tls.Config{RootCAs: pool})`.
- **Context:** every call takes a `context.Context`; use it for deadlines and
  cancellation. The default HTTP client has no overall timeout.
- **Long operations:** `AsyncCall` starts a method as a XAPI task and
  `WaitTask` waits for it, cancelling the task on the host if the context
  ends.
- **Errors:** XenAPI failures are `*xenapi.Error` values with a `Code` and all
  `Params`. Match them with `errors.Is(err, xenapi.ErrHandleInvalid)`.
- **Versions:** `xenapi.SchemaXAPIRelease` names the XAPI release the bindings
  were generated from. Enum values from newer hosts pass through unchanged.
  APIs that XAPI deprecated or removed are marked `Deprecated:`; they still
  work against hosts that support them.

The bindings are regenerated weekly from the upstream schema; each
regeneration is a minor release. Tags follow this module's own SemVer, not
XAPI's version numbers.

## Documentation

- [docs/xapi-notes.md](docs/xapi-notes.md): XenAPI behaviour worth knowing, and fetching at scale
- [docs/import-export.md](docs/import-export.md): disk import/export and backups
- [docs/updates.md](docs/updates.md): finding pending updates and patches
- [docs/lifecycle.md](docs/lifecycle.md): removed, deprecated and experimental APIs
- [MAINTAINING.md](MAINTAINING.md): how the code is generated and released
- [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md), [LICENSING.md](LICENSING.md)

## License

MIT, see [LICENSE](LICENSE).
