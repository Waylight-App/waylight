package domain

import (
	"net/netip"
	"testing"
)

var private = []netip.Prefix{
	netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.168.0.0/16"), netip.MustParsePrefix("fc00::/7"),
}

func TestAllowed(t *testing.T) {
	for addr, want := range map[string]bool{
		"192.168.178.20": true, "172.18.0.5": true, "fd00::5": true,
		"8.8.8.8": false, "127.0.0.1": false, "::1": false, "169.254.1.1": false,
	} {
		if got := Allowed(netip.MustParseAddr(addr), private); got != want {
			t.Errorf("%s: %v", addr, got)
		}
	}
	// ALLOWED_CIDRS narrows the private ranges further.
	if Allowed(netip.MustParseAddr("10.1.2.3"), []netip.Prefix{netip.MustParsePrefix("192.168.0.0/16")}) {
		t.Error("outside ALLOWED_CIDRS")
	}
}

func TestPlan(t *testing.T) {
	a := netip.MustParseAddr("192.168.178.20")
	var cands []Candidate
	cands = append(cands, Candidate{Addr: a, Port: 2283}, Candidate{Addr: a, Port: 2283},
		Candidate{Addr: netip.MustParseAddr("1.1.1.1"), Port: 2283}, Candidate{Addr: a, Port: 0})
	for p := range 20 {
		cands = append(cands, Candidate{Addr: netip.MustParseAddr("10.0.0.1"), Port: 8000 + p})
	}
	got := Plan(cands, private)
	if len(got) != MaxProbes || got[0].Port != 2283 || got[1].Port != 8000 {
		t.Fatalf("%+v", got)
	}
	if u := (Candidate{Addr: a, Port: 2283}).URL(); u != "http://192.168.178.20:2283" {
		t.Fatal(u)
	}
	if u := (Candidate{Addr: a, Port: 8123, Host: "homeassistant.local"}).URL(); u != "http://homeassistant.local:8123" {
		t.Fatal(u)
	}
}
