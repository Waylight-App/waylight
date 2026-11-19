package gateway

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"waylight/internal/modules/notifications/domain"
	"waylight/internal/modules/notifications/usecase"
	"waylight/internal/platform/database/dbtest"
	"waylight/internal/shared/pushproto"
)

func TestAppDeviceRepository(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "notifications", Migrations())
	repo := NewAppDeviceRepository(db, prefixCipher{})
	d := domain.AppDevice{Session: "s", AccountID: "a", Name: "Pixel", Platform: domain.PushFCM, Token: "tok",
		PublicKey: make([]byte, 32), UpdatedAt: time.Unix(1_800_000_000, 0).UTC()}
	if err := repo.Save(ctx, d); err != nil {
		t.Fatal(err)
	}
	var stored string
	_ = db.QueryRow(`SELECT token_enc FROM notification_app_devices`).Scan(&stored)
	if stored != "enc:tok" {
		t.Fatal("token must be encrypted:", stored)
	}
	d.Token = "new"
	_ = repo.Save(ctx, d)
	list, err := repo.List(ctx)
	if err != nil || len(list) != 1 || list[0].Token != "new" || !list[0].UpdatedAt.Equal(d.UpdatedAt) {
		t.Fatal(list, err)
	}
	if err := repo.Delete(ctx, "s"); err != nil {
		t.Fatal(err)
	}
	if list, _ := repo.List(ctx); len(list) != 0 {
		t.Fatal("not deleted")
	}
}

func TestRelayClient(t *testing.T) {
	ctx := context.Background()
	db := dbtest.Open(t, "notifications", Migrations())
	device, _ := ecdh.X25519().GenerateKey(rand.Reader)
	now := time.Unix(1_800_000_000, 0)
	var keys []string
	var content pushContent
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		pub, err := pushproto.Verify(r.Header.Get("Authorization"), body, now)
		if err != nil || r.URL.Path != "/v1/push" {
			t.Error("bad request:", r.URL.Path, err)
		}
		keys = append(keys, pushproto.KeyID(pub))
		var req struct{ Platform, Token, Payload string }
		_ = json.Unmarshal(body, &req)
		sealed, _ := base64.StdEncoding.DecodeString(req.Payload)
		plain, err := pushproto.Open(device, sealed)
		if err != nil {
			t.Error("cannot open payload:", err)
		}
		_ = json.Unmarshal(plain, &content)
		switch req.Token {
		case "gone":
			w.WriteHeader(http.StatusGone)
		case "busy":
			w.WriteHeader(http.StatusTooManyRequests)
		default:
			w.WriteHeader(http.StatusAccepted)
		}
	}))
	defer srv.Close()
	c := NewRelayClient(srv.Client(), srv.URL+"/", db, prefixCipher{}, func() time.Time { return now })
	d := domain.AppDevice{Platform: domain.PushFCM, Token: "tok", PublicKey: device.PublicKey().Bytes()}
	msg := domain.Message{Title: "Backup failed", Body: strings.Repeat("x", 3000)}
	if err := c.Push(ctx, d, msg); err != nil {
		t.Fatal(err)
	}
	if content.Title != "Backup failed" || len(content.Body) > maxBody+len("…") {
		t.Fatal("content:", content.Title, len(content.Body))
	}
	d.Token = "gone"
	if err := c.Push(ctx, d, msg); !errors.Is(err, usecase.ErrDeviceGone) {
		t.Fatal(err)
	}
	d.Token = "busy"
	if err := c.Push(ctx, d, msg); err == nil || errors.Is(err, usecase.ErrDeviceGone) {
		t.Fatal(err)
	}
	// A new client on the same database signs with the same stored key.
	c2 := NewRelayClient(srv.Client(), srv.URL, db, prefixCipher{}, func() time.Time { return now })
	d.Token = "tok"
	_ = c2.Push(ctx, d, msg)
	if keys[0] != keys[len(keys)-1] {
		t.Fatal("instance key changed")
	}
}
