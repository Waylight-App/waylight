package domain

// Size is one of the widget sizes shared by the phone grid (#585) and the
// web grid (#595). On the phone it places the widget on a grid of four
// columns; on the web it decides between the widget's compact view (small,
// strip) and its full view (medium, large). The empty size keeps the older
// behaviour: full width on the phone, a freely sized tile on the web.
type Size string

// The sizes, with their cells on the phone grid of four columns.
const (
	SizeAuto   Size = ""
	SizeSmall  Size = "small"  // 2 × 2
	SizeMedium Size = "medium" // 4 × 2
	SizeLarge  Size = "large"  // 4 × 4
	SizeStrip  Size = "strip"  // 4 × 1
)

// Valid reports whether s is one of the sizes or empty.
func (s Size) Valid() bool {
	switch s {
	case SizeAuto, SizeSmall, SizeMedium, SizeLarge, SizeStrip:
		return true
	}
	return false
}
