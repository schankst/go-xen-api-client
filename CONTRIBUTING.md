# Contributing

Issues and pull requests are welcome.

Most of the code is generated. Don't edit `*_gen.go` or `error.go` by hand;
change the generators in `internal/cmd/` and regenerate (see
[MAINTAINING.md](MAINTAINING.md)).

Before opening a PR, run `go vet ./...` and `go test -race ./...`.
