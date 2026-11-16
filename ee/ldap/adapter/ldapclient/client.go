package ldapclient

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

// LDAP application tags of the operations used here.
const (
	appBindRequest      = 0x60
	appBindResponse     = 0x61
	appUnbindRequest    = 0x42
	appSearchRequest    = 0x63
	appSearchEntry      = 0x64
	appSearchDone       = 0x65
	appSearchReference  = 0x73
	appExtendedRequest  = 0x77
	appExtendedResponse = 0x78

	filterAnd      = 0xa0
	filterEquality = 0xa3
	simpleAuth     = 0x80
	extendedName   = 0x80

	resultSuccess            = 0
	resultInvalidCredentials = 49

	startTLSOID = "1.3.6.1.4.1.1466.20037"
)

// ErrInvalidCredentials is a failed bind with a wrong DN or password.
var ErrInvalidCredentials = errors.New("ldap: invalid credentials")

// ResultError is any other non-success result.
type ResultError struct {
	Code    int
	Message string
}

func (e *ResultError) Error() string { return fmt.Sprintf("ldap: result %d: %s", e.Code, e.Message) }

// Entry is one search result.
type Entry struct {
	DN         string
	Attributes map[string][]string // keys lowercase
}

// Get returns the first value of an attribute.
func (e Entry) Get(attr string) string {
	if v := e.Attributes[strings.ToLower(attr)]; len(v) > 0 {
		return v[0]
	}
	return ""
}

// Conn is one connection. It is not safe for concurrent use.
type Conn struct {
	conn    net.Conn
	r       *bufio.Reader
	msgID   int
	timeout time.Duration
}

// Dial connects to ldaps://host[:636] or ldap://host[:389]; the latter is
// upgraded with StartTLS before anything else is sent. Plain LDAP is never
// used: a password must not cross the network unencrypted.
func Dial(ctx context.Context, rawURL string, cfg *tls.Config, timeout time.Duration) (*Conn, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("ldap: url: %w", err)
	}
	host := u.Hostname()
	if cfg == nil {
		cfg = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	if cfg.ServerName == "" {
		cfg = cfg.Clone()
		cfg.ServerName = host
	}
	d := &net.Dialer{Timeout: timeout}
	switch u.Scheme {
	case "ldaps":
		addr := net.JoinHostPort(host, portOr(u, "636"))
		c, err := (&tls.Dialer{NetDialer: d, Config: cfg}).DialContext(ctx, "tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("ldap: connect: %w", err)
		}
		return &Conn{conn: c, r: bufio.NewReader(c), timeout: timeout}, nil
	case "ldap":
		c, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, portOr(u, "389")))
		if err != nil {
			return nil, fmt.Errorf("ldap: connect: %w", err)
		}
		conn := &Conn{conn: c, r: bufio.NewReader(c), timeout: timeout}
		if err := conn.startTLS(ctx, cfg); err != nil {
			_ = c.Close()
			return nil, err
		}
		return conn, nil
	default:
		return nil, fmt.Errorf("ldap: unsupported scheme %q (use ldaps:// or ldap:// with StartTLS)", u.Scheme)
	}
}

func portOr(u *url.URL, def string) string {
	if p := u.Port(); p != "" {
		return p
	}
	return def
}

// Close sends an unbind and closes the connection.
func (c *Conn) Close() error {
	c.msgID++
	_ = c.write(seq(tagSequence, integer(tagInteger, c.msgID), tlv(appUnbindRequest, nil)))
	return c.conn.Close()
}

func (c *Conn) write(b []byte) error {
	if c.timeout > 0 {
		_ = c.conn.SetDeadline(time.Now().Add(c.timeout))
	}
	_, err := c.conn.Write(b)
	return err
}

// request sends one operation and returns its message ID.
func (c *Conn) request(op []byte) (int, error) {
	c.msgID++
	return c.msgID, c.write(seq(tagSequence, integer(tagInteger, c.msgID), op))
}

// next reads the next message for id and returns its protocol operation.
func (c *Conn) next(id int) (element, error) {
	for {
		msg, err := readElement(c.r)
		if err != nil {
			return element{}, fmt.Errorf("ldap: read: %w", err)
		}
		if msg.tag != tagSequence || len(msg.children) < 2 {
			return element{}, errMalformed
		}
		got, err := msg.children[0].int()
		if err != nil {
			return element{}, err
		}
		if got == id {
			return msg.children[1], nil
		}
		// A message for another ID (e.g. a notice of disconnection, ID 0) is skipped.
	}
}

