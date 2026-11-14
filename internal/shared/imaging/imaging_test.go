package imaging

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func solid(w, h int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestFitScalesDownKeepingTheAspectRatio(t *testing.T) {
	got := Fit(solid(400, 200, color.RGBA{200, 100, 50, 255}), 100).Bounds()
	if got.Dx() != 100 || got.Dy() != 50 {
		t.Fatalf("landscape: %v", got)
	}
	got = Fit(solid(200, 400, color.White), 100).Bounds()
	if got.Dx() != 50 || got.Dy() != 100 {
		t.Fatalf("portrait: %v", got)
	}
	small := solid(80, 60, color.White)
	if Fit(small, 100) != image.Image(small) {
		t.Fatal("small pictures are never scaled up")
	}
}

func TestFitAveragesColours(t *testing.T) {
	// Two columns, black and white, shrink to one grey pixel.
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.Black)
	img.Set(1, 0, color.White)
	r, _, _, _ := Fit(img, 1).At(0, 0).RGBA()
	if v := r >> 8; v < 120 || v > 135 {
		t.Fatalf("grey expected, got %d", v)
	}
}

func TestDecodeRejectsNonPictures(t *testing.T) {
	if _, err := Decode([]byte("<svg/>")); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

func TestDecodeRejectsHugeCanvas(t *testing.T) {
	// A PNG header claiming 20000×20000 pixels; DecodeConfig reads only the header.
	var buf bytes.Buffer
	_ = png.Encode(&buf, solid(1, 1, color.White))
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b[16:], 20000)
	binary.BigEndian.PutUint32(b[20:], 20000)
	binary.BigEndian.PutUint32(b[29:], crc32.ChecksumIEEE(b[12:29])) // IHDR checksum
	if _, err := Decode(b); !errors.Is(err, ErrTooManyPixels) {
		t.Fatalf("got %v", err)
	}
}

// withOrientation inserts an EXIF APP1 segment after the SOI marker.
func withOrientation(jpg []byte, o uint16) []byte {
	tiff := []byte{'M', 'M', 0, 42, 0, 0, 0, 8, 0, 1}
	entry := make([]byte, 12)
	binary.BigEndian.PutUint16(entry[0:], 0x0112)
	binary.BigEndian.PutUint16(entry[2:], 3) // SHORT
	binary.BigEndian.PutUint32(entry[4:], 1)
	binary.BigEndian.PutUint16(entry[8:], o)
	tiff = append(append(tiff, entry...), 0, 0, 0, 0)
	seg := append([]byte("Exif\x00\x00"), tiff...)
	app1 := []byte{0xFF, 0xE1, 0, 0}
	binary.BigEndian.PutUint16(app1[2:], uint16(len(seg)+2))
	out := append([]byte{}, jpg[:2]...)
	out = append(append(out, app1...), seg...)
	return append(out, jpg[2:]...)
}

func TestDecodeTurnsPhonePhotosUpright(t *testing.T) {
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, solid(40, 20, color.White), nil)
	img, err := Decode(withOrientation(buf.Bytes(), 6))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 20 || b.Dy() != 40 {
		t.Fatalf("orientation 6 should swap the sides: %v", b)
	}
	if o := jpegOrientation(buf.Bytes()); o != 1 {
		t.Fatalf("no EXIF means upright, got %d", o)
	}
}

func TestOrientMovesPixels(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 1))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255}) // red left, transparent right
	rot := Orient(img, 6)                     // 90° clockwise: red ends up on top
	if r, _, _, _ := rot.At(0, 0).RGBA(); r>>8 != 255 {
		t.Fatalf("red should be at the top: %v", rot.At(0, 0))
	}
	if r, _, _, _ := Orient(img, 3).At(1, 0).RGBA(); r>>8 != 255 {
		t.Fatal("180° should move red to the right")
	}
}

func TestEncodeJPEGFlattensTransparency(t *testing.T) {
	data, err := EncodeJPEG(image.NewRGBA(image.Rect(0, 0, 4, 4)), 80)
	if err != nil {
		t.Fatal(err)
	}
	img, _ := jpeg.Decode(bytes.NewReader(data))
	if r, _, _, _ := img.At(1, 1).RGBA(); r>>8 < 240 {
		t.Fatalf("transparent should become white, got %d", r>>8)
	}
}
