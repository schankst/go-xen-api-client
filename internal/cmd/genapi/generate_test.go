package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/golden from the generator's output")

// TestGolden generates from testdata/schema.json, a small schema that uses
// every type shape the generator supports, and compares the output with
// testdata/golden. Run `go test -update` after an intended template change.
func TestGolden(t *testing.T) {
	out := generate(t, filepath.Join("testdata", "schema.json"))
	golden := filepath.Join("testdata", "golden")

	if *update {
		if err := os.RemoveAll(golden); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(golden, 0o755); err != nil {
			t.Fatal(err)
		}
		for _, name := range goFiles(t, out) {
			data, err := os.ReadFile(filepath.Join(out, name))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(golden, name+".golden"), data, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return
	}

	compare(t, out, golden, ".golden")
}

// TestCommittedBindingsUpToDate fails when the committed *_gen.go files
// differ from what the generator produces for the committed xenapi.json,
// i.e. when someone edited generated code or a template without
// regenerating.
func TestCommittedBindingsUpToDate(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	out := generate(t, filepath.Join(root, "xenapi.json"))
	compare(t, out, root, "")
}

func TestUnsupportedType(t *testing.T) {
	schema := filepath.Join(t.TempDir(), "schema.json")
	err := os.WriteFile(schema, []byte(`[{"name": "c", "fields": [{"name": "f", "type": "mystery shape"}]}]`), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	g := newAPIGenerator(t.TempDir())
	err = g.run(schema)
	if err == nil || !strings.Contains(err.Error(), "mystery shape") {
		t.Fatalf("run() error = %v, want one naming the unsupported type", err)
	}
}

func generate(t *testing.T, schema string) string {
	t.Helper()
	out := t.TempDir()
	g := newAPIGenerator(out)
	if err := g.run(schema); err != nil {
		t.Fatalf("generating from %s: %v", schema, err)
	}
	return out
}

// goFiles lists the generated file names in dir.
func goFiles(t *testing.T, dir string) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dir, "*_gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(paths))
	for i, p := range paths {
		names[i] = filepath.Base(p)
	}
	slices.Sort(names)
	return names
}

// compare checks that dir holds exactly the generated files in out, each
// stored under its name plus suffix, with identical content.
func compare(t *testing.T, out, dir, suffix string) {
	t.Helper()
	got := goFiles(t, out)

	paths, err := filepath.Glob(filepath.Join(dir, "*_gen.go"+suffix))
	if err != nil {
		t.Fatal(err)
	}
	var want []string
	for _, p := range paths {
		want = append(want, strings.TrimSuffix(filepath.Base(p), suffix))
	}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("generated files %v, but %s has %v", got, dir, want)
	}

	for _, name := range got {
		g, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		w, err := os.ReadFile(filepath.Join(dir, name+suffix))
		if err != nil {
			t.Fatal(err)
		}
		if line, ok := firstDiff(g, w); ok {
			t.Errorf("%s differs from %s at line %d", name, dir, line)
		}
	}
}

func firstDiff(a, b []byte) (line int, differ bool) {
	if bytes.Equal(a, b) {
		return 0, false
	}
	al, bl := bytes.Split(a, []byte("\n")), bytes.Split(b, []byte("\n"))
	for i := range min(len(al), len(bl)) {
		if !bytes.Equal(al[i], bl[i]) {
			return i + 1, true
		}
	}
	return min(len(al), len(bl)) + 1, true
}
