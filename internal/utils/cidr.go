package utils

import (
	"encoding/binary"
	"fmt"
	"net"
)

// CIDRRange represents a parsed CIDR network range
type CIDRRange struct {
	cidr    string
	network *net.IPNet
	ip      net.IP
	prefix  int
}

// MaxCIDRAddresses is the maximum number of addresses allowed (65536 = /16)
const MaxCIDRAddresses = 65536

// MinPrefixLength is the minimum prefix length allowed (16 = /16)
const MinPrefixLength = 16

// CIDRError represents a CIDR parsing error
type CIDRError struct {
	CIDR    string
	Message string
}

func (e *CIDRError) Error() string {
	return fmt.Sprintf("invalid CIDR %q: %s", e.CIDR, e.Message)
}

// ParseCIDR parses and validates a CIDR string.
// Returns a CIDRRange on success, or an error if:
// - The CIDR format is invalid
// - The CIDR is not IPv4
// - The range exceeds 65536 addresses (/16 or larger)
func ParseCIDR(cidr string) (*CIDRRange, error) {
	if cidr == "" {
		return nil, &CIDRError{CIDR: cidr, Message: "empty CIDR string"}
	}

	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, &CIDRError{CIDR: cidr, Message: "not a valid CIDR notation"}
	}

	// Check for IPv4 only
	ipv4 := ip.To4()
	if ipv4 == nil {
		return nil, &CIDRError{CIDR: cidr, Message: "only IPv4 CIDR notation is supported"}
	}

	// Get prefix length
	ones, bits := network.Mask.Size()
	if bits != 32 {
		return nil, &CIDRError{CIDR: cidr, Message: "only IPv4 CIDR notation is supported"}
	}

	// Check range size: reject /16 or larger (prefix < 16)
	if ones < MinPrefixLength {
		return nil, &CIDRError{
			CIDR:    cidr,
			Message: fmt.Sprintf("CIDR range too large: /%d exceeds maximum of %d addresses (minimum prefix is /%d)", ones, MaxCIDRAddresses, MinPrefixLength),
		}
	}

	return &CIDRRange{
		cidr:    cidr,
		network: network,
		ip:      ipv4,
		prefix:  ones,
	}, nil
}

// Count returns the total number of usable IP addresses in the CIDR range.
// For /31 and /32, all addresses are usable (point-to-point links).
// For /30 and larger, network (.0) and broadcast (.255) addresses are excluded.
func (r *CIDRRange) Count() int {
	total := 1 << (32 - r.prefix)
	// /31 and /32 are special cases (point-to-point links)
	if r.prefix >= 31 {
		return total
	}
	// Exclude network and broadcast addresses
	return total - 2
}

// IPs returns all usable IP addresses in the CIDR range as strings.
// Network address (.0) and broadcast address (.255) are excluded for /30 and larger.
// The addresses are returned in sequential order.
func (r *CIDRRange) IPs() []string {
	total := 1 << (32 - r.prefix)

	// Get the network address as uint32
	networkIP := r.network.IP.To4()
	start := binary.BigEndian.Uint32(networkIP)

	// /31 and /32 are special cases - return all addresses
	if r.prefix >= 31 {
		ips := make([]string, 0, total)
		for i := 0; i < total; i++ {
			ipBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(ipBytes, start+uint32(i))
			ips = append(ips, net.IP(ipBytes).String())
		}
		return ips
	}

	// For /30 and larger, skip network (first) and broadcast (last) addresses
	ips := make([]string, 0, total-2)
	for i := 1; i < total-1; i++ {
		ipBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(ipBytes, start+uint32(i))
		ips = append(ips, net.IP(ipBytes).String())
	}

	return ips
}

// CIDR returns the original CIDR string.
func (r *CIDRRange) CIDR() string {
	return r.cidr
}

// Network returns the network address.
func (r *CIDRRange) Network() string {
	return r.network.IP.String()
}

// Prefix returns the prefix length.
func (r *CIDRRange) Prefix() int {
	return r.prefix
}

// Contains checks if the given IP address is within the CIDR range.
func (r *CIDRRange) Contains(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return r.network.Contains(parsedIP)
}
