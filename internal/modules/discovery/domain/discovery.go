package domain

import (
	"net/netip"
	"strconv"
)

// MaxProbes bounds the requests one search sends, so opening a setup never
// turns into a scan of the network.
const MaxProbes = 12

// Candidate is an address that may run the service, and where it came from.
type Candidate struct {
	Addr netip.Addr
	Port int
	// Host is the name it was found by, kept in the URL so certificates and
	// virtual hosts match; empty to use the address.
	Host string
	Hint string
}

// URL is the base URL a probe and the setup use.
func (c Candidate) URL() string {
	host := c.Host
	if host == "" {
		host = c.Addr.String()
		if c.Addr.Is6() {
			host = "[" + host + "]"
		}
	}
	return "http://" + host + ":" + strconv.Itoa(c.Port)
}

// Found is a confirmed service.
type Found struct {
	URL, Hint string
}

// Allowed reports whether a candidate may be probed: only private addresses
// (RFC 1918, ULA) that the instance also allows for integrations.
func Allowed(addr netip.Addr, allowed []netip.Prefix) bool {
	addr = addr.Unmap()
	if !addr.IsValid() || !addr.IsPrivate() || addr.IsLoopback() {
		return false
	}
	for _, p := range allowed {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

// Plan removes duplicates and addresses that may not be probed, keeps the
// order (Docker first, then host names, then the gateway) and stops at
// MaxProbes.
func Plan(cands []Candidate, allowed []netip.Prefix) []Candidate {
	seen := map[string]bool{}
	var out []Candidate
	for _, c := range cands {
		key := c.Addr.String() + ":" + strconv.Itoa(c.Port)
		if c.Port <= 0 || c.Port > 65535 || seen[key] || !Allowed(c.Addr, allowed) {
			continue
		}
		seen[key] = true
		out = append(out, c)
		if len(out) == MaxProbes {
			break
		}
	}
	return out
}
