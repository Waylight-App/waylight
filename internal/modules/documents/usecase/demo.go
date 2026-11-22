package usecase

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"time"

	"waylight/internal/modules/documents/domain"
)

// samplePDF is a one-page PDF without content, enough to download and open.
const samplePDF = "%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n" +
	"3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 200 200]>>endobj\ntrailer<</Root 1 0 R/Size 4>>\n%%EOF\n"

// samplePNG is a small grey square standing in for a photographed receipt.
func samplePNG() []byte {
	img := image.NewGray(image.Rect(0, 0, 48, 48))
	for i := range img.Pix {
		img.Pix[i] = 0xd8
	}
	img.SetGray(24, 24, color.Gray{Y: 0x40})
	var b bytes.Buffer
	_ = png.Encode(&b, img) // encoding into memory cannot fail
	return b.Bytes()
}

// SeedDemo adds a few documents with small sample files to the public demo,
// once. Dates are relative to today, so one warranty always ends soon.
func SeedDemo(ctx context.Context, d Deps) error {
	c := NewDocuments(d)
	existing, err := c.List(ctx)
	if err != nil || len(existing) > 0 {
		return err
	}
	day := func(n int) string { return d.Clock.Now().AddDate(0, 0, n).Format(time.DateOnly) }
	for _, s := range []struct {
		in   domain.Input
		name string
		data []byte
	}{
		{domain.Input{Title: "Washing machine", PurchasedOn: day(-700), WarrantyEnd: day(395), Note: "Receipt from the electronics store"}, "receipt.pdf", []byte(samplePDF)},
		{domain.Input{Title: "Headphones", PurchasedOn: day(-335), WarrantyEnd: day(20)}, "receipt.png", samplePNG()},
		{domain.Input{Title: "Laptop", PurchasedOn: day(-1200), WarrantyEnd: day(-100), Note: "Two years, ended"}, "invoice.pdf", []byte(samplePDF)},
		{domain.Input{Title: "Passport copy"}, "passport.png", samplePNG()},
	} {
		v, err := c.Create(ctx, s.in, "")
		if err != nil {
			return err
		}
		if _, err := c.AttachFile(ctx, v.ID, s.name, s.data); err != nil {
			return err
		}
	}
	return nil
}
