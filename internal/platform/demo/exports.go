package demo

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//go:embed exports/*.json
var smartExports embed.FS

// WriteExports fills a private directory below parent (rewritten on every
// start) with example
// host exports for the disk health and WireGuard widgets and returns its two
// subdirectories. A demo never reads the real host (DISKHEALTH_DIR and
// WIREGUARD_DIR are refused), so it shows these instead. parent is the data
// directory: containers usually run with a read-only root and no writable
// /tmp, but the data directory is always writable.
func WriteExports(parent string, now time.Time) (smartDir, wireguardDir string, err error) {
	root := filepath.Join(parent, "demo-exports")
	if err := os.RemoveAll(root); err != nil {
		return "", "", fmt.Errorf("demo exports: %w", err)
	}
	if err := os.Mkdir(root, 0o700); err != nil {
		return "", "", fmt.Errorf("demo exports: %w", err)
	}
	smartDir, wireguardDir = filepath.Join(root, "smart"), filepath.Join(root, "wireguard")
	for _, d := range []string{smartDir, wireguardDir} {
		if err := os.Mkdir(d, 0o700); err != nil {
			return "", "", fmt.Errorf("demo exports: %w", err)
		}
	}
	for _, name := range []string{"sda.json", "nvme0.json"} {
		raw, err := smartExports.ReadFile("exports/" + name)
		if err != nil {
			return "", "", fmt.Errorf("demo exports: %w", err)
		}
		// The reports are fresh, so the widget does not call them stale.
		body := strings.Replace(string(raw), "TIME_T", strconv.FormatInt(now.Add(-10*time.Minute).Unix(), 10), 1)
		if err := os.WriteFile(filepath.Join(smartDir, name), []byte(body), 0o600); err != nil {
			return "", "", fmt.Errorf("demo exports: %w", err)
		}
	}
	if err := os.WriteFile(filepath.Join(wireguardDir, "peers.dump"), []byte(wireguardDump(now)), 0o600); err != nil {
		return "", "", fmt.Errorf("demo exports: %w", err)
	}
	if err := os.WriteFile(filepath.Join(wireguardDir, "peer-names.txt"), []byte(wireguardNames), 0o600); err != nil {
		return "", "", fmt.Errorf("demo exports: %w", err)
	}
	return smartDir, wireguardDir, nil
}

// Made-up public keys in WireGuard's format (44 characters of base64).
var wireguardPeers = []struct {
	key, endpoint, ips string
	handshakeAgo       time.Duration
	rx, tx             int64
}{
	{"aGVsbG9kZW1vcGhvbmVhbm5hMDAwMDAwMDAwMDAwMDA=", "203.0.113.17:51820", "10.13.13.2/32", 40 * time.Second, 812_000_000, 96_000_000},
	{"aGVsbG9kZW1vbGFwdG9wam9uYXMwMDAwMDAwMDAwMDA=", "198.51.100.8:40112", "10.13.13.3/32", 3 * time.Minute, 2_310_000_000, 410_000_000},
	{"aGVsbG9kZW1vdGFibGV0bWFtYTAwMDAwMDAwMDAwMDA=", "(none)", "10.13.13.4/32", 26 * time.Hour, 41_000_000, 6_000_000},
	{"aGVsbG9kZW1vb2ZmaWNlbmFzMDAwMDAwMDAwMDAwMDA=", "192.0.2.44:51820", "10.13.13.5/32,192.168.20.0/24", 70 * time.Second, 15_800_000_000, 9_200_000_000},
}

const wireguardNames = "aGVsbG9kZW1vcGhvbmVhbm5hMDAwMDAwMDAwMDAwMDA= Anna's phone\n" +
	"aGVsbG9kZW1vbGFwdG9wam9uYXMwMDAwMDAwMDAwMDA= Jonas' laptop\n" +
	"aGVsbG9kZW1vdGFibGV0bWFtYTAwMDAwMDAwMDAwMDA= Tablet at grandma's\n" +
	"aGVsbG9kZW1vb2ZmaWNlbmFzMDAwMDAwMDAwMDAwMDA= Office NAS\n"

// wireguardDump has the shape of `wg show all dump`; the interface line
// carries no real key.
func wireguardDump(now time.Time) string {
	var b strings.Builder
	b.WriteString("wg0\t(hidden)\t(hidden)\t51820\toff\n")
	for _, p := range wireguardPeers {
		fmt.Fprintf(&b, "wg0\t%s\t(none)\t%s\t%s\t%d\t%d\t%d\t25\n",
			p.key, p.endpoint, p.ips, now.Add(-p.handshakeAgo).Unix(), p.rx, p.tx)
	}
	return b.String()
}
