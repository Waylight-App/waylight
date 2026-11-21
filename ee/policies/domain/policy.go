// Package domain holds the admin policies of the ee/ features (#304).
//
// Ring: Entities.
package domain

import (
	"slices"
	"strings"
	"time"

	"waylight/internal/shared/kernel"
)

// Settings are the policies as admins set them.
type Settings struct {
	RequireSecondFactor bool
	// MinLength of new passwords above the built-in 12; 0 keeps 12.
	MinLength int
	// MaxSessionHours ends sessions this long after sign-in; 0 = no limit.
	MaxSessionHours int
	// MemberModules restricts members to these modules; nil = all modules.
	MemberModules []string
}

// Limits of the settings.
const (
	BuiltInMinPassword = 12
	MaxPasswordPolicy  = 64
	MaxSessionHours    = 24 * 90
)

// ErrInvalid explains settings outside the limits.
var ErrInvalid = kernel.NewPublicError(kernel.ErrInvalid, "The minimum password length must be 0 or 12 to 64, the session length 0 to 2160 hours.")

// Validate checks the limits and normalises the module list.
func (s Settings) Validate() (Settings, error) {
	if s.MinLength != 0 && (s.MinLength < BuiltInMinPassword || s.MinLength > MaxPasswordPolicy) {
		return Settings{}, ErrInvalid
	}
	if s.MaxSessionHours < 0 || s.MaxSessionHours > MaxSessionHours {
		return Settings{}, ErrInvalid
	}
	if s.MemberModules != nil {
		var mods []string
		for _, m := range s.MemberModules {
			m = strings.ToLower(strings.TrimSpace(m))
			if m != "" && !slices.Contains(mods, m) {
				mods = append(mods, m)
			}
		}
		slices.Sort(mods)
		s.MemberModules = append([]string{}, mods...)
	}
	return s, nil
}

// MaxSessionAge is the session limit as a duration.
func (s Settings) MaxSessionAge() time.Duration { return time.Duration(s.MaxSessionHours) * time.Hour }
