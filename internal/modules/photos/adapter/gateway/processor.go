package gateway

import (
	"errors"

	"trisse-home/internal/modules/photos/domain"
	"trisse-home/internal/modules/photos/usecase"
	"trisse-home/internal/shared/imaging"
)

// Processor implements usecase.Processor with the standard library.
type Processor struct{}

// Process implements usecase.Processor: upright, at most domain.MaxEdge
// pixels, re-encoded as JPEG (which also drops EXIF data such as the
// location), plus a small thumbnail for the settings.
func (Processor) Process(data []byte) (usecase.Processed, error) {
	img, err := imaging.Decode(data)
	if errors.Is(err, imaging.ErrUnsupported) || errors.Is(err, imaging.ErrTooManyPixels) {
		return usecase.Processed{}, usecase.ErrUnsupportedImage
	}
	if err != nil {
		return usecase.Processed{}, err
	}
	full := imaging.Fit(img, domain.MaxEdge)
	image, err := imaging.EncodeJPEG(full, domain.JPEGQuality)
	if err != nil {
		return usecase.Processed{}, err
	}
	thumb, err := imaging.EncodeJPEG(imaging.Fit(full, domain.ThumbEdge), domain.JPEGQuality)
	if err != nil {
		return usecase.Processed{}, err
	}
	b := full.Bounds()
	return usecase.Processed{Image: image, Thumb: thumb, Width: b.Dx(), Height: b.Dy()}, nil
}
