// Command generrors regenerates error.go from the upstream xen-api error
// definitions (a separate source from xenapi.json, not covered by
// `go generate` / genapi - see MAINTAINING.md's "Implementation notes"
// for the source and the name/value rule, and README.md's "Versioning" for
// how it relates to the schema release).
//
// Run from the repo root: go run ./internal/cmd/generrors
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/schankst/go-xen-api-client/internal/naming"
)

const sourceURL = "https://raw.githubusercontent.com/xapi-project/xen-api/master/ocaml/xapi-consts/api_errors.ml"

var (
	reSimple     = regexp.MustCompile(`^let\s+([A-Za-z0-9_]+)\s*=\s*add_error\s+"([A-Za-z0-9_]+)"$`)
	reConcat     = regexp.MustCompile(`^let\s+([A-Za-z0-9_]+)\s*=\s*add_error\s+\$\s+([A-Za-z0-9_]+)\s*\^\s*([A-Za-z0-9_]+)$`)
	reConcatLit  = regexp.MustCompile(`^let\s+([A-Za-z0-9_]+)\s*=\s*add_error\s+\$\s+([A-Za-z0-9_]+)\s*\^\s*"([^"]*)"$`)
	reStrValue   = regexp.MustCompile(`^let\s+([A-Za-z0-9_]+)\s*=\s*"([^"]*)"$`)
	reComment    = regexp.MustCompile(`^\(\*.*\*\)$`)
	reWhitespace = regexp.MustCompile(`\s+`)
)

type ordEntry struct {
	name  string
	value string
}

func fetchSource() (io.ReadCloser, error) {
	resp, err := http.Get(sourceURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status %s fetching %s", resp.Status, sourceURL)
	}
	return resp.Body, nil
}

func chunksFromSource(r io.Reader) []string {
	var chunks []string
	var cur []string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || reComment.MatchString(line) {
			if len(cur) > 0 {
				chunks = append(chunks, strings.Join(cur, " "))
				cur = nil
			}
			continue
		}
		cur = append(cur, line)
	}
	if len(cur) > 0 {
		chunks = append(chunks, strings.Join(cur, " "))
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return chunks
}

func parseEntries(chunks []string) (entries []ordEntry, skipped []string) {
	vars := map[string]string{}

	for _, chunk := range chunks {
		norm := reWhitespace.ReplaceAllString(chunk, " ")
		if m := reSimple.FindStringSubmatch(norm); m != nil {
			entries = append(entries, ordEntry{m[1], m[2]})
			vars[m[1]] = m[2]
			continue
		}
		if m := reConcat.FindStringSubmatch(norm); m != nil {
			a, okA := vars[m[2]]
			b, okB := vars[m[3]]
			if !okA || !okB {
				skipped = append(skipped, "UNRESOLVED CONCAT: "+norm)
				continue
			}
			val := a + b
			entries = append(entries, ordEntry{m[1], val})
			vars[m[1]] = val
			continue
		}
		if m := reConcatLit.FindStringSubmatch(norm); m != nil {
			a, okA := vars[m[2]]
			if !okA {
				skipped = append(skipped, "UNRESOLVED CONCAT-LIT: "+norm)
				continue
			}
			val := a + m[3]
			entries = append(entries, ordEntry{m[1], val})
			vars[m[1]] = val
			continue
		}
		if m := reStrValue.FindStringSubmatch(norm); m != nil {
			vars[m[1]] = m[2]
			continue
		}
		skipped = append(skipped, norm)
	}
	return
}

func main() {
	body, err := fetchSource()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetching source:", err)
		os.Exit(1)
	}
	defer body.Close()

	entries, skipped := parseEntries(chunksFromSource(body))
	if len(entries) < 400 {
		// Sanity floor: if the upstream format changed enough that most
		// entries silently fell into "skipped", fail loudly rather than
		// commit a near-empty error.go.
		fmt.Fprintf(os.Stderr, "only parsed %d entries (expected several hundred) - source format may have changed, see skipped chunks below:\n", len(entries))
		for _, s := range skipped {
			fmt.Fprintln(os.Stderr, "  SKIP:", s)
		}
		os.Exit(1)
	}

	src, divergent, err := render(entries)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile("error.go", src, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "writing error.go:", err)
		os.Exit(1)
	}

	fmt.Printf("wrote %d error codes to error.go (%d with a wire value that differs from their name, %d source chunks skipped as non-error boilerplate)\n", len(entries), divergent, len(skipped))
}

// render returns the gofmt-formatted error.go for entries: one ErrorCode
// constant per entry, named Err plus the CamelCase OCaml identifier. It
// fails if two entries map to the same Go name.
func render(entries []ordEntry) (src []byte, divergent int, err error) {
	var w bytes.Buffer
	// Usually the wire value is the OCaml identifier in upper case. About
	// two dozen entries differ (abbreviated identifier, extra prefix, or a
	// typo in the value itself); those get a comment and a const block of
	// their own, so the comment doesn't disturb the neighbours' alignment.
	var blocks [][]string
	plain := false // whether the last block takes further plain entries
	seen := map[string]string{}
	for _, e := range entries {
		name := "Err" + naming.SnakeToCamelXenAPI(e.name)
		if prev, dup := seen[name]; dup {
			return nil, 0, fmt.Errorf("%s and %s both map to %s", prev, e.name, name)
		}
		seen[name] = e.name

		line := fmt.Sprintf("\t%s ErrorCode = %q\n", name, e.value)
		if !strings.EqualFold(e.name, e.value) {
			divergent++
			blocks = append(blocks, []string{fmt.Sprintf("\t// The wire value is %q.\n", e.value), line})
			plain = false
			continue
		}
		if !plain {
			blocks = append(blocks, nil)
			plain = true
		}
		blocks[len(blocks)-1] = append(blocks[len(blocks)-1], line)
	}

	fmt.Fprint(&w, "// Code generated by internal/cmd/generrors; DO NOT EDIT.\n\npackage xenapi\n")
	for _, b := range blocks {
		fmt.Fprint(&w, "\nconst (\n"+strings.Join(b, "")+")\n")
	}

	src, err = format.Source(w.Bytes())
	if err != nil {
		return nil, 0, fmt.Errorf("generated error.go does not parse: %w", err)
	}
	return src, divergent, nil
}
