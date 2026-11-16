// Package domain holds the parts of SCIM 2.0 (RFC 7643, RFC 7644) that
// trisse-home implements for directory provisioning (#396).
//
// Ring: Entities.
package domain

import (
	"errors"
	"strconv"
	"strings"
)

// ErrUnsupportedFilter answers filters beyond `attribute eq "value"`, which is
// what identity providers send to look up a user or group before creating it.
var ErrUnsupportedFilter = errors.New(`only filters of the form attribute eq "value" are supported`)

// Filter is an equality filter on one attribute.
type Filter struct {
	Attribute string // lowercase, e.g. "username", "externalid", "displayname"
	Value     string
}

// ParseFilter reads `userName eq "anna"`; an empty filter matches everything.
func ParseFilter(raw string) (*Filter, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	attr, rest, ok := strings.Cut(raw, " ")
	if !ok {
		return nil, ErrUnsupportedFilter
	}
	op, value, ok := strings.Cut(strings.TrimSpace(rest), " ")
	if !ok || !strings.EqualFold(op, "eq") {
		return nil, ErrUnsupportedFilter
	}
	unquoted, err := strconv.Unquote(strings.TrimSpace(value))
	if err != nil {
		return nil, ErrUnsupportedFilter
	}
	return &Filter{Attribute: strings.ToLower(attr), Value: unquoted}, nil
}

// Matches compares case-insensitively, as SCIM defines for userName.
func (f *Filter) Matches(attribute, value string) bool {
	return f == nil || (f.Attribute == attribute && strings.EqualFold(f.Value, value))
}

// ParseBool reads true/false as JSON booleans or strings; Microsoft Entra ID
// sends "True" and "False" as strings in PATCH requests.
func ParseBool(v any) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true
	case string:
		b, err := strconv.ParseBool(strings.ToLower(x))
		return b, err == nil
	}
	return false, false
}

// MemberPath reads the member ID out of a path like `members[value eq "id"]`.
func MemberPath(path string) (string, bool) {
	inner, ok := strings.CutPrefix(strings.TrimSpace(path), "members[")
	if !ok {
		return "", false
	}
	inner, ok = strings.CutSuffix(inner, "]")
	if !ok {
		return "", false
	}
	f, err := ParseFilter(inner)
	if err != nil || f == nil || f.Attribute != "value" {
		return "", false
	}
	return f.Value, true
}
