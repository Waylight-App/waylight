package integrations

import (
	"context"
	"errors"
	"testing"
	"time"

	"waylight/internal/modules"
	"waylight/internal/modules/integrations/adapter/gateway"
	"waylight/internal/modules/integrations/domain"
	"waylight/internal/platform/database/dbtest"
)

type plainCipher struct{}

func (plainCipher) Encrypt(s string) (string, error) { return s, nil }
func (plainCipher) Decrypt(s string) (string, error) { return s, nil }

type immichModule struct{ modules.Base }

func (immichModule) Name() string               { return "immich" }
func (immichModule) DataKind() modules.DataKind { return modules.InstanceData }
func (immichModule) ConnectionSpecs() []modules.ConnectionSpec {
	return []modules.ConnectionSpec{{Service: "immich", Title: "Immich", Fields: []modules.CredentialField{{Name: "apiKey", Secret: true}}}}
}

func TestResolveConnection(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "integrations", gateway.Migrations())
	m := New(Deps{DB: db, Cipher: plainCipher{}, Modules: []modules.Module{immichModule{}}})
	repo := gateway.NewRepository(db, plainCipher{})
	now := time.Now()
	for _, c := range []struct {
		id, service string
		cred        domain.Credential
	}{{"ok", "immich", domain.Credential{"apiKey": "k"}}, {"empty", "immich", domain.Credential{}}, {"other", "grafana", domain.Credential{"token": "t"}}} {
		x, err := domain.NewConnection(c.id, c.service, c.id, "http://h", c.cred, now)
		if err != nil {
			t.Fatal(err)
		}
		if err := repo.Save(ctx, x); err != nil {
			t.Fatal(err)
		}
	}
	got, err := m.ResolveConnection(ctx, "immich", "ok")
	if err != nil || got.URL != "http://h" || got.Credential["apiKey"] != "k" {
		t.Fatalf("%+v %v", got, err)
	}
	for _, id := range []string{"empty", "other", "missing"} {
		if _, err := m.ResolveConnection(ctx, "immich", id); !errors.Is(err, modules.ErrNoConnection) {
			t.Errorf("%s: %v", id, err)
		}
	}
	var late modules.LateResolver
	if _, err := late.ResolveConnection(ctx, "immich", "ok"); !errors.Is(err, modules.ErrNoConnection) {
		t.Fatalf("unset: %v", err)
	}
	late.Set(m)
	if _, err := late.ResolveConnection(ctx, "immich", "ok"); err != nil {
		t.Fatalf("set: %v", err)
	}
}
