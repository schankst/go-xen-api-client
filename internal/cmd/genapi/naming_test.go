package main

import "testing"

func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"vm", "VM"},
		{"host_metrics", "HostMetrics"},
		{"PIF_metrics", "PIFMetrics"},
		{"guest_os_id", "GuestOSID"},
		{"vm_uuid", "VMUUID"},
		{"oauth_token", "OAuthToken"},
		{"network_sriov", "NetworkSriov"},
		{"double__underscore", "DoubleUnderscore"},
		{"trailing_", "Trailing"},
	}
	for _, tt := range tests {
		if got := snakeToCamel(tt.in); got != tt.want {
			t.Errorf("snakeToCamel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
