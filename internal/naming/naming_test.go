package naming

import "testing"

func TestSnakeToCamelBasics(t *testing.T) {
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
		{"network_sriov", "NetworkSRIOV"},
		{"double__underscore", "DoubleUnderscore"},
		{"trailing_", "Trailing"},
	}
	for _, tt := range tests {
		if got := SnakeToCamel(tt.in); got != tt.want {
			t.Errorf("SnakeToCamel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"sr_has_no_pbds", "SRHasNoPBDs"},
		{"vm_toomany_vcpus", "VMToomanyVCPUs"},
		{"pif_is_management_iface", "PIFIsManagementIface"},
		{"ha_not_enabled", "HANotEnabled"},
		{"vms_failed_to_cooperate", "VMsFailedToCooperate"},
		{"uuid_invalid", "UUIDInvalid"},
		{"host_has_resident_vms", "HostHasResidentVMs"},
		{"cannot_add_tunnel_to_sriov_logical", "CannotAddTunnelToSRIOVLogical"},
		{"status", "Status"},
		{"assigned_ips", "AssignedIPs"},
		{"ipv4_configuration_mode", "IPv4ConfigurationMode"},
		{"qos_algorithm_type", "QoSAlgorithmType"},
		{"pv_drivers_detected", "PVDriversDetected"},
		{"has_vendor_device", "HasVendorDevice"},
	}
	for _, tt := range tests {
		if got := SnakeToCamel(tt.in); got != tt.want {
			t.Errorf("SnakeToCamel(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
