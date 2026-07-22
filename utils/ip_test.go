package utils

import (
	"testing"
)

func TestIPToInt(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want uint32
	}{
		{"zero", "0.0.0.0", 0},
		{"max", "255.255.255.255", 0xFFFFFFFF},
		{"common", "192.168.1.1", 0xC0A80101},
		{"loopback", "127.0.0.1", 0x7F000001},
		{"all_ones_last", "10.0.0.255", 0x0A0000FF},
		{"invalid_string", "not-an-ip", 0},
		{"empty", "", 0},
		{"ipv6", "::1", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IPToInt(tt.ip)
			if got != tt.want {
				t.Fatalf("IPToInt(%q) = 0x%08X, want 0x%08X", tt.ip, got, tt.want)
			}
		})
	}
}

func TestIPToString(t *testing.T) {
	tests := []struct {
		name  string
		ipInt uint32
		want  string
	}{
		{"zero", 0, "0.0.0.0"},
		{"max", 0xFFFFFFFF, "255.255.255.255"},
		{"common", 0xC0A80101, "192.168.1.1"},
		{"loopback", 0x7F000001, "127.0.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IPToString(tt.ipInt)
			if got != tt.want {
				t.Fatalf("IPToString(0x%08X) = %q, want %q", tt.ipInt, got, tt.want)
			}
		})
	}
}

func TestIPRoundtrip(t *testing.T) {
	ips := []string{
		"0.0.0.0",
		"255.255.255.255",
		"192.168.1.1",
		"127.0.0.1",
		"10.0.0.1",
		"172.16.254.1",
	}

	for _, ip := range ips {
		t.Run(ip, func(t *testing.T) {
			intVal := IPToInt(ip)
			back := IPToString(intVal)
			if back != ip {
				t.Fatalf("roundtrip failed: %s → 0x%08X → %s", ip, intVal, back)
			}
		})
	}
}
