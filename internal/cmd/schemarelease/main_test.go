package main

import "testing"

func TestLatestRelease(t *testing.T) {
	tests := []struct {
		name, schema, want string
		ok                 bool
	}{
		{"none", `{"release": "rio"}`, "", false},
		{"single", `{"release": "1.2.3"}`, "1.2.3", true},
		{"numeric not lexical", `{"release": "9.0.0"} {"release": "10.0.0"}`, "10.0.0", true},
		{"minor and patch", `{"release": "1.10.0"} {"release": "1.9.9"} {"release": "1.10.1"}`, "1.10.1", true},
		{"next is newer than same version", `{"release": "2.0.0-next"} {"release": "2.0.0"}`, "2.0.0-next", true},
		{"next is older than next patch", `{"release": "2.0.0-next"} {"release": "2.0.1"}`, "2.0.1", true},
		{"whitespace", `{"release" :  "3.1.4"}`, "3.1.4", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := latestRelease([]byte(tt.schema))
			if got != tt.want || ok != tt.ok {
				t.Errorf("latestRelease() = %q, %v; want %q, %v", got, ok, tt.want, tt.ok)
			}
		})
	}
}
