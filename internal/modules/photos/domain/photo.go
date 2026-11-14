// Package domain holds the photo frame entities and their limits (#286).
//
// Ring: Entities.
package domain

import (
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"trisse-home/internal/shared/kernel"
)

// Limits. Pictures are stored once in screen size, so 300 MB hold several
// thousand photos while keeping backups of a small server reasonable.
const (
	MaxUploadBytes = 25 << 20
	QuotaBytes     = 300 << 20
	MaxPhotos      = 2000
	// MaxEdge is the longer side of a stored picture: sharp on a full-HD
	// screen, about 300–500 KB as JPEG.
	MaxEdge = 1920
	// ThumbEdge is the longer side of the thumbnails in the settings.
	ThumbEdge   = 320
	JPEGQuality = 82
	// DefaultAlbum is used when a widget has not chosen one.
	DefaultAlbum  = "Photos"
	maxAlbumRunes = 40
)

// Errors.
var (
	ErrNotFound    = kernel.NewPublicError(kernel.ErrNotFound, "The picture does not exist (any more).")
	ErrQuotaFull   = kernel.NewPublicError(kernel.ErrConflict, "The photo storage is full (300 MB or 2000 pictures). Delete pictures first.")
	ErrUnsupported = kernel.NewValidationError("file", "Upload a JPEG, PNG or GIF picture.")
	ErrTooLarge    = kernel.NewValidationError("file", "A picture must be at most 25 MB.")
	ErrAlbum       = kernel.NewValidationError("album", "The album name must have between 1 and 40 characters.")
	ErrOrder       = kernel.NewValidationError("ids", "The order must list every picture of the album exactly once.")
)

// Photo is one stored picture; Size counts the picture and its thumbnail.
type Photo struct {
	ID        string
	Album     string
	Position  int
	Width     int
	Height    int
	Size      int64
	CreatedAt time.Time
}

// Album summarises the pictures of one album.
type Album struct {
	Name  string
	Count int
	Bytes int64
}

// Usage is the storage used by all albums.
type Usage struct {
	Count int
	Bytes int64
}

// Fits reports whether one more picture of size bytes stays within the limits.
func (u Usage) Fits(size int64) bool {
	return u.Count < MaxPhotos && u.Bytes+size <= QuotaBytes
}

// NormaliseAlbum trims and validates an album name; empty means the default album.
func NormaliseAlbum(name string) (string, error) {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		return DefaultAlbum, nil
	}
	if utf8.RuneCountInString(name) > maxAlbumRunes {
		return "", ErrAlbum
	}
	return name, nil
}

// CheckOrder verifies that ids is a permutation of the album's pictures.
func CheckOrder(photos []Photo, ids []string) error {
	if len(ids) != len(photos) {
		return ErrOrder
	}
	have := make([]string, 0, len(photos))
	for _, p := range photos {
		have = append(have, p.ID)
	}
	want := slices.Clone(ids)
	slices.Sort(have)
	slices.Sort(want)
	if !slices.Equal(have, want) {
		return ErrOrder
	}
	return nil
}
