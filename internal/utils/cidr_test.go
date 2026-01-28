package utils

import (
	"testing"
)

func TestParseCIDR_ValidInputs(t *testing.T) {
	tests := []struct {
		name          string
		cidr          string
		expectedCount int // Usable host count (excludes network and broadcast for /30 and larger)
		expectedNet   string
		expectedPfx   int
	}{
		{
			name:          "single host /32",
			cidr:          "192.168.1.1/32",
			expectedCount: 1, // /32 returns all addresses
			expectedNet:   "192.168.1.1",
			expectedPfx:   32,
		},
		{
			name:          "point-to-point /31",
			cidr:          "192.168.1.0/31",
			expectedCount: 2, // /31 returns all addresses (point-to-point)
			expectedNet:   "192.168.1.0",
			expectedPfx:   31,
		},
		{
			name:          "small subnet /30",
			cidr:          "192.168.1.0/30",
			expectedCount: 2, // 4 total - 2 (network + broadcast) = 2 usable
			expectedNet:   "192.168.1.0",
			expectedPfx:   30,
		},
		{
			name:          "typical subnet /24",
			cidr:          "192.168.1.0/24",
			expectedCount: 254, // 256 total - 2 = 254 usable
			expectedNet:   "192.168.1.0",
			expectedPfx:   24,
		},
		{
			name:          "larger subnet /20",
			cidr:          "10.0.0.0/20",
			expectedCount: 4094, // 4096 total - 2 = 4094 usable
			expectedNet:   "10.0.0.0",
			expectedPfx:   20,
		},
		{
			name:          "maximum allowed /16",
			cidr:          "172.16.0.0/16",
			expectedCount: 65534, // 65536 total - 2 = 65534 usable
			expectedNet:   "172.16.0.0",
			expectedPfx:   16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ParseCIDR(tt.cidr)
			if err != nil {
				t.Fatalf("ParseCIDR(%q) returned error: %v", tt.cidr, err)
			}

			if r.Count() != tt.expectedCount {
				t.Errorf("Count() = %d, want %d", r.Count(), tt.expectedCount)
			}

			if r.Network() != tt.expectedNet {
				t.Errorf("Network() = %q, want %q", r.Network(), tt.expectedNet)
			}

			if r.Prefix() != tt.expectedPfx {
				t.Errorf("Prefix() = %d, want %d", r.Prefix(), tt.expectedPfx)
			}
		})
	}
}

func TestParseCIDR_InvalidInputs(t *testing.T) {
	tests := []struct {
		name string
		cidr string
	}{
		{name: "empty string", cidr: ""},
		{name: "no prefix", cidr: "192.168.1.0"},
		{name: "invalid IP", cidr: "999.999.999.999/24"},
		{name: "invalid prefix", cidr: "192.168.1.0/33"},
		{name: "negative prefix", cidr: "192.168.1.0/-1"},
		{name: "garbage input", cidr: "not-a-cidr"},
		{name: "missing IP", cidr: "/24"},
		{name: "IPv6 address", cidr: "2001:db8::/32"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ParseCIDR(tt.cidr)
			if err == nil {
				t.Errorf("ParseCIDR(%q) should return error, got %+v", tt.cidr, r)
			}
		})
	}
}

func TestParseCIDR_RangeTooLarge(t *testing.T) {
	tests := []struct {
		name string
		cidr string
	}{
		{name: "/15 range", cidr: "10.0.0.0/15"},
		{name: "/8 range", cidr: "10.0.0.0/8"},
		{name: "/0 range", cidr: "0.0.0.0/0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ParseCIDR(tt.cidr)
			if err == nil {
				t.Errorf("ParseCIDR(%q) should reject large range, got %+v", tt.cidr, r)
			}

			cidrErr, ok := err.(*CIDRError)
			if !ok {
				t.Errorf("expected *CIDRError, got %T", err)
			}

			if cidrErr.CIDR != tt.cidr {
				t.Errorf("error CIDR = %q, want %q", cidrErr.CIDR, tt.cidr)
			}
		})
	}
}

