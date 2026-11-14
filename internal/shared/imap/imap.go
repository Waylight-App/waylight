// Package imap is a minimal, read-only IMAP4rev1 client over implicit TLS
// for integrations that look into a mailbox (parcel discovery #166, the mail
// widget #283). Folders are opened with EXAMINE, so the server refuses any
// change; callers fetch with BODY.PEEK so not even \Seen changes.
//
// Ring: Interface Adapters (helpers for gateways).
package imap

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Limits of one connection.
const (
	Timeout         = 60 * time.Second
	MaxLiteralBytes = 320 << 10
	maxLineBytes    = 64 << 10
)

// DialFunc opens a TCP connection; production passes the guarded dialer of
// the integration clients, so link-local and metadata addresses are refused.
type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)

// Account is a mailbox login.
type Account struct {
	Host     string
	Port     int
	Username string
	Password string
}

// Client opens connections.
type Client struct {
	Dial DialFunc
	// Plain skips TLS; tests only.
	Plain bool
}

// ErrLogin is returned when the server refuses the login.
var ErrLogin = errors.New("login failed: check user name and app password")

// Conn is an open, logged-in connection.
type Conn struct {
	nc  net.Conn
	r   *bufio.Reader
	tag int
}

// Response is one server response; literals are cut out of Text, At marks
// the position in Text where each literal was.
type Response struct {
	Text string
	Lits [][]byte
	At   []int
}

var (
	literalTail = regexp.MustCompile(`\{(\d+)\}$`)
	uidValidity = regexp.MustCompile(`\[UIDVALIDITY (\d+)\]`)
	uidNextRe   = regexp.MustCompile(`\[UIDNEXT (\d+)\]`)
)

// Open connects, checks the greeting and logs in.
func (cl Client) Open(ctx context.Context, a Account) (*Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	nc, err := cl.Dial(ctx, "tcp", net.JoinHostPort(a.Host, strconv.Itoa(a.Port)))
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if !cl.Plain {
		tc := tls.Client(nc, &tls.Config{ServerName: a.Host, MinVersion: tls.VersionTLS12})
		if err := tc.HandshakeContext(ctx); err != nil {
			_ = nc.Close()
			return nil, fmt.Errorf("TLS (only IMAP over TLS, usually port 993, is supported): %w", err)
		}
		nc = tc
	}
	_ = nc.SetDeadline(time.Now().Add(Timeout))
	c := &Conn{nc: nc, r: bufio.NewReaderSize(nc, 16<<10)}
	greeting, err := c.read()
	if err != nil || !strings.HasPrefix(greeting.Text, "* OK") {
		c.Close()
		return nil, fmt.Errorf("not an IMAP server: %q %w", Clip(greeting.Text), err)
	}
	if _, err := c.Cmd("LOGIN " + Quote(a.Username) + " " + Quote(a.Password)); err != nil {
		c.Close()
		return nil, ErrLogin
	}
	return c, nil
}

// Close logs out and closes the connection.
func (c *Conn) Close() {
	_, _ = c.Cmd("LOGOUT")
	_ = c.nc.Close()
}

// Examine opens a folder read-only and returns UIDVALIDITY and UIDNEXT.
func (c *Conn) Examine(folder string) (validity, next uint32, err error) {
	lines, err := c.Cmd("EXAMINE " + Quote(EncodeMailbox(folder)))
	if err != nil {
		return 0, 0, fmt.Errorf("folder %q: %w", folder, err)
	}
	for _, l := range lines {
		if m := uidValidity.FindStringSubmatch(l.Text); m != nil {
			v, _ := strconv.ParseUint(m[1], 10, 32)
			validity = uint32(v)
		}
		if m := uidNextRe.FindStringSubmatch(l.Text); m != nil {
			v, _ := strconv.ParseUint(m[1], 10, 32)
			next = uint32(v)
		}
	}
	return validity, next, nil
}

// SearchUIDs runs "UID SEARCH <criteria>" and returns the UIDs, sorted.
func (c *Conn) SearchUIDs(criteria string) ([]uint32, error) {
	lines, err := c.Cmd("UID SEARCH " + criteria)
	if err != nil {
		return nil, err
	}
	var uids []uint32
	for _, l := range lines {
		if rest, ok := strings.CutPrefix(l.Text, "* SEARCH"); ok {
			for _, f := range strings.Fields(rest) {
				if u, err := strconv.ParseUint(f, 10, 32); err == nil {
					uids = append(uids, uint32(u))
				}
			}
		}
	}
	slices.Sort(uids)
	return uids, nil
}

