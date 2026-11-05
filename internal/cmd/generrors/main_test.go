package main

import (
	"slices"
	"strings"
	"testing"
)

func TestChunksFromSource(t *testing.T) {
	src := `(* comment *)
let a = add_error "A"

let b =
  add_error "B"
(* another *)
let c = "x"
`
	got := chunksFromSource(strings.NewReader(src))
	want := []string{`let a = add_error "A"`, `let b = add_error "B"`, `let c = "x"`}
	if !slices.Equal(got, want) {
		t.Errorf("chunksFromSource() = %q, want %q", got, want)
	}
}

func TestParseEntries(t *testing.T) {
	chunks := []string{
		`let message_deprecated = add_error "MESSAGE_DEPRECATED"`,
		`let auth_suffix = "_SUFFIX"`,
		`let joined = add_error $ message_deprecated ^ auth_suffix`,
		`let with_lit = add_error $ message_deprecated ^ "_LIT"`,
		`let spaced  =   add_error   "SPACED"`,
		`let unresolved = add_error $ missing ^ auth_suffix`,
		`let unresolved_lit = add_error $ missing ^ "_X"`,
		`let something_else = 42`,
	}
	entries, skipped := parseEntries(chunks)

	want := []ordEntry{
		{"message_deprecated", "MESSAGE_DEPRECATED"},
		{"joined", "MESSAGE_DEPRECATED_SUFFIX"},
		{"with_lit", "MESSAGE_DEPRECATED_LIT"},
		{"spaced", "SPACED"},
	}
	if !slices.Equal(entries, want) {
		t.Errorf("entries = %v, want %v", entries, want)
	}
	if len(skipped) != 3 {
		t.Errorf("skipped = %q, want the two unresolved concatenations and the non-error binding", skipped)
	}
}
