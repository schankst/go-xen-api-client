//go:build ignore
// +build ignore

// patch_enums.go re-applies the enum-tolerance patch to convert_gen.go
// after a regeneration (`go run xenapi.go`) has overwritten it with
// upstream-style strict enum parsing.
//
// Run with: go run patch_enums.go
package main

import (
	"fmt"
	"os"
	"regexp"
)

var reUnpatched = regexp.MustCompile(
	`err = fmt\.Errorf\("Unable to parse XenAPI response: got value %q for enum %s at %s, but this is not any of the known values", strValue, "([A-Za-z0-9]+)", context\)`,
)

func main() {
	const path = "convert_gen.go"

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reading", path, ":", err)
		os.Exit(1)
	}

	matches := reUnpatched.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		fmt.Println("no unpatched enum converters found (already patched, or file structure changed)")
		return
	}

	patched := reUnpatched.ReplaceAll(data, []byte(`value = $1(strValue) // unknown enum value from a newer XAPI version; passed through as-is`))

	if err := os.WriteFile(path, patched, 0644); err != nil {
		fmt.Fprintln(os.Stderr, "writing", path, ":", err)
		os.Exit(1)
	}

	fmt.Printf("patched %d enum converter(s) in %s\n", len(matches), path)
}
