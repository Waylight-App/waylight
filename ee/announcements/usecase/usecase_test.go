package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/ee/announcements/domain"
	"waylight/internal/shared/kernel"
)

type memStore struct {
	list map[string]domain.Announcement
}

func (m *memStore) List(context.Context) ([]domain.Announcement, error) {
	var out []domain.Announcement
	for _, a := range m.list {
		out = append(out, a)
	}
	return out, nil
}
func (m *memStore) Get(_ context.Context, id string) (domain.Announcement, error) {
	a, ok := m.list[id]
	if !ok {
		return domain.Announcement{}, domain.ErrNotFound
	}
	return a, nil
}
func (m *memStore) Save(_ context.Context, a domain.Announcement) error { m.list[a.ID] = a; return nil }
func (m *memStore) Delete(_ context.Context, id string) error           { delete(m.list, id); return nil }
func (m *memStore) ReadBy(context.Context, string) (map[string]bool, error) {
	return map[string]bool{}, nil
}
func (m *memStore) MarkRead(context.Context, string, string, time.Time) error { return nil }
func (m *memStore) ForgetAccount(context.Context, string) error               { return nil }
func (m *memStore) Settings(context.Context) (domain.Settings, error)         { return domain.Settings{}, nil }
func (m *memStore) SaveSettings(context.Context, domain.Settings) error       { return nil }

type post struct{ channel, title, body, link string }

type fakeChat struct {
	posts []post
	fail  error
}

func (f *fakeChat) ChatChannels(context.Context) ([]ChatChannel, error) {
	return []ChatChannel{{ID: "c1", Name: "#general", Kind: "slack"}}, nil
}
func (f *fakeChat) PostToChat(_ context.Context, ch, title, body, link string) error {
	if f.fail != nil {
		return f.fail
	}
	f.posts = append(f.posts, post{ch, title, body, link})
	return nil
}

type clock struct{ t time.Time }

func (c clock) Now() time.Time { return c.t }

type ids struct{ n int }

func (i *ids) NewID() string { i.n++; return string(rune('a' + i.n - 1)) }

var admin = kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "owner", Role: "owner"})

func service(chat *fakeChat, licensed bool) (*Service, *memStore) {
	store := &memStore{list: map[string]domain.Announcement{}}
	return New(Deps{Store: store, Chat: chat, MayPost: func(context.Context) bool { return licensed },
		Link:  func(id string) string { return "https://home.example/?announcement=" + id },
		Clock: clock{time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)}, IDs: &ids{}}), store
}

func TestPostsOnceOnPublishAndOnlyOnRequestWhenEditing(t *testing.T) {
	chat := &fakeChat{}
	s, store := service(chat, true)
	v, err := s.Create(admin, Input{Title: "Office closed", Body: "Friday", PostTo: "c1"})
	if err != nil || v.PostError != "" {
		t.Fatalf("%+v %v", v, err)
	}
	if len(chat.posts) != 1 || chat.posts[0] != (post{"c1", "Office closed", "Friday", "https://home.example/?announcement=a"}) {
		t.Fatalf("posts %+v", chat.posts)
	}
	if store.list["a"].PostedChannel != "c1" || store.list["a"].PostedAt.IsZero() {
		t.Fatalf("not recorded: %+v", store.list["a"])
	}
	if _, err := s.Update(admin, "a", Input{Title: "Office closed all day", Body: "Friday"}); err != nil || len(chat.posts) != 1 {
		t.Fatalf("an edit posted again: %v %d", err, len(chat.posts))
	}
	if _, err := s.Update(admin, "a", Input{Title: "Office closed all day", Body: "Friday", PostTo: "c1"}); err != nil || len(chat.posts) != 2 {
		t.Fatalf("asked repost: %v %d", err, len(chat.posts))
	}
}

func TestPostFailureKeepsTheAnnouncementAndHidesDetails(t *testing.T) {
	chat := &fakeChat{fail: kernel.TargetFailure(kernel.ErrUnavailable,
		errors.New(`Post "https://hooks.slack.com/services/T0/B0/secret": dial tcp: timeout`))}
	s, store := service(chat, true)
	v, err := s.Create(admin, Input{Title: "Cake", PostTo: "c1"})
	if err != nil || v.PostError != errPostFailed || len(store.list) != 1 || store.list["a"].PostedChannel != "" {
		t.Fatalf("%+v %v", v, err)
	}
}

func TestPostingNeedsTheLicence(t *testing.T) {
	s, store := service(&fakeChat{}, false)
	if _, err := s.Create(admin, Input{Title: "Cake", PostTo: "c1"}); !errors.Is(err, domain.ErrNoChat) || len(store.list) != 0 {
		t.Fatalf("got %v, stored %d", err, len(store.list))
	}
	if list, err := s.Channels(admin); err != nil || len(list) != 0 {
		t.Fatalf("channels %v %v", list, err)
	}
}