func TestCIDRRange_IPs(t *testing.T) {
	tests := []struct {
		name     string
		cidr     string
		expected []string
	}{
		{
			name:     "single host",
			cidr:     "192.168.1.100/32",
			expected: []string{"192.168.1.100"},
		},
		{
			name:     "point-to-point /31",
			cidr:     "192.168.1.0/31",
			expected: []string{"192.168.1.0", "192.168.1.1"}, // All addresses for /31
		},
		{
			name:     "/30 subnet",
			cidr:     "192.168.1.0/30",
			expected: []string{"192.168.1.1", "192.168.1.2"}, // Excludes .0 (network) and .3 (broadcast)
		},
		{
			name:     "/29 subnet",
			cidr:     "10.0.0.0/29",
			expected: []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4", "10.0.0.5", "10.0.0.6"}, // Excludes .0 and .7
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := ParseCIDR(tt.cidr)
			if err != nil {
				t.Fatalf("ParseCIDR(%q) returned error: %v", tt.cidr, err)
			}

			ips := r.IPs()
			if len(ips) != len(tt.expected) {
				t.Fatalf("IPs() returned %d addresses, want %d: got %v", len(ips), len(tt.expected), ips)
			}

			for i, ip := range ips {
				if ip != tt.expected[i] {
					t.Errorf("IPs()[%d] = %q, want %q", i, ip, tt.expected[i])
				}
			}
		})
	}
}

func TestCIDRRange_Count(t *testing.T) {
	tests := []struct {
		prefix   int
		expected int // Usable host count
	}{
		{32, 1},     // /32: 1 address (special case)
		{31, 2},     // /31: 2 addresses (point-to-point, special case)
		{30, 2},     // /30: 4 - 2 = 2 usable
		{29, 6},     // /29: 8 - 2 = 6 usable
		{28, 14},    // /28: 16 - 2 = 14 usable
		{24, 254},   // /24: 256 - 2 = 254 usable
		{20, 4094},  // /20: 4096 - 2 = 4094 usable
		{16, 65534}, // /16: 65536 - 2 = 65534 usable
	}

	for _, tt := range tests {
		cidr := "10.0.0.0/" + string(rune('0'+tt.prefix/10)) + string(rune('0'+tt.prefix%10))
		if tt.prefix < 10 {
			cidr = "10.0.0.0/" + string(rune('0'+tt.prefix))
		}

		t.Run(cidr, func(t *testing.T) {
			r, err := ParseCIDR(cidr)
			if err != nil {
				t.Fatalf("ParseCIDR(%q) returned error: %v", cidr, err)
			}

			if r.Count() != tt.expected {
				t.Errorf("Count() = %d, want %d", r.Count(), tt.expected)
			}
		})
	}
}

func TestCIDRRange_Contains(t *testing.T) {
	r, err := ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatalf("ParseCIDR failed: %v", err)
	}

	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.0", true},
		{"192.168.1.1", true},
		{"192.168.1.255", true},
		{"192.168.0.255", false},
		{"192.168.2.0", false},
		{"10.0.0.1", false},
		{"invalid-ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			if r.Contains(tt.ip) != tt.expected {
				t.Errorf("Contains(%q) = %v, want %v", tt.ip, !tt.expected, tt.expected)
			}
		})
	}
}

func TestCIDRRange_CIDR(t *testing.T) {
	cidr := "192.168.1.0/24"
	r, err := ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("ParseCIDR failed: %v", err)
	}

	if r.CIDR() != cidr {
		t.Errorf("CIDR() = %q, want %q", r.CIDR(), cidr)
	}
}

func TestCIDRError_Error(t *testing.T) {
	err := &CIDRError{CIDR: "bad-cidr", Message: "test error"}
	expected := `invalid CIDR "bad-cidr": test error`

	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
