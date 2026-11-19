package gateway

import (
	"bytes"
	"context"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"waylight/internal/modules/notifications/domain"
	"waylight/internal/modules/notifications/usecase"
	"waylight/internal/shared/pushproto"
)

// AppDeviceRepository implements usecase.AppDeviceRepository; push tokens
// are encrypted at rest like the channel secrets.
type AppDeviceRepository struct {
	db     *sql.DB
	cipher Cipher
}

// NewAppDeviceRepository returns the repository.
func NewAppDeviceRepository(db *sql.DB, cipher Cipher) *AppDeviceRepository {
	return &AppDeviceRepository{db: db, cipher: cipher}
}

// List returns all devices, most recently registered first.
func (r *AppDeviceRepository) List(ctx context.Context) ([]domain.AppDevice, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT session, account_id, name, platform, token_enc, public_key, updated_at
		FROM notification_app_devices ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query devices: %w", err)
	}
	defer rows.Close()
	var out []domain.AppDevice
	for rows.Next() {
		var d domain.AppDevice
		var platform, token, updated string
		if err := rows.Scan(&d.Session, &d.AccountID, &d.Name, &platform, &token, &d.PublicKey, &updated); err != nil {
			return nil, fmt.Errorf("scan device: %w", err)
		}
		if d.Token, err = r.cipher.Decrypt(token); err != nil {
			return nil, fmt.Errorf("decrypt push token (was SECRET_KEY changed?): %w", err)
		}
		d.Platform = domain.PushPlatform(platform)
		d.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, d)
	}
	return out, rows.Err()
}

// Save inserts or replaces the device of its session.
func (r *AppDeviceRepository) Save(ctx context.Context, d domain.AppDevice) error {
	token, err := r.cipher.Encrypt(d.Token)
	if err != nil {
		return fmt.Errorf("encrypt push token: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO notification_app_devices
		(session, account_id, name, platform, token_enc, public_key, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (session) DO UPDATE SET account_id = excluded.account_id, name = excluded.name,
		platform = excluded.platform, token_enc = excluded.token_enc, public_key = excluded.public_key,
		updated_at = excluded.updated_at`,
		d.Session, d.AccountID, d.Name, string(d.Platform), token, d.PublicKey, d.UpdatedAt.UTC().Format(time.RFC3339))
	return err
}

// Delete removes the device of a session; unknown sessions are no error.
func (r *AppDeviceRepository) Delete(ctx context.Context, session string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM notification_app_devices WHERE session = ?`, session)
	return err
}

// RelayClient implements usecase.PushRelay: it seals each message for the
// device and posts it, signed with the instance key, to the push relay.
type RelayClient struct {
	client *http.Client
	url    string
	db     *sql.DB
	cipher Cipher
	now    func() time.Time

	mu  sync.Mutex
	key ed25519.PrivateKey
}

// NewRelayClient returns a client for the relay at base (without /v1/push).
func NewRelayClient(client *http.Client, base string, db *sql.DB, cipher Cipher, now func() time.Time) *RelayClient {
	return &RelayClient{client: client, url: strings.TrimRight(base, "/") + "/v1/push", db: db, cipher: cipher, now: now}
}

// instanceKey loads the Ed25519 key or creates it on first use; the relay
// knows the instance only by this key (ADR 0025).
func (c *RelayClient) instanceKey(ctx context.Context) (ed25519.PrivateKey, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.key != nil {
		return c.key, nil
	}
	var enc string
	err := c.db.QueryRowContext(ctx, `SELECT seed_enc FROM notification_push_key WHERE id = 1`).Scan(&enc)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		seed := make([]byte, ed25519.SeedSize)
		if _, err := rand.Read(seed); err != nil {
			return nil, err
		}
		if enc, err = c.cipher.Encrypt(base64.StdEncoding.EncodeToString(seed)); err != nil {
			return nil, fmt.Errorf("encrypt push key: %w", err)
		}
		// INSERT OR IGNORE plus a read keeps one key if two deliveries race.
		if _, err := c.db.ExecContext(ctx, `INSERT OR IGNORE INTO notification_push_key (id, seed_enc) VALUES (1, ?)`, enc); err != nil {
			return nil, fmt.Errorf("store push key: %w", err)
		}
		if err := c.db.QueryRowContext(ctx, `SELECT seed_enc FROM notification_push_key WHERE id = 1`).Scan(&enc); err != nil {
			return nil, err
		}
	case err != nil:
		return nil, fmt.Errorf("load push key: %w", err)
	}
	plain, err := c.cipher.Decrypt(enc)
	if err != nil {
		return nil, fmt.Errorf("decrypt push key (was SECRET_KEY changed?): %w", err)
	}
	seed, err := base64.StdEncoding.DecodeString(plain)
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, errors.New("stored push key is invalid")
	}
	c.key = ed25519.NewKeyFromSeed(seed)
	return c.key, nil
}

// pushContent is what the app shows; only the device can read it.
type pushContent struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Link  string `json:"link,omitempty"`
}

// maxBody keeps the sealed content below pushproto.MaxPayload: the
// envelope adds 61 bytes, the JSON a few more.
const maxBody = 2400

// Push implements usecase.PushRelay.
func (c *RelayClient) Push(ctx context.Context, d domain.AppDevice, m domain.Message) error {
	key, err := c.instanceKey(ctx)
	if err != nil {
		return err
	}
	pub, err := ecdh.X25519().NewPublicKey(d.PublicKey)
	if err != nil {
		return fmt.Errorf("device key: %w", err)
	}
	body := m.Body
	if len(body) > maxBody {
		body = body[:maxBody] + "…"
	}
	plain, _ := json.Marshal(pushContent{Title: m.Title, Body: body}) // plain struct: cannot fail
	sealed, err := pushproto.Seal(pub, plain)
	if err != nil {
		return err
	}
	req, _ := json.Marshal(map[string]string{ // plain map: cannot fail
		"platform": string(d.Platform), "token": d.Token, "payload": base64.StdEncoding.EncodeToString(sealed),
	})
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(req))
	if err != nil {
		return err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", pushproto.Authorization(key, req, c.now()))
	res, err := c.client.Do(r)
	if err != nil {
		return fmt.Errorf("push relay: %w", err)
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 1<<12))
	switch res.StatusCode {
	case http.StatusAccepted, http.StatusOK:
		return nil
	case http.StatusGone, http.StatusForbidden:
		// 403: the token is bound to another instance, so this one can
		// never reach the device; drop it like a gone token.
		return usecase.ErrDeviceGone
	}
	return fmt.Errorf("push relay answered %d", res.StatusCode)
}
