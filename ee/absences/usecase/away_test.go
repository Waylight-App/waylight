package usecase

import (
	"context"
	"testing"

	"waylight/ee/absences/domain"
)

type memStore struct{ list []domain.Absence }

func (m *memStore) Between(_ context.Context, from, to string) ([]domain.Absence, error) {
	var out []domain.Absence
	for _, a := range m.list {
		if a.From <= to && a.To >= from {
			out = append(out, a)
		}
	}
	return out, nil
}
func (m *memStore) Get(context.Context, string) (domain.Absence, error) {
	return domain.Absence{}, domain.ErrNotFound
}
func (m *memStore) Save(context.Context, domain.Absence) error  { return nil }
func (m *memStore) Delete(context.Context, string) error        { return nil }
func (m *memStore) DeleteAccount(context.Context, string) error { return nil }

func TestAwayOn(t *testing.T) {
	store := &memStore{list: []domain.Absence{
		{AccountID: "anna", From: "2026-11-16", To: "2026-11-20", Kind: domain.KindHoliday},
		{AccountID: "ben", From: "2026-11-16", To: "2026-11-16", Kind: domain.KindRemote},
		{AccountID: "cleo", From: "2026-11-16", To: "2026-11-17", Kind: domain.KindRemote, Private: true},
		{AccountID: "anna", From: "2026-11-14", To: "2026-11-22", Kind: domain.KindSick, Private: true},
		{AccountID: "dora", From: "2026-11-17", To: "2026-11-18", Kind: domain.KindHoliday},
	}}
	s := New(Deps{Store: store})
	got, err := s.AwayOn(context.Background(), "2026-11-16")
	if err != nil {
		t.Fatal(err)
	}
	want := []Away{{AccountID: "anna", Until: "2026-11-22"}, {AccountID: "ben", Until: "2026-11-16", Remote: true}, {AccountID: "cleo", Until: "2026-11-17"}}
	if len(got) != len(want) {
		t.Fatalf("got %+v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%d: got %+v, want %+v (a private remote day never says remote; the longest absence wins)", i, got[i], want[i])
		}
	}
	if _, err := s.AwayOn(context.Background(), "16.11."); err == nil {
		t.Fatal("invalid day")
	}
}
