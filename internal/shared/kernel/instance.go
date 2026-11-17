package kernel

import "context"

// ProductName is what the instance is called without an organisation's
// own branding.
const ProductName = "Waylight"

// InstanceName returns what the instance is called right now: the
// organisation's name from the branding (#305), or ProductName. The
// composition root supplies it; modules call Of, which works when it is nil.
type InstanceName func(ctx context.Context) string

// Of returns the name, ProductName when f is nil or answers nothing.
func (f InstanceName) Of(ctx context.Context) string {
	if f == nil {
		return ProductName
	}
	if n := f(ctx); n != "" {
		return n
	}
	return ProductName
}
