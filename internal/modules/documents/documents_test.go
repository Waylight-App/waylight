package documents

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"waylight/internal/modules/documents/adapter/gateway"
	"waylight/internal/modules/documents/domain"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/kernel"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type seqIDs struct{ n int }

func (s *seqIDs) NewID() string { s.n++; return strconv.Itoa(s.n) }

var pdf = []byte("%PDF-1.4\n")

func newModule(t *testing.T) *Module {
	t.Helper()
	return New(dbtest.Open(t, "documents", gateway.Migrations()), fixedClock{time.Date(2026, 11, 22, 9, 0, 0, 0, time.UTC)}, &seqIDs{}, time.UTC)
}

func TestListOrderAndCompactState(t *testing.T) {
	m := newModule(t)
	ctx := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	if s, _ := m.CompactState(ctx); !s.Empty || s.Title.String() != "No warranties running" {
		t.Fatalf("empty: %+v", s)
	}
	for _, in := range []domain.Input{
		{Title: "Ended", WarrantyEnd: "2026-10-01"},
		{Title: "Later", WarrantyEnd: "2027-06-30"},
		{Title: "No warranty"},
		{Title: "Soon", WarrantyEnd: "2026-12-10"},
	} {
		if _, err := m.uc.Create(ctx, in, ""); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := m.uc.List(ctx)
	var titles []string
	for _, d := range list {
		titles = append(titles, d.Title)
	}
	if got := titles; len(got) != 4 || got[0] != "Soon" || got[1] != "Later" || got[2] != "No warranty" || got[3] != "Ended" {
		t.Fatalf("order %v", titles)
	}
	s, err := m.CompactState(ctx)
	if err != nil || s.Title.String() != "Soon: warranty ends 2026-12-10" || s.Level != kernel.CompactAttention || *s.Count != 1 {
		t.Fatalf("%+v %v", s, err)
	}
}

func TestAttachAndDownload(t *testing.T) {
	m := newModule(t)
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	v, err := m.uc.Create(anna, domain.Input{Title: "Kettle"}, "everyone")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := m.uc.File(anna, v.ID); !errors.Is(err, domain.ErrNoFile) {
		t.Fatalf("no file yet: %v", err)
	}
	if _, err := m.uc.AttachFile(ben, v.ID, "x.pdf", pdf); !errors.Is(err, kernel.ErrNotOwner) {
		t.Fatalf("ben attached to anna's document: %v", err)
	}
	if _, err := m.uc.AttachFile(anna, v.ID, "x.pdf", []byte("<html>")); err == nil {
		t.Fatal("a web page was accepted")
	}
	if got, err := m.uc.AttachFile(anna, v.ID, "receipt.pdf", pdf); err != nil || got.File.ContentType != "application/pdf" {
		t.Fatalf("%+v %v", got, err)
	}
	d, data, err := m.uc.File(ben, v.ID)
	if err != nil || d.File.Name != "receipt.pdf" || string(data) != string(pdf) {
		t.Fatalf("ben downloads: %+v %q %v", d, data, err)
	}
}

func TestSeedDemo(t *testing.T) {
	m := newModule(t)
	ctx := context.Background()
	for range 2 { // the second run must not add anything
		if err := m.SeedDemo(ctx); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := m.uc.List(kernel.WithPrincipal(ctx, kernel.AdminPrincipal()))
	if len(list) != 4 || list[0].File.Name == "" {
		t.Fatalf("%d documents", len(list))
	}
}
