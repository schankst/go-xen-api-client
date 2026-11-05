# Licensing

The code in this repository is MIT-licensed (see [LICENSE](LICENSE)). It is
generated from, or built against, these sources:

| Source | License | Use |
| --- | --- | --- |
| [`xenapi.json`](https://github.com/xapi-project/xapi-project.github.io/tree/master/_data) | none stated (checked October 2026) | XenAPI's published interface description, from which the `*_gen.go` files are generated |
| [`api_errors.ml`](https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi-consts/api_errors.ml) | LGPL-2.1 | only the error code names and values are extracted into `error.go` |

`xmlrpc/` is this repository's own implementation. Versions before v0.2.0
contained a vendored copy of the unlicensed amfranz/go-xmlrpc-client; that
code was replaced entirely in v0.2.0.