// Cmd sends a command and returns the untagged responses.
func (c *Conn) Cmd(command string) ([]Response, error) {
	c.tag++
	tag := "T" + strconv.Itoa(c.tag)
	if _, err := io.WriteString(c.nc, tag+" "+command+"\r\n"); err != nil {
		return nil, fmt.Errorf("send: %w", err)
	}
	var out []Response
	for {
		r, err := c.read()
		if err != nil {
			return nil, err
		}
		if rest, ok := strings.CutPrefix(r.Text, tag+" "); ok {
			if strings.HasPrefix(rest, "OK") {
				return out, nil
			}
			return nil, fmt.Errorf("server answered: %s", Clip(rest))
		}
		out = append(out, r)
	}
}

// read reads one response including its literals.
func (c *Conn) read() (Response, error) {
	var r Response
	for {
		line, err := c.line()
		if err != nil {
			return r, err
		}
		m := literalTail.FindStringSubmatchIndex(line)
		if m == nil {
			r.Text += line
			return r, nil
		}
		n, _ := strconv.Atoi(line[m[2]:m[3]])
		if n > MaxLiteralBytes {
			return r, fmt.Errorf("message part too large (%d bytes)", n)
		}
		r.Text += line[:m[0]]
		buf := make([]byte, n)
		if _, err := io.ReadFull(c.r, buf); err != nil {
			return r, fmt.Errorf("read literal: %w", err)
		}
		r.Lits = append(r.Lits, buf)
		r.At = append(r.At, len(r.Text))
	}
}

func (c *Conn) line() (string, error) {
	var b strings.Builder
	for {
		chunk, err := c.r.ReadSlice('\n')
		b.Write(chunk)
		if b.Len() > maxLineBytes {
			return "", errors.New("response line too long")
		}
		if err == nil {
			return strings.TrimRight(b.String(), "\r\n"), nil
		}
		if !errors.Is(err, bufio.ErrBufferFull) {
			return "", fmt.Errorf("read: %w", err)
		}
	}
}

// Section returns the part of a FETCH response that follows the item
// matched by re (sent as literal, quoted string or NIL).
func (r Response) Section(re *regexp.Regexp) map[string][]byte {
	out := map[string][]byte{}
	for _, m := range re.FindAllStringSubmatchIndex(r.Text, -1) {
		name, pos := r.Text[m[2]:m[3]], m[1]
		if i := slices.Index(r.At, pos); i >= 0 {
			out[name] = r.Lits[i]
			continue
		}
		if pos < len(r.Text) && r.Text[pos] == '"' {
			out[name] = []byte(unquote(r.Text[pos:]))
		}
	}
	return out
}

// Quote makes an IMAP quoted string.
func Quote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}

// unquote reads a quoted string at the start of s.
func unquote(s string) string {
	var b strings.Builder
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			if i+1 < len(s) {
				i++
				b.WriteByte(s[i])
			}
		case '"':
			return b.String()
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// Clip shortens server text for error messages.
func Clip(s string) string {
	if len(s) > 120 {
		return s[:120] + "…"
	}
	return s
}

// Words decodes RFC 2047 encoded header words; Latin-1 and Windows-1252
// are read as such, UTF-8 stays.
var Words = &mime.WordDecoder{CharsetReader: CharsetReader}

// ToUTF8 converts Latin-1/Windows-1252 text; UTF-8 and ASCII stay.
func ToUTF8(raw []byte, charset string) string {
	cs := strings.ToLower(charset)
	if utf8.Valid(raw) && !strings.HasPrefix(cs, "iso-8859") && !strings.HasPrefix(cs, "windows-125") {
		return string(raw)
	}
	r := make([]rune, len(raw))
	for i, c := range raw {
		r[i] = rune(c)
	}
	return string(r)
}

// CharsetReader is the mime.WordDecoder hook for ToUTF8.
func CharsetReader(charset string, input io.Reader) (io.Reader, error) {
	raw, err := io.ReadAll(io.LimitReader(input, 4096))
	if err != nil {
		return nil, err
	}
	return strings.NewReader(ToUTF8(raw, charset)), nil
}
