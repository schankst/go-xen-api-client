# Licensing notes

This fork's own code is under the same [MIT license](LICENSE) as upstream.
This file documents the licensing situation of everything it's built
from or against, for anyone doing due diligence before depending on it.

## `xmlrpc` (this repo's own code, not a dependency)

Until v0.1.1, `xmlrpc/` was the
[amfranz/go-xmlrpc-client](https://github.com/amfranz/go-xmlrpc-client)
module, pulled in as an ordinary compiled dependency; it has **no license
file, and never had one**. v0.1.2 then vendored its source into this repo
as `xmlrpc/`, which kept that unlicensed code in place but stopped it
being a module dependency. Vendoring did **not** resolve the licensing
question - it made it more visible: from v0.1.2 on, unlicensed third-party
source code was committed and redistributed in this repo (v0.1.2-v0.1.5).

So `xmlrpc` was **rewritten from scratch** in v0.2.0: same public contract
(`Client`, `NewClient`, `Struct`, `Base64`, `Call`), materially different
architecture (no `net/rpc`, `strings.Builder`-based encoding instead of
string concatenation, a proper recursive-descent XML decoder instead of
regex-based response parsing — see `xmlrpc/doc.go` and the v0.2.0 release
notes for what changed and the benchmarks that came out of it). No part
of this repo is unlicensed third-party source code anymore.

## `xenapi.json` (data, not code — used to generate most of this package)

Source: https://github.com/xapi-project/xapi-project.github.io — also
carries **no LICENSE file** (checked October 2026: the repo root has no
`LICENSE`/`COPYING`, and GitHub's license API reports the repository as
unlicensed).

This is a materially different situation from the `xmlrpc` case above:
it isn't general-purpose source code, it's XenAPI's own published
interface description (class names, field names, types, and
documentation strings), whose explicit purpose is to let third parties
generate client libraries against it — the same role an OpenAPI/Swagger
spec or a `.proto` file plays for other APIs. That doesn't make the
absence of a license a non-issue, just a lower-risk one than
redistributing someone's general-purpose source code with no license at
all: this fork is doing exactly what the data was published for.

## `error.go`'s source data

Source: https://github.com/xapi-project/xen-api/blob/master/ocaml/xapi-consts/api_errors.ml
— licensed **LGPL-2.1** (`xen-api`'s repo-wide license).

What's extracted from it here is a mechanical list of bare
`NAME -> "STRING"` pairs (arguably not creative expression at all, more
like extracting a list of HTTP status codes) — not the surrounding OCaml
code, comments, or structure. This is a much lower-risk case than either
of the above, but noted here for the same reason: so it's documented
rather than silently assumed fine.
