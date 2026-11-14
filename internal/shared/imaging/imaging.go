// Package imaging decodes uploaded pictures, turns them upright and scales
// them down (#286). It uses only the standard library: a box filter is
// plenty for shrinking photos to screen size, and it keeps the binary lean
// (ADR 0011).
package imaging

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
)

// ErrUnsupported means the data is not a JPEG, PNG or GIF picture.
var ErrUnsupported = errors.New("imaging: unsupported picture format")

// ErrTooManyPixels guards against decompression bombs: a small file can
// claim a huge canvas that would exhaust memory when decoded.
var ErrTooManyPixels = errors.New("imaging: picture has too many pixels")

// MaxPixels is the largest canvas Decode accepts (about a 50-megapixel photo).
const MaxPixels = 50_000_000

// Decode reads a JPEG, PNG or GIF (first frame) and applies the EXIF
// orientation of JPEGs, so phone photos come out upright.
func Decode(data []byte) (image.Image, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, ErrUnsupported
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width*cfg.Height > MaxPixels {
		return nil, ErrTooManyPixels
	}
	var img image.Image
	switch format {
	case "jpeg":
		img, err = jpeg.Decode(bytes.NewReader(data))
	case "png":
		img, err = png.Decode(bytes.NewReader(data))
	case "gif":
		img, err = gif.Decode(bytes.NewReader(data))
	default:
		return nil, ErrUnsupported
	}
	if err != nil {
		return nil, fmt.Errorf("imaging: decode %s: %w", format, err)
	}
	if format == "jpeg" {
		img = Orient(img, jpegOrientation(data))
	}
	return img, nil
}

// Fit scales img down so that its longer edge is at most maxEdge pixels,
// averaging the source pixels that fall into each target pixel. Smaller
// pictures are returned unchanged; nothing is ever scaled up.
func Fit(img image.Image, maxEdge int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxEdge && h <= maxEdge {
		return img
	}
	tw, th := maxEdge, max(1, h*maxEdge/w)
	if h > w {
		tw, th = max(1, w*maxEdge/h), maxEdge
	}
	src := toRGBA(img)
	dst := image.NewRGBA(image.Rect(0, 0, tw, th))
	for y := range th {
		y0, y1 := y*h/th, max((y+1)*h/th, y*h/th+1)
		for x := range tw {
			x0, x1 := x*w/tw, max((x+1)*w/tw, x*w/tw+1)
			var r, g, bl, a, n uint32
			for sy := y0; sy < y1; sy++ {
				row := src.Pix[sy*src.Stride:]
				for sx := x0; sx < x1; sx++ {
					p := row[sx*4 : sx*4+4 : sx*4+4]
					r, g, bl, a = r+uint32(p[0]), g+uint32(p[1]), bl+uint32(p[2]), a+uint32(p[3])
					n++
				}
			}
			o := dst.PixOffset(x, y)
			// An average of bytes is itself at most 255.
			dst.Pix[o], dst.Pix[o+1], dst.Pix[o+2], dst.Pix[o+3] = uint8(r/n), uint8(g/n), uint8(bl/n), uint8(a/n) //nolint:gosec // see above
		}
	}
	return dst
}

// EncodeJPEG writes img as a JPEG; transparent areas become white, since
// JPEG has no alpha channel and black would look like an error.
func EncodeJPEG(img image.Image, quality int) ([]byte, error) {
	flat := image.NewRGBA(img.Bounds())
	draw.Draw(flat, flat.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(flat, flat.Bounds(), img, img.Bounds().Min, draw.Over)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, flat, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("imaging: encode jpeg: %w", err)
	}
	return buf.Bytes(), nil
}

// toRGBA returns img as a zero-based *image.RGBA (non-premultiplied
// averaging is not needed: the box filter works on premultiplied values,
// which is what image.RGBA stores).
func toRGBA(img image.Image) *image.RGBA {
	if r, ok := img.(*image.RGBA); ok && r.Rect.Min == (image.Point{}) {
		return r
	}
	b := img.Bounds()
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(out, out.Bounds(), img, b.Min, draw.Src)
	return out
}
