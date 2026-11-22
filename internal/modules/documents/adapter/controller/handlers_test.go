package controller

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"waylight/internal/modules/documents/domain"
	"waylight/internal/modules/documents/usecase"
)

type fakeUC struct {
	usecase.Documents
	name string
	data []byte
}

func (f *fakeUC) AttachFile(_ context.Context, _, name string, data []byte) (usecase.DocumentView, error) {
	f.name, f.data = name, data
	_, err := domain.NewFile(name, data)
	return usecase.DocumentView{}, err
}

func (f *fakeUC) File(context.Context, string) (domain.Document, []byte, error) {
	return domain.Document{File: domain.File{Name: "Quittung ä.pdf", ContentType: "application/pdf"}}, []byte("%PDF"), nil
}

func upload(h *Handlers, field string, data []byte) *httptest.ResponseRecorder {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile(field, "receipt.pdf")
	_, _ = fw.Write(data)
	_ = mw.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/documents/1/file", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	h.UploadFile(rec, req)
	return rec
}

func TestUploadFile(t *testing.T) {
	uc := &fakeUC{}
	h := NewHandlers(uc)
	if rec := upload(h, "file", []byte("%PDF-1.4")); rec.Code != http.StatusOK || uc.name != "receipt.pdf" {
		t.Fatalf("ok: %d %s", rec.Code, rec.Body)
	}
	if rec := upload(h, "file", []byte("<html>")); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("type: %d", rec.Code)
	}
	if rec := upload(h, "other", []byte("%PDF")); rec.Code != http.StatusBadRequest {
		t.Fatalf("field: %d", rec.Code)
	}
	if rec := upload(h, "file", bytes.Repeat([]byte{0}, domain.MaxFileBytes+uploadOverhead)); rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("size: %d", rec.Code)
	}
	rec := httptest.NewRecorder()
	h.UploadFile(rec, httptest.NewRequest(http.MethodPost, "/api/documents/1/file", strings.NewReader("{}")))
	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("json: %d", rec.Code)
	}
}

func TestFileIsAnAttachment(t *testing.T) {
	rec := httptest.NewRecorder()
	NewHandlers(&fakeUC{}).File(rec, httptest.NewRequest(http.MethodGet, "/api/documents/1/file", nil))
	h := rec.Header()
	if rec.Code != http.StatusOK || h.Get("Content-Type") != "application/pdf" || h.Get("X-Content-Type-Options") != "nosniff" ||
		!strings.HasPrefix(h.Get("Content-Disposition"), "attachment;") || !strings.Contains(h.Get("Content-Disposition"), "filename*=UTF-8''Quittung%20%C3%A4.pdf") {
		t.Fatalf("%d %v", rec.Code, h)
	}
}