// result reads an LDAPResult (resultCode, matchedDN, diagnosticMessage).
func result(op element) error {
	if len(op.children) < 3 {
		return errMalformed
	}
	code, err := op.children[0].int()
	if err != nil {
		return err
	}
	switch code {
	case resultSuccess:
		return nil
	case resultInvalidCredentials:
		return ErrInvalidCredentials
	default:
		return &ResultError{Code: code, Message: string(op.children[2].value)}
	}
}

func (c *Conn) startTLS(ctx context.Context, cfg *tls.Config) error {
	id, err := c.request(seq(appExtendedRequest, octets(extendedName, startTLSOID)))
	if err != nil {
		return fmt.Errorf("ldap: starttls: %w", err)
	}
	op, err := c.next(id)
	if err != nil {
		return err
	}
	if op.tag != appExtendedResponse {
		return errMalformed
	}
	if err := result(op); err != nil {
		return fmt.Errorf("ldap: starttls refused: %w", err)
	}
	tc := tls.Client(c.conn, cfg)
	if c.timeout > 0 {
		_ = c.conn.SetDeadline(time.Now().Add(c.timeout))
	}
	if err := tc.HandshakeContext(ctx); err != nil {
		return fmt.Errorf("ldap: starttls handshake: %w", err)
	}
	c.conn, c.r = tc, bufio.NewReader(tc)
	return nil
}

// Bind authenticates with a DN and password. An empty password is refused
// here: LDAP treats it as an anonymous bind, which would always "succeed".
func (c *Conn) Bind(dn, password string) error {
	if password == "" {
		return ErrInvalidCredentials
	}
	id, err := c.request(seq(appBindRequest, integer(tagInteger, 3), octets(tagOctetString, dn), octets(simpleAuth, password)))
	if err != nil {
		return fmt.Errorf("ldap: bind: %w", err)
	}
	op, err := c.next(id)
	if err != nil {
		return err
	}
	if op.tag != appBindResponse {
		return errMalformed
	}
	return result(op)
}

// Search finds entries below base whose attribute equals value (and whose
// objectClass equals class, when given). The filter is built as BER, so no
// value can change its meaning: there is no string filter to inject into.
func (c *Conn) Search(base, attr, value, class string, attrs []string, sizeLimit int) ([]Entry, error) {
	filter := seq(filterEquality, octets(tagOctetString, attr), octets(tagOctetString, value))
	if class != "" {
		filter = seq(filterAnd, seq(filterEquality, octets(tagOctetString, "objectClass"), octets(tagOctetString, class)), filter)
	}
	var list []byte
	for _, a := range attrs {
		list = append(list, octets(tagOctetString, a)...)
	}
	id, err := c.request(seq(appSearchRequest,
		octets(tagOctetString, base),
		integer(tagEnumerated, 2), // wholeSubtree
		integer(tagEnumerated, 0), // neverDerefAliases
		integer(tagInteger, sizeLimit),
		integer(tagInteger, 10), // seconds
		boolean(false),
		filter,
		tlv(tagSequence, list),
	))
	if err != nil {
		return nil, fmt.Errorf("ldap: search: %w", err)
	}
	var out []Entry
	for {
		op, err := c.next(id)
		if err != nil {
			return nil, err
		}
		switch op.tag {
		case appSearchEntry:
			e, err := entry(op)
			if err != nil {
				return nil, err
			}
			out = append(out, e)
		case appSearchReference:
			// Referrals to other servers are not followed.
		case appSearchDone:
			return out, result(op)
		default:
			return nil, errMalformed
		}
	}
}

func entry(op element) (Entry, error) {
	if len(op.children) < 2 {
		return Entry{}, errMalformed
	}
	e := Entry{DN: string(op.children[0].value), Attributes: map[string][]string{}}
	for _, a := range op.children[1].children {
		if len(a.children) < 2 {
			return Entry{}, errMalformed
		}
		name := strings.ToLower(string(a.children[0].value))
		for _, v := range a.children[1].children {
			e.Attributes[name] = append(e.Attributes[name], string(v.value))
		}
	}
	return e, nil
}
