// Package domain holds the entries of the audit log and its rules (#303).
//
// Ring: Entities.
package domain

import (
	"errors"
	"time"
)

// Entry is one recorded change: who, when, what kind of change on which
// object. The content of the object is never recorded.
type Entry struct {
	ID        string
	At        time.Time
	AccountID string
	Role      string
	Module    string
	Route     string
	ObjectID  string
}

// Retention limits, in days. Zero keeps entries for good.
const (
	DefaultRetentionDays = 365
	MinRetentionDays     = 30
	MaxRetentionDays     = 3650
)

// ErrRetention is a retention outside the limits.
var ErrRetention = errors.New("the retention must be 0 (keep for good) or between 30 and 3650 days")

// ValidRetention checks a retention period.
func ValidRetention(days int) error {
	if days == 0 || (days >= MinRetentionDays && days <= MaxRetentionDays) {
		return nil
	}
	return ErrRetention
}

// Filter selects entries; empty fields match everything.
type Filter struct {
	AccountID string
	Module    string
	From, To  time.Time
	Limit     int
}

// MaxLimit bounds one page of the log in the interface.
const MaxLimit = 500
