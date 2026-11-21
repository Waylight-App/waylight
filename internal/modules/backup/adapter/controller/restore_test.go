package controller

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"waylight/internal/modules/backup/domain"
	"waylight/internal/modules/backup/usecase"
)

type fakeRestorer struct {
	req     usecase.RestoreRequest
	content string
}

func (f *fakeRestorer) Restore(_ context.Context, req usecase.RestoreRequest) (usecase.RestoreResult, error) {
	f.req = req
	if req.Upload != nil {
		b, _ := io.ReadAll(req.Upload)
		f.content = string(b)
	}
	return usecase.RestoreResult{SafetyCopy: usecase.FileView{Name: "waylight-20261121-120000.db"}}, nil
}

func (f *fakeRestorer) Preview(_ context.Context, name string) (usecase.PreviewView, error) {
	if !domain.ValidName(name) {
		return usecase.PreviewView{}, domain.ErrBackupNotFound
	}
	return usecase.PreviewView{File: usecase.FileView{Name: name}, Preview: domain.Preview{Problem: domain.ErrTooNew, Tabs: 2}}, nil
}

func restoreForm(t *testing.T, confirm string) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	_ = mw.WriteField("password", "secret")
	_ = mw.WriteField("code", " 123456 ")
	_ = mw.WriteField("confirm", confirm)
	_ = mw.WriteField("source", "upload")
	fw, _ := mw.CreateFormFile("file", "backup.db")
	_, _ = fw.Write([]byte("sqlite-bytes"))
	_ = mw.Close()
	return &body, mw.FormDataContentType()
}

func TestRestoreHandler(t *testing.T) {
	uc := &fakeRestorer{}
	body, contentType := restoreForm(t, "replace")
	req := httptest.NewRequest(http.MethodPost, "/api/backups/restore", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()
	NewRestoreHandlers(uc, 1<<20).Restore(rec, req)
	if rec.Code != http.StatusAccepted || uc.req.Password != "secret" || uc.req.Code != "123456" || !uc.req.Confirmed ||
		uc.req.Source != usecase.FromUpload || uc.content != "sqlite-bytes" ||
		!strings.Contains(rec.Body.String(), `"safetyCopy":{"name":"waylight-20261121-120000.db"`) {
		t.Fatalf("%d %s %+v %q", rec.Code, rec.Body, uc.req, uc.content)
	}

	// Only the exact word counts as a confirmation.
	for _, confirm := range []string{"on", "true", "Replace", ""} {
		body, contentType := restoreForm(t, confirm)
		req := httptest.NewRequest(http.MethodPost, "/api/backups/restore", body)
		req.Header.Set("Content-Type", contentType)
		NewRestoreHandlers(uc, 1<<20).Restore(httptest.NewRecorder(), req)
		if uc.req.Confirmed {
			t.Fatalf("%q counted as a confirmation", confirm)
		}
	}

	rec = httptest.NewRecorder()
	NewRestoreHandlers(uc, 1<<20).Restore(rec, httptest.NewRequest(http.MethodPost, "/api/backups/restore", bytes.NewReader([]byte(`{}`))))
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("json body: %d", rec.Code)
	}
}

func TestPreviewHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/backups/{name}/preview", NewRestoreHandlers(&fakeRestorer{}, 1<<20).Preview)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/backups/waylight-20261111-030000.db/preview", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"restorable":false`) ||
		!strings.Contains(rec.Body.String(), `"messageKey":"srv_`) || !strings.Contains(rec.Body.String(), `"tabs":2`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/backups/evil.db/preview", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("invalid name: %d", rec.Code)
	}
}
