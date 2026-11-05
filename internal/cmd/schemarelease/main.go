// Command schemarelease sets SchemaXAPIRelease in client.go to the newest
// released version named in xenapi.json. Run it from the repo root.
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
	raw                 string
}

func (v version) less(o version) bool {
	if v.major != o.major {
		return v.major < o.major
	}
	if v.minor != o.minor {
		return v.minor < o.minor
	}
	return v.patch < o.patch
}

// latestRelease returns the newest released version named in schema.
// "-next" versions are unreleased and ignored.
func latestRelease(schema []byte) (string, bool) {
	var latest version
	for _, m := range reRelease.FindAllSubmatch(schema, -1) {
		major, _ := strconv.Atoi(string(m[1]))
		minor, _ := strconv.Atoi(string(m[2]))
		patch, _ := strconv.Atoi(string(m[3]))
		if len(m[4]) > 0 {
			continue
		}
		v := version{major, minor, patch, fmt.Sprintf("%d.%d.%d", major, minor, patch)}
		if latest.raw == "" || latest.less(v) {
			latest = v
		}
	}
	return latest.raw, latest.raw != ""
}

func main() {
	data, err := os.ReadFile("xenapi.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "reading xenapi.json:", err)
		os.Exit(1)
	}

	latest, ok := latestRelease(data)
	if !ok {
		fmt.Fprintln(os.Stderr, "no numeric release versions found in xenapi.json")
		os.Exit(1)
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
	updated := reConst.ReplaceAll(clientGo, fmt.Appendf(nil, `const SchemaXAPIRelease = %q`, latest))

	if err := os.WriteFile("client.go", updated, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "writing client.go:", err)
		os.Exit(1)
	}

	fmt.Printf("SchemaXAPIRelease set to %q\n", latest)
}
