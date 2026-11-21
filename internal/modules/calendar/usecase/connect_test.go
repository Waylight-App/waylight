package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"waylight/internal/modules/calendar/domain"
	"waylight/internal/shared/kernel"
)

type fakeFinder struct{ found []domain.FoundCalendar }

func (f fakeFinder) Discover(_ context.Context, server, _, password string) ([]domain.FoundCalendar, error) {
	if password != "pw" {
		return nil, fmt.Errorf("%w (401)", kernel.ErrTargetAuth)
	}
	return f.found, nil
}

// TestConnect: the setup assistant enters the server address; the
// calendars on it are found, checked and saved, and the preview gets the
// next events (#588).
func TestConnect(t *testing.T) {
	ctx := context.Background()
	repo := &memRepo{}
	finder := fakeFinder{found: []domain.FoundCalendar{
		{URL: "https://c.example/remote.php/dav/calendars/me/personal/", Name: "Personal"},
		{URL: "https://c.example/remote.php/dav/calendars/me/work/", Name: "Work"},
	}}
	uc := NewCalendar(Deps{Repo: repo, Source: &fakeSource{fail: map[string]bool{}}, Writer: &fakeWriter{}, Finder: finder,
		Subscriptions: &memSubs{}, Feeds: &fakeFeeds{}, IDs: &seqIDs{}, Clock: clk{}})

	_, err := uc.Connect(ctx, ConnectInput{Server: "https://c.example", Username: "me", Password: "bad"})
	var pub *kernel.PublicError
	if !errors.As(err, &pub) || pub.Template != "The key is not accepted. Was it copied completely?" {
		t.Fatalf("wrong password: %v", err)
	}
	if repo.s != nil {
		t.Fatal("nothing is saved after a failed check")
	}

	got, err := uc.Connect(ctx, ConnectInput{Server: " https://c.example ", Username: " me ", Password: "pw"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Calendars) != 2 || len(got.Settings.CalendarURLs) != 2 || repo.s == nil || repo.s.Username != "me" {
		t.Fatalf("%+v", got)
	}
	if len(got.Next) == 0 || len(got.Next) > connectPreview {
		t.Fatalf("preview: %+v", got.Next)
	}

	empty := NewCalendar(Deps{Repo: &memRepo{}, Source: &fakeSource{}, Writer: &fakeWriter{}, Finder: fakeFinder{},
		Subscriptions: &memSubs{}, Feeds: &fakeFeeds{}, IDs: &seqIDs{}, Clock: clk{}})
	if _, err := empty.Connect(ctx, ConnectInput{Server: "https://c.example", Username: "me", Password: "pw"}); !errors.Is(err, domain.ErrNoCalendars) {
		t.Fatalf("no calendars: %v", err)
	}
}
