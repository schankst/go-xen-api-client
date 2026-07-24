# Contributing

This is a personal fork, maintained mainly for the maintainer's own use.
Issues (bug reports, questions, "this XenAPI version broke it") are
welcome. Pull requests are welcome too, especially for:

- Schema regeneration issues (new XenAPI constructs `xenapi.go` doesn't
  understand yet - see the "Automated regeneration" section in the
  README for what that usually looks like).
- Bugs in the hand-maintained parts (`client.go`, `xmlrpc/`, the
  generator tools).

Given most of the code is generated (see "Implementation notes" in the
README), please don't send PRs that hand-edit `*_gen.go` files or
`error.go` directly - those changes will be silently lost the next time
the regeneration workflow runs. Fix the generator (`xenapi.go`,
`gen_errors.go`) or the schema-fetch step instead, then regenerate.

Before opening a PR: `go build ./...`, `go vet ./...`, and
`go test -short ./...` should all pass.
