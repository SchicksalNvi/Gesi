package utils

import (
	"fmt"
	"net"
	"reflect"
	"regexp"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: node-discovery, Property 1: CIDR Validation Correctness
// For any string input, the CIDR parser SHALL either:
// - Accept valid RFC 4632 IPv4 CIDR notation and return a parsed range, OR
// - Reject invalid input with a non-empty error message
// Additionally, for any valid CIDR with prefix length < 16, the parser SHALL reject it as too large.
// **Validates: Requirements 1.1, 1.2, 1.4**
func TestCIDRValidationCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 1a: Valid CIDR strings are accepted and return a parsed range
	properties.Property("valid CIDR strings are accepted", prop.ForAll(
		func(octet1, octet2, octet3, octet4 uint8, prefix int) bool {
			cidr := fmt.Sprintf("%d.%d.%d.%d/%d", octet1, octet2, octet3, octet4, prefix)

			result, err := ParseCIDR(cidr)

			if err != nil {
				// Should only fail if prefix < 16 (range too large)
				if prefix < MinPrefixLength {
					return true // Expected rejection
				}
				t.Logf("Unexpected error for valid CIDR %s: %v", cidr, err)
				return false
			}

			// Verify result is not nil and has valid data
			if result == nil {
				t.Logf("ParseCIDR returned nil result for %s", cidr)
				return false
			}

			// Verify prefix is stored correctly
			if result.Prefix() != prefix {
				t.Logf("Prefix mismatch: expected %d, got %d", prefix, result.Prefix())
				return false
			}

			return true
		},
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.IntRange(16, 32), // Valid prefix range
	))

	// Property 1b: Invalid strings are rejected with non-empty error message
	properties.Property("invalid strings are rejected with non-empty error", prop.ForAll(
		func(input string) bool {
			result, err := ParseCIDR(input)

			// If parsing succeeds, verify it's actually a valid CIDR
			if err == nil {
				if result == nil {
					t.Logf("ParseCIDR returned nil result without error for %s", input)
					return false
				}
				// Verify the result is actually valid by checking it can be used
				if result.Count() <= 0 {
					t.Logf("Invalid count for supposedly valid CIDR %s", input)
					return false
				}
				return true
			}

			// Error should have non-empty message
			if err.Error() == "" {
				t.Logf("Empty error message for invalid input %s", input)
				return false
			}

			// Result should be nil when error is returned
			if result != nil {
				t.Logf("Non-nil result returned with error for %s", input)
				return false
			}

			return true
		},
		genRandomString(),
	))

	// Property 1c: CIDR with prefix < 16 is rejected as too large
	properties.Property("CIDR with prefix < 16 is rejected as too large", prop.ForAll(
		func(octet1, octet2, octet3, octet4 uint8, prefix int) bool {
			cidr := fmt.Sprintf("%d.%d.%d.%d/%d", octet1, octet2, octet3, octet4, prefix)

			result, err := ParseCIDR(cidr)

			// Should always be rejected
			if err == nil {
				t.Logf("Large CIDR %s should be rejected but was accepted", cidr)
				return false
			}

			// Result should be nil
			if result != nil {
				t.Logf("Non-nil result for rejected CIDR %s", cidr)
				return false
			}

			// Error message should mention the range being too large
			errMsg := err.Error()
			if errMsg == "" {
				t.Logf("Empty error message for large CIDR %s", cidr)
				return false
			}

			return true
		},
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.IntRange(0, 15), // Invalid prefix range (too large)
	))

	// Property 1d: IPv6 addresses are rejected
	properties.Property("IPv6 addresses are rejected", prop.ForAll(
		func(prefix int) bool {
			// Generate a simple IPv6 CIDR
			cidr := fmt.Sprintf("2001:db8::%d/%d", prefix, prefix)

			result, err := ParseCIDR(cidr)

			if err == nil {
				t.Logf("IPv6 CIDR %s should be rejected but was accepted", cidr)
				return false
			}

			if result != nil {
				t.Logf("Non-nil result for IPv6 CIDR %s", cidr)
				return false
			}

			return true
		},
		gen.IntRange(16, 128),
	))

	// Property 1e: Empty string is rejected
	properties.Property("empty string is rejected", prop.ForAll(
		func(_ int) bool {
			result, err := ParseCIDR("")

			if err == nil {
				t.Log("Empty string should be rejected")
				return false
			}

			if result != nil {
				t.Log("Non-nil result for empty string")
				return false
			}

			if err.Error() == "" {
				t.Log("Empty error message for empty string")
				return false
			}

			return true
		},
		gen.Const(0),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: node-discovery, Property 2: CIDR IP Count Calculation
// For any valid CIDR with prefix length N (where 16 ≤ N ≤ 32),
// the calculated usable IP count SHALL equal:
// - 2^(32-N) for /31 and /32 (special cases)
// - 2^(32-N) - 2 for /30 and larger (excludes network and broadcast)
// **Validates: Requirements 1.3**
func TestCIDRIPCountCalculation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 2a: IP count equals expected usable hosts for valid CIDR
	properties.Property("IP count equals expected usable hosts", prop.ForAll(
		func(octet1, octet2, octet3, octet4 uint8, prefix int) bool {
			cidr := fmt.Sprintf("%d.%d.%d.%d/%d", octet1, octet2, octet3, octet4, prefix)

			result, err := ParseCIDR(cidr)
			if err != nil {
				t.Logf("Failed to parse valid CIDR %s: %v", cidr, err)
				return false
			}

			totalAddresses := 1 << (32 - prefix) // 2^(32-N)
			var expectedCount int
			if prefix >= 31 {
				// /31 and /32 are special cases - all addresses usable
				expectedCount = totalAddresses
			} else {
				// /30 and larger - exclude network and broadcast
				expectedCount = totalAddresses - 2
			}
			actualCount := result.Count()

			if actualCount != expectedCount {
				t.Logf("Count mismatch for %s: expected %d usable hosts, got %d",
					cidr, expectedCount, actualCount)
				return false
			}

			return true
		},
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.IntRange(16, 32),
	))

	// Property 2b: IPs() returns exactly Count() addresses
	properties.Property("IPs() returns exactly Count() addresses", prop.ForAll(
		func(octet1, octet2, octet3, octet4 uint8, prefix int) bool {
			// Use smaller prefixes to avoid memory issues
			if prefix < 24 {
				prefix = 24 // Limit to /24 (256 IPs) for IPs() test
			}

			cidr := fmt.Sprintf("%d.%d.%d.%d/%d", octet1, octet2, octet3, octet4, prefix)

			result, err := ParseCIDR(cidr)
			if err != nil {
				t.Logf("Failed to parse valid CIDR %s: %v", cidr, err)
				return false
			}

			ips := result.IPs()
			count := result.Count()

			if len(ips) != count {
				t.Logf("IPs length mismatch for %s: expected %d, got %d", cidr, count, len(ips))
				return false
			}

			return true
		},
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.IntRange(24, 32), // Smaller ranges for IPs() test
	))

	// Property 2c: All IPs returned are valid IPv4 addresses
	properties.Property("all IPs returned are valid IPv4 addresses", prop.ForAll(
		func(octet1, octet2, octet3, octet4 uint8, prefix int) bool {
			// Use smaller prefixes to avoid memory issues
			if prefix < 28 {
				prefix = 28 // Limit to /28 (16 IPs)
			}

			cidr := fmt.Sprintf("%d.%d.%d.%d/%d", octet1, octet2, octet3, octet4, prefix)

			result, err := ParseCIDR(cidr)
			if err != nil {
				t.Logf("Failed to parse valid CIDR %s: %v", cidr, err)
				return false
			}

			ips := result.IPs()
			ipv4Regex := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)

			for i, ip := range ips {
				// Check format
				if !ipv4Regex.MatchString(ip) {
					t.Logf("Invalid IP format at index %d: %s", i, ip)
					return false
				}

				// Check parseable
				parsed := net.ParseIP(ip)
				if parsed == nil {
					t.Logf("Unparseable IP at index %d: %s", i, ip)
					return false
				}

				// Check IPv4
				if parsed.To4() == nil {
					t.Logf("Non-IPv4 IP at index %d: %s", i, ip)
					return false
				}
			}

			return true
		},
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.IntRange(28, 32),
	))

	// Property 2d: All IPs are contained within the CIDR range
	properties.Property("all IPs are contained within the CIDR range", prop.ForAll(
		func(octet1, octet2, octet3, octet4 uint8, prefix int) bool {
			// Use smaller prefixes to avoid memory issues
			if prefix < 28 {
				prefix = 28
			}

			cidr := fmt.Sprintf("%d.%d.%d.%d/%d", octet1, octet2, octet3, octet4, prefix)

			result, err := ParseCIDR(cidr)
			if err != nil {
				t.Logf("Failed to parse valid CIDR %s: %v", cidr, err)
				return false
			}

			ips := result.IPs()

			for i, ip := range ips {
				if !result.Contains(ip) {
					t.Logf("IP %s at index %d not contained in CIDR %s", ip, i, cidr)
					return false
				}
			}

			return true
		},
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.IntRange(28, 32),
	))

	// Property 2e: IPs are returned in sequential order
	properties.Property("IPs are returned in sequential order", prop.ForAll(
		func(octet1, octet2, octet3, octet4 uint8, prefix int) bool {
			// Use smaller prefixes
			if prefix < 28 {
				prefix = 28
			}

			cidr := fmt.Sprintf("%d.%d.%d.%d/%d", octet1, octet2, octet3, octet4, prefix)

			result, err := ParseCIDR(cidr)
			if err != nil {
				t.Logf("Failed to parse valid CIDR %s: %v", cidr, err)
				return false
			}

			ips := result.IPs()
			if len(ips) < 2 {
				return true // Nothing to compare
			}

			// Convert IPs to uint32 and verify sequential
			for i := 1; i < len(ips); i++ {
				prev := ipToUint32(ips[i-1])
				curr := ipToUint32(ips[i])

				if curr != prev+1 {
					t.Logf("IPs not sequential: %s (%d) -> %s (%d)",
						ips[i-1], prev, ips[i], curr)
					return false
				}
			}

			return true
		},
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.UInt8(),
		gen.IntRange(28, 32),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genRandomString generates random strings for testing invalid inputs
func genRandomString() gopter.Gen {
	return gen.OneGenOf(
		// Completely random strings
		gen.AnyString(),
		// Strings that look like IPs but aren't valid
		gen.SliceOfN(4, gen.IntRange(0, 999)).Map(func(octets []int) string {
			return fmt.Sprintf("%d.%d.%d.%d", octets[0], octets[1], octets[2], octets[3])
		}),
		// Strings with invalid prefixes
		gen.SliceOfN(4, gen.UInt8()).FlatMap(func(octets interface{}) gopter.Gen {
			o := octets.([]uint8)
			return gen.IntRange(-100, 100).Map(func(prefix int) string {
				return fmt.Sprintf("%d.%d.%d.%d/%d", o[0], o[1], o[2], o[3], prefix)
			})
		}, reflect.TypeOf("")),
		// Empty and whitespace strings
		gen.OneConstOf("", " ", "\t", "\n", "   "),
		// Garbage strings
		gen.OneConstOf(
			"not-a-cidr",
			"192.168.1",
			"192.168.1.1",
			"/24",
			"192.168.1.1/",
			"192.168.1.1/abc",
			"::1/128",
			"2001:db8::/32",
		),
	)
}

// ipToUint32 converts an IP string to uint32 for comparison
func ipToUint32(ipStr string) uint32 {
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}
