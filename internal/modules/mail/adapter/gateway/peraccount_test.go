package gateway

import (
	"context"
	"errors"
	"testing"

	"trisse-home/internal/modules/mail/domain"
	"trisse-home/internal/platform/database/dbtest"
	"trisse-home/internal/shared/kernel"
)

// TestMailboxesPerAccount: members see only their own mailboxes, never the
// admin's; guests keep the admin's (ADR 0018).
func TestMailboxesPerAccount(t *testing.T) {
	repo := NewRepository(dbtest.Open(t, "mail", Migrations()), prefixCipher{})
	admin := kernel.WithPrincipal(context.Background(), kernel.AdminPrincipal())
	anna := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "anna", Role: "member"})
	ben := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "ben", Role: "member"})
	guest := kernel.WithPrincipal(context.Background(), kernel.Principal{AccountID: "g1", Role: "guest"})
	for ctx, id := range map[context.Context]string{admin: "m-admin", anna: "m-anna"} {
		if err := repo.Save(ctx, domain.Account{ID: id, Name: id, Host: "imap.example.org", Port: 993, Username: id, Password: "pw", Folder: "INBOX"}); err != nil {
			t.Fatal(err)
		}
	}
	count := func(ctx context.Context) int {
		list, err := repo.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		return len(list)
	}
	if count(anna) != 1 || count(ben) != 0 || count(guest) != 1 || count(admin) != 1 {
		t.Fatalf("anna %d, ben %d, guest %d, admin %d", count(anna), count(ben), count(guest), count(admin))
	}
	if _, err := repo.Get(ben, "m-admin"); !errors.Is(err, domain.ErrAccountNotFound) {
		t.Fatalf("ben opened the admin's mailbox: %v", err)
	}
	if err := repo.Save(ben, domain.Account{ID: "m-admin", Name: "x", Host: "evil.example", Port: 993, Username: "x", Password: "x", Folder: "INBOX"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := repo.Get(admin, "m-admin"); got.Host != "imap.example.org" {
		t.Fatalf("ben redirected the admin's mailbox to %q", got.Host)
	}
	_ = repo.Delete(ben, "m-admin")
	if count(admin) != 1 {
		t.Fatal("ben deleted the admin's mailbox")
	}
}
