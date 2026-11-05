// Package naming converts XenAPI names to Go identifiers.
package naming

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// SnakeToCamel converts a snake_case name to an exported CamelCase Go name,
// spelling common initialisms in all caps: "guest_os_id" -> "GuestOSID".
func SnakeToCamel(s string) string {
	return convert(s, commonInitialisms, false)
}

// SnakeToCamelXenAPI is SnakeToCamel that also spells XenAPI's class
// acronyms in caps, including their plural: "sr_has_no_pbds" ->
// "SRHasNoPBDs".
func SnakeToCamelXenAPI(s string) string {
	return convert(s, xenAPIInitialisms, true)
}

func convert(s string, initialisms map[string]bool, plurals bool) string {
	var b strings.Builder
	for word := range strings.SplitSeq(s, "_") {
		if word == "oauth" {
			b.WriteString("OAuth")
			continue
		}
		upper := strings.ToUpper(word)
		if initialisms[upper] {
			b.WriteString(upper)
			continue
		}
		// "HA" + "s" would turn "has" into "HAs".
		if stem, ok := strings.CutSuffix(upper, "S"); plurals && ok && initialisms[stem] && stem != "HA" {
			b.WriteString(stem + "s")
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

// xenAPIInitialisms adds the acronyms XenAPI spells in caps in its class
// names, plus VCPU and HA, to commonInitialisms.
var xenAPIInitialisms = func() map[string]bool {
	m := map[string]bool{
		"DR": true, "HA": true, "LVHD": true, "PBD": true, "PCI": true,
		"PGPU": true, "PIF": true, "PUSB": true, "PVS": true, "SDN": true,
		"SM": true, "SR": true, "USB": true, "VBD": true, "VCPU": true,
		"VDI": true, "VGPU": true, "VIF": true, "VLAN": true, "VMPP": true,
		"VMSS": true, "VTPM": true, "VUSB": true,
	}
	for k := range commonInitialisms {
		m[k] = true
	}
	return m
}()
