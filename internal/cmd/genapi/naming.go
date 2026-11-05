package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// snakeToCamel converts a snake_case name to an exported CamelCase Go name,
// spelling common initialisms in all caps: "guest_os_id" -> "GuestOSID".
func snakeToCamel(s string) string {
	var b strings.Builder
	for word := range strings.SplitSeq(s, "_") {
		if word == "oauth" {
			b.WriteString("OAuth")
			continue
		}
		if upper := strings.ToUpper(word); commonInitialisms[upper] {
			b.WriteString(upper)
			continue
		}
		if word == "" {
			continue
		}
		r, size := utf8.DecodeRuneInString(word)
		b.WriteRune(unicode.ToUpper(r))
		b.WriteString(word[size:])
	}
	return b.String()
}

// commonInitialisms is golint's initialism list, which the generated names
// have always followed. Changing it renames exported identifiers.
var commonInitialisms = map[string]bool{
	"ACL": true, "API": true, "ASCII": true, "CPU": true, "CSS": true,
	"DNS": true, "EOF": true, "ETA": true, "GPU": true, "GUID": true,
	"HTML": true, "HTTP": true, "HTTPS": true, "ID": true, "IP": true,
	"JSON": true, "LHS": true, "OS": true, "QPS": true, "RAM": true,
	"RHS": true, "RPC": true, "SLA": true, "SMTP": true, "SQL": true,
	"SSH": true, "TCP": true, "TLS": true, "TTL": true, "UDP": true,
	"UI": true, "UID": true, "UUID": true, "URI": true, "URL": true,
	"UTF8": true, "VM": true, "XML": true, "XMPP": true, "XSRF": true,
	"XSS": true,
}
