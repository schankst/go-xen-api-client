// Package naming converts XenAPI names to Go identifiers.
package naming

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// SnakeToCamel converts a snake_case XenAPI name to an exported CamelCase Go
// name, spelling initialisms in caps, also in the plural:
// "sr_has_no_pbds" -> "SRHasNoPBDs", "guest_os_id" -> "GuestOSID".
func SnakeToCamel(s string) string {
	var b strings.Builder
	for word := range strings.SplitSeq(s, "_") {
		if word == "" {
			continue
		}
		lower := strings.ToLower(word)
		if special, ok := spellings[lower]; ok {
			b.WriteString(special)
			continue
		}
		upper := strings.ToUpper(word)
		if initialisms[upper] {
			b.WriteString(upper)
			continue
		}
		// "HA" + "s" would turn "has" into "HAs".
		if stem, ok := strings.CutSuffix(upper, "S"); ok && initialisms[stem] && stem != "HA" {
			b.WriteString(stem + "s")
			continue
		}
		r, size := utf8.DecodeRuneInString(word)
		b.WriteRune(unicode.ToUpper(r))
		b.WriteString(word[size:])
	}
	return b.String()
}

// spellings are initialisms whose usual spelling isn't all caps.
var spellings = map[string]string{
	"ipv4":  "IPv4",
	"ipv6":  "IPv6",
	"mxgpu": "MxGPU",
	"oauth": "OAuth",
	"qos":   "QoS",
}

// initialisms combines golint's list with the acronyms XenAPI uses in its
// class, field and message names.
var initialisms = map[string]bool{
	// golint
	"ACL": true, "API": true, "ASCII": true, "CPU": true, "CSS": true,
	"DNS": true, "EOF": true, "ETA": true, "GPU": true, "GUID": true,
	"HTML": true, "HTTP": true, "HTTPS": true, "ID": true, "IP": true,
	"JSON": true, "LHS": true, "OS": true, "QPS": true, "RAM": true,
	"RHS": true, "RPC": true, "SLA": true, "SMTP": true, "SQL": true,
	"SSH": true, "TCP": true, "TLS": true, "TTL": true, "UDP": true,
	"UI": true, "UID": true, "UUID": true, "URI": true, "URL": true,
	"UTF8": true, "VM": true, "XML": true, "XMPP": true, "XSRF": true,
	"XSS": true,
	// XenAPI classes
	"DR": true, "LVHD": true, "PBD": true, "PCI": true, "PGPU": true,
	"PIF": true, "PUSB": true, "PVS": true, "SDN": true, "SM": true,
	"SR": true, "USB": true, "VBD": true, "VDI": true, "VGPU": true,
	"VIF": true, "VLAN": true, "VMPP": true, "VMSS": true, "VTPM": true,
	"VUSB": true,
	// other XenAPI terms
	"BIOS": true, "CA": true, "CBT": true, "CIFS": true, "CRL": true,
	"DHCP": true, "GRE": true, "GUI": true, "GVT": true, "HA": true,
	"HVM": true, "IGMP": true, "IO": true, "IQN": true, "ISCSI": true,
	"ISO": true, "LACP": true, "LDAPS": true, "LLDP": true, "MAC": true,
	"MTU": true, "NBD": true, "NFS": true, "NTP": true, "NUMA": true,
	"NVRAM": true, "PSR": true, "PV": true, "PVH": true, "RBAC": true,
	"RDP": true, "RFB": true, "RRD": true, "SHA1": true, "SID": true,
	"SLB": true, "SRIOV": true, "SSL": true, "UEFI": true, "VCPU": true,
	"VHD": true, "VXLAN": true, "WLB": true, "XAPI": true,
}
