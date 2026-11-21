// Package gateway reads the default gateway, resolves host names and
// probes candidates over HTTP.
//
// Ring: Interface Adapters.
package gateway

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"
)

// maxProbeBytes bounds what a probe reads; markers sit near the start.
const maxProbeBytes = 64 << 10

// routeFile is the kernel's IPv4 routing table.
const routeFile = "/proc/net/route"

// DefaultGateway reads the IPv4 default route from /proc/net/route; false
// on systems without it.
func DefaultGateway() (netip.Addr, bool) {
	f, err := os.Open(routeFile)
	if err != nil {
		return netip.Addr{}, false
	}
	defer f.Close()
	return parseRoutes(f)
}

// parseRoutes finds the default route (destination 0) in the route table;
// addresses are little-endian hex.
func parseRoutes(r io.Reader) (netip.Addr, bool) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) < 3 || f[1] != "00000000" {
			continue
		}
		raw, err := hex.DecodeString(f[2])
		if err != nil || len(raw) != 4 {
			continue
		}
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], binary.LittleEndian.Uint32(raw))
		if a := netip.AddrFrom4(b); !a.IsUnspecified() {
			return a, true
		}
	}
	return netip.Addr{}, false
}

// Resolve looks a host name up with the system resolver.
func Resolve(ctx context.Context, host string) ([]netip.Addr, error) {
	return net.DefaultResolver.LookupNetIP(ctx, "ip", host)
}

// Prober requests the probe path of a candidate without credentials.
type Prober struct{ client *http.Client }

// NewProber returns a prober; client must follow the integration URL rules.
func NewProber(client *http.Client) *Prober { return &Prober{client: client} }

// Probe reports whether baseURL+path answers with a body that contains the
// marker; any status counts, since some services answer 401 to strangers.
func (p *Prober) Probe(ctx context.Context, baseURL, path, contains string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return false
	}
	res, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, maxProbeBytes))
	return err == nil && strings.Contains(string(body), contains)
}
