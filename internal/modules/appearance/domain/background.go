package domain

import (
	"time"

	"trisse-home/internal/shared/kernel"
)

// Limits of uploaded backgrounds (#277). A background is shown full screen,
// so it is kept larger than a photo frame picture but still bounded.
const (
	MaxBackgroundUploadBytes = 25 << 20
	MaxBackgroundEdge        = 2560
	BackgroundThumbEdge      = 320
	BackgroundJPEGQuality    = 82
	MaxBackgrounds           = 30
)

// Errors of uploaded backgrounds.
var (
	ErrBackgroundNotFound    = kernel.NewPublicError(kernel.ErrNotFound, "The background picture does not exist (any more).")
	ErrBackgroundUnsupported = kernel.NewValidationError("file", "Upload a JPEG, PNG or GIF picture.")
	ErrBackgroundTooLarge    = kernel.NewValidationError("file", "A background picture must be at most 25 MB.")
	ErrTooManyBackgrounds    = kernel.NewPublicError(kernel.ErrConflict, "At most 30 background pictures can be stored. Delete one first.")
)

// Background is an uploaded background picture.
type Background struct {
	ID        string
	Width     int
	Height    int
	Size      int64
	CreatedAt time.Time
}

// URL is the value stored as background when this picture is chosen.
func (b Background) URL() string { return kernel.UploadedBackgroundPrefix + b.ID }
