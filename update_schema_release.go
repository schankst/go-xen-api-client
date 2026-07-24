//go:build ignore
// +build ignore

// update_schema_release.go scans xenapi.json for the newest XAPI release
// version and rewrites the SchemaXAPIRelease constant in client.go to
// match, after a regeneration.
//
// Run with: go run update_schema_release.go
package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

var (
	reRelease = regexp.MustCompile(`"release"\s*:\s*"([0-9]+)\.([0-9]+)\.([0-9]+)(-next)?"`)
	reConst   = regexp.MustCompile(`const SchemaXAPIRelease = "[^"]*"`)
)

type version struct {
	major, minor, patch int
	next                bool
	raw                 string
}

func (v version) less(o version) bool {
	if v.major != o.major {
		return v.major < o.major
	}
	if v.minor != o.minor {
		return v.minor < o.minor
	}
	if v.patch != o.patch {
		return v.patch < o.patch
	}
	// Same X.Y.Z: the "-next" pre-release represents the version being
	// worked on after that patch shipped, so it's newer.
	return !v.next && o.next
}

func main() {
	data, err := os.ReadFile("xenapi.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "reading xenapi.json:", err)
		os.Exit(1)
	}

	matches := reRelease.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		fmt.Fprintln(os.Stderr, "no numeric release versions found in xenapi.json")
		os.Exit(1)
	}

	var latest version
	for _, m := range matches {
		major, _ := strconv.Atoi(m[1])
		minor, _ := strconv.Atoi(m[2])
		patch, _ := strconv.Atoi(m[3])
		v := version{major, minor, patch, m[4] == "-next", fmt.Sprintf("%d.%d.%d%s", major, minor, patch, m[4])}
		if latest.raw == "" || latest.less(v) {
			latest = v
		}
	}

	clientGo, err := os.ReadFile("client.go")
	if err != nil {
		fmt.Fprintln(os.Stderr, "reading client.go:", err)
		os.Exit(1)
	}
	if !reConst.Match(clientGo) {
		fmt.Fprintln(os.Stderr, "SchemaXAPIRelease const not found in client.go")
		os.Exit(1)
	}
	updated := reConst.ReplaceAll(clientGo, []byte(fmt.Sprintf(`const SchemaXAPIRelease = %q`, latest.raw)))

	if err := os.WriteFile("client.go", updated, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "writing client.go:", err)
		os.Exit(1)
	}

	fmt.Printf("SchemaXAPIRelease set to %q\n", latest.raw)
}
