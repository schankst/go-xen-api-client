//go:build ignore
// +build ignore

// gen_errors.go regenerates error.go from the upstream xen-api error
// definitions (a separate source from xenapi.json, not covered by
// `go generate` / xenapi.go - see MAINTAINING.md's "Implementation notes"
// for the source and the name/value rule, and README.md's "Versioning" for
// how it relates to the schema release).
//
// Run with: go run gen_errors.go
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

	chunks := chunksFromSource(body)
	entries, skipped := parseEntries(chunks)

	seen := map[string]bool{}
	dupCount := 0
	for _, e := range entries {
		if seen[e.name] {
			dupCount++
			fmt.Fprintln(os.Stderr, "duplicate entry name:", e.name)
		}
		seen[e.name] = true
	}
	if dupCount > 0 {
		fmt.Fprintf(os.Stderr, "%d duplicate name(s) found; aborting rather than emitting an invalid error.go\n", dupCount)
		os.Exit(1)
	}
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

	// Rendered into memory and run through go/format before being written,
	// so error.go is gofmt-clean by construction (the const block's
	// alignment depends on the longest error name, which changes upstream).
	w := &bytes.Buffer{}

	fmt.Fprint(w, `package xenapi

import (
	"strings"
)

const (
`)
	// Most constants have ERR_<UPPERCASED IDENTIFIER> = "<VALUE>" with the
	// two matching barring case. About two dozen upstream entries don't:
	// the OCaml identifier is abbreviated, carries a prefix the wire value
	// lacks, or the value itself has a typo (DEOS for DOES). A reader who
	// assumes the wire value is just the constant's name lowercased then
	// compares against a string that never matches, or "fixes" a documented
	// typo. Flag exactly those, and only those, so the marker stays
	// meaningful instead of appearing on every line.
	//
	// A flagged constant closes the preceding const block and starts its
	// own, so a comment never interrupts the alignment run of its
	// neighbours - inside one block gofmt aligns every line against the
	// longest name, and a flagged entry's comment would otherwise drag
	// unrelated constants into a wider column.
	divergent := 0
	openBlock := false
	for _, e := range entries {
		flagged := !strings.EqualFold(e.name, e.value)
		if flagged {
			divergent++
			if openBlock {
				fmt.Fprint(w, ")\n\nconst (\n")
			}
			fmt.Fprintf(w, "\t// The wire value is %q - it does not match the constant's name.\n", e.value)
			fmt.Fprintf(w, "\tERR_%s = %q\n", strings.ToUpper(e.name), e.value)
			fmt.Fprint(w, ")\n\nconst (\n")
			openBlock = true
			continue
		}
		fmt.Fprintf(w, "\tERR_%s = %q\n", strings.ToUpper(e.name), e.value)
	}
	fmt.Fprint(w, `)

// Error is returned for any XenAPI call that fails at the protocol level.
// Its Error() string is a fallback for display; check Code() (one of the
// ERR_ constants above) to distinguish error kinds programmatically.
type Error struct {
	code    string
	objtype string
	uuid    string
}

// Error implements the error interface. objtype/uuid are only present for
// error codes that carry them (see Type/UUID) and are omitted when empty,
// rather than rendered as blank fields.
func (e *Error) Error() string {
	parts := []string{"API Error:", e.code}
	if e.objtype != "" {
		parts = append(parts, e.objtype)
	}
	if e.uuid != "" {
		parts = append(parts, e.uuid)
	}
	return strings.Join(parts, " ")
}

// Code returns the XenAPI error code, e.g. ERR_HANDLE_INVALID.
func (e *Error) Code() string {
	return e.code
}

// Type returns the affected object's class name (e.g. "VM"), if the error
// code carries one - HANDLE_INVALID does, most others don't and this is "".
func (e *Error) Type() string {
	return e.objtype
}

// UUID returns the affected object's UUID, if the error code carries one -
// as with Type, only some error codes (e.g. HANDLE_INVALID) do.
func (e *Error) UUID() string {
	return e.uuid
}
`)

	formatted, err := format.Source(w.Bytes())
	if err != nil {
		fmt.Fprintln(os.Stderr, "generated error.go does not parse:", err)
		os.Exit(1)
	}
	if err := os.WriteFile("error.go", formatted, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "writing error.go:", err)
		os.Exit(1)
	}

	fmt.Printf("wrote %d error constants to error.go (%d with a wire value that differs from their name, %d source chunks skipped as non-error boilerplate)\n", len(entries), divergent, len(skipped))
}
