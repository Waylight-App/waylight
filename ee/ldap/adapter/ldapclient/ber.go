// Package ldapclient is a small LDAP v3 client (RFC 4511) with the standard
// library: simple bind, subtree search with an equality filter, StartTLS.
// That is all directory sign-in needs (#396); like the OpenID Connect code of
// ADR 0013, it avoids a dependency for a narrow, well-specified protocol.
//
// Ring: Frameworks and drivers (used by the gateway of ee/ldap).
package ldapclient

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

// BER tags used by LDAP (RFC 4511, appendix B, and X.690).
const (
	tagBoolean     = 0x01
	tagInteger     = 0x02
	tagOctetString = 0x04
	tagEnumerated  = 0x0a
	tagSequence    = 0x30
	tagSet         = 0x31
)

// maxElement bounds one BER element; directory answers for one user are small.
const maxElement = 4 << 20

var errMalformed = errors.New("ldap: malformed BER data")

// element is one decoded BER TLV.
type element struct {
	tag      byte
	value    []byte
	children []element // set for constructed elements
}

func (e element) constructed() bool { return e.tag&0x20 != 0 }

// appendLength encodes a definite length.
func appendLength(b []byte, n int) []byte {
	if n < 0x80 {
		return append(b, byte(n)) //nolint:gosec // below 0x80, fits one byte
	}
	var tmp []byte
	for n > 0 {
		tmp = append([]byte{byte(n & 0xff)}, tmp...)
		n >>= 8
	}
	return append(append(b, 0x80|byte(len(tmp))), tmp...) //nolint:gosec // at most 8 length bytes
}

// tlv encodes one element from its tag and content.
func tlv(tag byte, content []byte) []byte {
	return append(appendLength([]byte{tag}, len(content)), content...)
}

func seq(tag byte, parts ...[]byte) []byte {
	var content []byte
	for _, p := range parts {
		content = append(content, p...)
	}
	return tlv(tag, content)
}

func octets(tag byte, s string) []byte { return tlv(tag, []byte(s)) }

func integer(tag byte, n int) []byte {
	// Minimal two's complement; LDAP only needs small non-negative numbers.
	if n == 0 {
		return tlv(tag, []byte{0})
	}
	var b []byte
	for v := n; v > 0; v >>= 8 {
		b = append([]byte{byte(v)}, b...)
	}
	if b[0]&0x80 != 0 {
		b = append([]byte{0}, b...)
	}
	return tlv(tag, b)
}

func boolean(v bool) []byte {
	if v {
		return tlv(tagBoolean, []byte{0xff})
	}
	return tlv(tagBoolean, []byte{0})
}

// readElement reads one TLV from r, with its children when constructed.
func readElement(r *bufio.Reader) (element, error) {
	tag, err := r.ReadByte()
	if err != nil {
		return element{}, err
	}
	if tag&0x1f == 0x1f {
		return element{}, errMalformed // multi-byte tags do not occur in LDAP
	}
	first, err := r.ReadByte()
	if err != nil {
		return element{}, err
	}
	n := int(first)
	if first&0x80 != 0 {
		count := int(first & 0x7f)
		if count == 0 || count > 4 {
			return element{}, errMalformed // indefinite or absurd lengths
		}
		n = 0
		for range count {
			b, err := r.ReadByte()
			if err != nil {
				return element{}, err
			}
			n = n<<8 | int(b)
		}
	}
	if n > maxElement {
		return element{}, fmt.Errorf("ldap: element of %d bytes is too large", n)
	}
	value := make([]byte, n)
	if _, err := io.ReadFull(r, value); err != nil {
		return element{}, err
	}
	e := element{tag: tag, value: value}
	if e.constructed() {
		e.children, err = parseChildren(value)
		if err != nil {
			return element{}, err
		}
	}
	return e, nil
}

func parseChildren(b []byte) ([]element, error) {
	var out []element
	r := bufio.NewReader(bytes.NewReader(b))
	for {
		e, err := readElement(r)
		if errors.Is(err, io.EOF) {
			return out, nil
		}
		if err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				return nil, errMalformed
			}
			return nil, err
		}
		out = append(out, e)
	}
}

func (e element) int() (int, error) {
	if len(e.value) == 0 || len(e.value) > 4 {
		return 0, errMalformed
	}
	n := 0
	for _, b := range e.value {
		n = n<<8 | int(b)
	}
	return n, nil
}
