package gateway

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"trisse-home/internal/modules/photos/domain"
	"trisse-home/internal/modules/photos/usecase"
	"trisse-home/internal/platform/database/dbtest"
)

func TestRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "photos", Migrations())
	r := NewRepository(db)
	now := time.Now()
	for _, id := range []string{"a", "b"} {
		if _, err := r.Add(ctx, domain.Photo{ID: id, Album: "A", Width: 2, Height: 1, Size: 10, CreatedAt: now}, []byte("img-"+id), []byte("th")); err != nil {
			t.Fatal(err)
		}
	}
	_, _ = r.Add(ctx, domain.Photo{ID: "c", Album: "B", Size: 5, CreatedAt: now}, []byte("x"), []byte("y"))
	if err := r.Reorder(ctx, "A", []string{"b", "a"}); err != nil {
		t.Fatal(err)
	}
	list, _ := r.List(ctx, "A")
	if len(list) != 2 || list[0].ID != "b" || list[1].Position != 1 {
		t.Fatalf("%+v", list)
	}
	if u, _ := r.Usage(ctx); u.Count != 3 || u.Bytes != 25 {
		t.Fatalf("%+v", u)
	}
	if al, _ := r.Albums(ctx); len(al) != 2 || al[0].Name != "A" || al[0].Count != 2 {
		t.Fatalf("%+v", al)
	}
	if img, _ := r.Image(ctx, "a", false); string(img) != "img-a" {
		t.Fatalf("%q", img)
	}
	if err := r.Delete(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Image(ctx, "a", true); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("file should be gone with the row: %v", err)
	}
	if err := r.Delete(ctx, "a"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
}

func TestProcessor(t *testing.T) {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 4000, 2000))
	img.Set(0, 0, color.White)
	_ = png.Encode(&buf, img)
	out, err := Processor{}.Process(buf.Bytes())
	if err != nil || out.Width != domain.MaxEdge || out.Height != domain.MaxEdge/2 || len(out.Thumb) == 0 {
		t.Fatalf("%v %dx%d", err, out.Width, out.Height)
	}
	if _, err := (Processor{}).Process([]byte("<svg/>")); !errors.Is(err, usecase.ErrUnsupportedImage) {
		t.Fatal(err)
	}
}
