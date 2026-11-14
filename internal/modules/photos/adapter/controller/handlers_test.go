package controller

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trisse-home/internal/modules/photos/domain"
	"trisse-home/internal/modules/photos/usecase"
)

type fakeUC struct {
	usecase.Photos
	got []byte
}

func (f *fakeUC) Upload(_ context.Context, album string, data []byte) (domain.Photo, error) {
	f.got = data
	return domain.Photo{ID: "1", Album: album}, nil
}

func (f *fakeUC) Image(context.Context, string, bool) ([]byte, error) { return []byte("jpeg"), nil }

func TestUpload(t *testing.T) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "a.jpg")
	_, _ = fw.Write([]byte("picture"))
	_ = mw.Close()
	uc := &fakeUC{}
	req := httptest.NewRequest(http.MethodPost, "/api/photos?album=Summer", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	NewHandlers(uc).Upload(rec, req)
	if rec.Code != http.StatusCreated || string(uc.got) != "picture" || !strings.Contains(rec.Body.String(), `"album":"Summer"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}

	rec = httptest.NewRecorder()
	NewHandlers(uc).Upload(rec, httptest.NewRequest(http.MethodPost, "/api/photos", strings.NewReader("raw")))
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("raw body: %d", rec.Code)
	}
}

func TestImageIsCacheable(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(&fakeUC{}).Image(rec, httptest.NewRequest(http.MethodGet, "/api/photos/1/image", nil))
	if rec.Header().Get("Content-Type") != "image/jpeg" || !strings.Contains(rec.Header().Get("Cache-Control"), "immutable") {
		t.Fatalf("%v", rec.Header())
	}
}
