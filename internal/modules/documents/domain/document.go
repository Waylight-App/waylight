package domain

import (
	"net/http"
	"path"
	"strings"
	"time"
	"unicode/utf8"

	"waylight/internal/shared/kernel"
)

// Limits of a document.
const (
	MaxTitleLength = 80
	MaxNoteLength  = 500
	MaxFileBytes   = 10 << 20
	maxNameRunes   = 120
	// EndingSoonDays is how long before its end a warranty counts as ending.
	EndingSoonDays = 30
)

// Errors.
var (
	// ErrNotFound is returned for unknown documents and for documents the
	// viewer may not see.
	ErrNotFound  = kernel.NewPublicError(kernel.ErrNotFound, "The document does not exist.")
	ErrNoFile    = kernel.NewPublicError(kernel.ErrNotFound, "The document has no file.")
	errEmptyFile = kernel.NewValidationError("file", "The file is empty.")
	errFileLarge = kernel.NewValidationError("file", "The file must be at most 10 MB.")
	errFileType  = kernel.NewValidationError("file", "Choose a photo or a PDF.")
)

// Status is the state of a warranty on one day.
type Status string

// The states a warranty can be in.
const (
	StatusNone  Status = "none" // no warranty end entered
	Running     Status = "running"
	EndingSoon  Status = "ending"
	StatusEnded Status = "ended"
)

// File describes the file attached to a document.
type File struct {
	Name        string
	ContentType string
	Size        int64
}

// Document is a product with its warranty and an optional file (#688).
type Document struct {
	ID    string
	Title string
	// PurchasedOn and WarrantyEnd are midnight UTC; zero when not entered.
	PurchasedOn time.Time
	WarrantyEnd time.Time
	Note        string
	File        File
	Access      kernel.Access
	CreatedAt   time.Time
}

// Input is what a person enters for a document.
type Input struct {
	Title       string
	PurchasedOn string // YYYY-MM-DD or empty
	WarrantyEnd string // YYYY-MM-DD or empty
	Note        string
}

// New validates the input and creates a document.
func New(id string, in Input, access kernel.Access, now time.Time) (Document, error) {
	d := Document{ID: id, Access: access, CreatedAt: now}
	if err := d.apply(in); err != nil {
		return Document{}, err
	}
	return d, nil
}

// Update changes everything but the owner and the file.
func (d Document) Update(in Input) (Document, error) {
	err := d.apply(in)
	return d, err
}

func optionalDay(field, text string) (time.Time, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return time.Time{}, nil
	}
	day, err := time.Parse(time.DateOnly, text)
	if err != nil {
		return time.Time{}, kernel.NewValidationError(field, "Enter a date like 2026-12-24.")
	}
	return day, nil
}

func (d *Document) apply(in Input) error {
	title := strings.TrimSpace(in.Title)
	if title == "" || utf8.RuneCountInString(title) > MaxTitleLength {
		return kernel.NewValidationError("title", "The title must be between 1 and 80 characters long.")
	}
	bought, err := optionalDay("purchasedOn", in.PurchasedOn)
	if err != nil {
		return err
	}
	end, err := optionalDay("warrantyEnd", in.WarrantyEnd)
	if err != nil {
		return err
	}
	if !bought.IsZero() && !end.IsZero() && end.Before(bought) {
		return kernel.NewValidationError("warrantyEnd", "The warranty cannot end before the purchase.")
	}
	note := strings.TrimSpace(in.Note)
	if utf8.RuneCountInString(note) > MaxNoteLength {
		return kernel.NewValidationError("note", "The note may be at most 500 characters long.")
	}
	d.Title, d.PurchasedOn, d.WarrantyEnd, d.Note = title, bought, end, note
	return nil
}

// StatusOn returns the warranty state on day; the end day itself still
// counts as covered.
func (d Document) StatusOn(day time.Time) Status {
	if d.WarrantyEnd.IsZero() {
		return StatusNone
	}
	today := dateOf(day)
	switch {
	case d.WarrantyEnd.Before(today):
		return StatusEnded
	case !d.WarrantyEnd.After(today.AddDate(0, 0, EndingSoonDays)):
		return EndingSoon
	}
	return Running
}

// Day formats a stored day as YYYY-MM-DD; empty for a day not entered.
func Day(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.DateOnly)
}

func dateOf(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// NewFile checks an upload and describes it. The content type comes from the
// bytes, never from what the client claims, so a renamed script cannot pass
// as a photo.
func NewFile(name string, data []byte) (File, error) {
	if len(data) == 0 {
		return File{}, errEmptyFile
	}
	if len(data) > MaxFileBytes {
		return File{}, errFileLarge
	}
	ct := http.DetectContentType(data)
	switch ct {
	case "image/jpeg", "image/png", "image/webp", "application/pdf":
	default:
		return File{}, errFileType
	}
	return File{Name: CleanFileName(name), ContentType: ct, Size: int64(len(data))}, nil
}

// CleanFileName keeps the base name without path, control characters or
// quotes, at most 120 characters; "file" when nothing is left.
func CleanFileName(name string) string {
	name = path.Base(strings.ReplaceAll(strings.TrimSpace(name), `\`, "/"))
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || r == '"' || r == '/' {
			return -1
		}
		return r
	}, name)
	if name == "" || name == "." || name == ".." {
		return "file"
	}
	if r := []rune(name); len(r) > maxNameRunes {
		name = string(r[:maxNameRunes])
	}
	return name
}
