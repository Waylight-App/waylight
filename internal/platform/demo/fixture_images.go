package demo

import (
	"bytes"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"math"
)

// palette holds the warm and night colours of the Waylight look; pictures
// pick two of them by their seed.
var palette = []color.RGBA{
	{0xf2, 0xb4, 0x41, 0xff}, {0x13, 0x18, 0x29, 0xff}, {0x2e, 0x36, 0x50, 0xff},
	{0xe9, 0xa4, 0x1f, 0xff}, {0x8f, 0xd6, 0xa8, 0xff}, {0x6b, 0x76, 0x94, 0xff},
	{0xff, 0xd2, 0x7a, 0xff}, {0x1f, 0x7a, 0x45, 0xff}, {0xb3, 0x26, 0x1e, 0xff},
}

// picture is a PNG with a diagonal gradient between two palette colours: a
// stand-in for covers, photos and thumbnails that never comes from the
// internet.
func picture(seed string, w, h int) string {
	f := fnv.New32a()
	_, _ = f.Write([]byte(seed))
	n := f.Sum32()
	a, b := palette[int(n%uint32(len(palette)))], palette[int((n/7+3)%uint32(len(palette)))]
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			t := float64(x+y) / float64(w+h)
			img.Set(x, y, color.RGBA{
				R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
				G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
				B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
				A: 0xff,
			})
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img) // in-memory buffer, cannot fail
	return buf.String()
}

// aurora is the picture of the day: green curtains of light over a dark sea
// with a thin horizon.
func aurora(w, h int) string {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	horizon := h * 3 / 4
	for y := range h {
		for x := range w {
			sky := 1 - float64(y)/float64(horizon)
			c := color.RGBA{uint8(8 + 10*sky), uint8(12 + 14*sky), uint8(30 + 30*sky), 0xff}
			if y < horizon {
				// Two wavy curtains, brightest along their lower edge.
				for i, base := range []float64{0.55, 0.35} {
					edge := float64(horizon) * (base + 0.08*math.Sin(float64(x)/float64(w)*7+float64(i)*2))
					d := edge - float64(y)
					if d > 0 && d < float64(h)/3 {
						v := math.Exp(-d/(float64(h)/10)) * (0.6 + 0.4*math.Sin(float64(x)/9))
						c.G = uint8(math.Min(255, float64(c.G)+200*v))
						c.R = uint8(math.Min(255, float64(c.R)+40*v))
						c.B = uint8(math.Min(255, float64(c.B)+70*v))
					}
				}
			} else {
				c = color.RGBA{6, 9, 18, 0xff}
				if y == horizon && x%37 < 3 {
					c = color.RGBA{0xf2, 0xb4, 0x41, 0xff} // lights of a ferry
				}
			}
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img) // in-memory buffer, cannot fail
	return buf.String()
}
