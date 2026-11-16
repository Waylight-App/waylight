package ldapclient

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net"
	"testing"
	"time"
)

// testTLS returns a server config with a fresh certificate for 127.0.0.1 and
// a client config that trusts it.
func testTLS(t *testing.T) (*tls.Config, *tls.Config) {
	t.Helper()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "ldap.test"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, KeyUsage: x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, _ := x509.ParseCertificate(der)
	pool := x509.NewCertPool()
	pool.AddCert(cert)
	return &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}}},
		&tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
}

// fakeDirectory is a minimal LDAP server for the tests.
type fakeDirectory struct {
	passwords map[string]string // DN -> password
	entries   []Entry
	tls       *tls.Config
}

func (f *fakeDirectory) serve(t *testing.T, ldaps bool) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			if ldaps {
				c = tls.Server(c, f.tls)
			}
			go f.handle(c)
		}
	}()
	return ln.Addr().String()
}

func ldapResult(app byte, code int) []byte {
	return seq(app, integer(tagEnumerated, code), octets(tagOctetString, ""), octets(tagOctetString, ""))
}

func (f *fakeDirectory) handle(c net.Conn) {
	defer c.Close()
	r := bufio.NewReader(c)
	reply := func(id int, op []byte) { _, _ = c.Write(seq(tagSequence, integer(tagInteger, id), op)) }
	for {
		msg, err := readElement(r)
		if err != nil {
			return
		}
		id, _ := msg.children[0].int()
		op := msg.children[1]
		switch op.tag {
		case appExtendedRequest:
			reply(id, ldapResult(appExtendedResponse, resultSuccess))
			tc := tls.Server(c, f.tls)
			if tc.Handshake() != nil {
				return
			}
			c, r = tc, bufio.NewReader(tc)
		case appBindRequest:
			code := resultInvalidCredentials
			if want, ok := f.passwords[string(op.children[1].value)]; ok && want == string(op.children[2].value) {
				code = resultSuccess
			}
			reply(id, ldapResult(appBindResponse, code))
		case appSearchRequest:
			filter := op.children[6]
			eq := filter
			if filter.tag == filterAnd {
				eq = filter.children[1]
			}
			attr, value := string(eq.children[0].value), string(eq.children[1].value)
			for _, e := range f.entries {
				if e.Get(attr) != value {
					continue
				}
				var attrs []byte
				for name, vals := range e.Attributes {
					var set []byte
					for _, v := range vals {
						set = append(set, octets(tagOctetString, v)...)
					}
					attrs = append(attrs, seq(tagSequence, octets(tagOctetString, name), tlv(tagSet, set))...)
				}
				reply(id, seq(appSearchEntry, octets(tagOctetString, e.DN), tlv(tagSequence, attrs)))
			}
			reply(id, ldapResult(appSearchDone, resultSuccess))
		case appUnbindRequest:
			return
		}
	}
}

func directory(t *testing.T) (*fakeDirectory, *tls.Config) {
	server, client := testTLS(t)
	return &fakeDirectory{
		tls:       server,
		passwords: map[string]string{"cn=svc,dc=example,dc=org": "svc-pw", "uid=anna,ou=people,dc=example,dc=org": "anna-pw"},
		entries: []Entry{{DN: "uid=anna,ou=people,dc=example,dc=org", Attributes: map[string][]string{
			"uid": {"anna"}, "mail": {"anna@example.org"}, "memberof": {"cn=Office,ou=groups,dc=example,dc=org", "cn=Sales,ou=groups,dc=example,dc=org"},
		}}},
	}, client
}

func TestLDAPSBindAndSearch(t *testing.T) {
	dir, cfg := directory(t)
	addr := dir.serve(t, true)
	c, err := Dial(context.Background(), "ldaps://"+addr, cfg, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.Bind("cn=svc,dc=example,dc=org", "svc-pw"); err != nil {
		t.Fatalf("service bind: %v", err)
	}
	entries, err := c.Search("dc=example,dc=org", "uid", "anna", "", []string{"mail", "memberOf"}, 2)
	if err != nil || len(entries) != 1 {
		t.Fatalf("search: %v %v", entries, err)
	}
	if entries[0].Get("mail") != "anna@example.org" || len(entries[0].Attributes["memberof"]) != 2 {
		t.Fatalf("entry: %+v", entries[0])
	}
	if err := c.Bind(entries[0].DN, "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password: %v", err)
	}
	if err := c.Bind(entries[0].DN, ""); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("empty password must never bind: %v", err)
	}
	if err := c.Bind(entries[0].DN, "anna-pw"); err != nil {
		t.Fatalf("user bind: %v", err)
	}
}

func TestStartTLS(t *testing.T) {
	dir, cfg := directory(t)
	addr := dir.serve(t, false)
	c, err := Dial(context.Background(), "ldap://"+addr, cfg, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, ok := c.conn.(*tls.Conn); !ok {
		t.Fatal("connection not upgraded to TLS")
	}
	if err := c.Bind("cn=svc,dc=example,dc=org", "svc-pw"); err != nil {
		t.Fatalf("bind after StartTLS: %v", err)
	}
}

func TestUntrustedCertificateIsRefused(t *testing.T) {
	dir, _ := directory(t)
	addr := dir.serve(t, true)
	if _, err := Dial(context.Background(), "ldaps://"+addr, &tls.Config{MinVersion: tls.VersionTLS12}, 5*time.Second); err == nil {
		t.Fatal("connected to an untrusted server")
	}
	if _, err := Dial(context.Background(), "http://"+addr, nil, time.Second); err == nil {
		t.Fatal("accepted a non-LDAP scheme")
	}
}

func TestBERRoundTrip(t *testing.T) {
	long := make([]byte, 300)
	b := seq(tagSequence, integer(tagInteger, 70000), octets(tagOctetString, string(long)), boolean(true))
	e, err := readElement(bufio.NewReader(bytesReaderOf(b)))
	if err != nil || len(e.children) != 3 {
		t.Fatalf("%v %v", e, err)
	}
	if n, _ := e.children[0].int(); n != 70000 || len(e.children[1].value) != 300 {
		t.Fatalf("decoded %d, %d bytes", n, len(e.children[1].value))
	}
	if _, err := readElement(bufio.NewReader(bytesReaderOf([]byte{0x30, 0x80}))); err == nil {
		t.Fatal("indefinite length accepted")
	}
}

func bytesReaderOf(b []byte) *bytes.Reader { return bytes.NewReader(b) }
