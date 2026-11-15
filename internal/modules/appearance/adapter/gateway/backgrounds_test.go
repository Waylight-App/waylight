package gateway

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"testing"
	"time"

	"trisse-home/internal/modules/appearance/domain"
	"trisse-home/internal/modules/appearance/usecase"
	"trisse-home/internal/platform/database/dbtest"
)

func TestBackgroundStore(t *testing.T) {
	ctx := context.Background()
	s := NewBackgroundStore(dbtest.Open(t, "appearance", Migrations()))
	if err := s.AddBackground(ctx, domain.Background{ID: "a", Width: 2, Height: 1, Size: 3, CreatedAt: time.Now()}, []byte("img"), []byte("th")); err != nil {
		t.Fatal(err)
	}
	if list, _ := s.ListBackgrounds(ctx); len(list) != 1 || list[0].Width != 2 {
		t.Fatalf("%+v", list)
	}
	if d, _ := s.BackgroundImage(ctx, "a", true); string(d) != "th" {
		t.Fatalf("%q", d)
	}
	if err := s.DeleteBackground(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.BackgroundImage(ctx, "a", false); !errors.Is(err, domain.ErrBackgroundNotFound) {
		t.Fatal(err)
	}
}

func TestImageProcessor(t *testing.T) {
	var buf bytes.Buffer
	_ = png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4000, 2000)))
	out, err := ImageProcessor{}.Process(buf.Bytes(), domain.MaxBackgroundEdge, domain.BackgroundThumbEdge)
	if err != nil || out.Width != domain.MaxBackgroundEdge || out.Height != domain.MaxBackgroundEdge/2 {
		t.Fatalf("%v %dx%d", err, out.Width, out.Height)
	}
	if _, err := (ImageProcessor{}).Process([]byte("<svg/>"), 10, 5); !errors.Is(err, usecase.ErrUnsupportedImage) {
		t.Fatal(err)
	}
}
