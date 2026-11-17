// Package gateway stores connections and app registrations, and runs OAuth
// against the known services (#326).
//
// Ring: Interface adapters (gateway).
package gateway

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"

	"waylight/ee/connections/domain"
	"waylight/internal/shared/sqlutil"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns the module's SQL migrations.
func Migrations() fs.FS {
	sub, _ := fs.Sub(migrationFiles, "migrations") // embedded directory, cannot fail
	return sub
}

// Cipher encrypts tokens and secrets (the platform's SECRET_KEY cipher).
type Cipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// Store implements usecase.Tokens and usecase.Clients with SQLite.
type Store struct {
	db     *sql.DB
	cipher Cipher
}

// NewStore returns the store.
func NewStore(db *sql.DB, cipher Cipher) *Store { return &Store{db: db, cipher: cipher} }

func (s *Store) enc(v string) (string, error) {
	if v == "" {
		return "", nil
	}
	return s.cipher.Encrypt(v)
}

func (s *Store) dec(v string) (string, error) {
	if v == "" {
		return "", nil
	}
	return s.cipher.Decrypt(v)
}

func (s *Store) scanConn(row interface{ Scan(...any) error }) (domain.Connection, error) {
	var c domain.Connection
	var access, refresh, expiry, status, connected string
	if err := row.Scan(&c.AccountID, &c.Service, &access, &refresh, &expiry, &status, &connected); err != nil {
		return domain.Connection{}, err
	}
	var err error
	if c.Token.Access, err = s.dec(access); err != nil {
		return domain.Connection{}, fmt.Errorf("decrypt access token: %w", err)
	}
	if c.Token.Refresh, err = s.dec(refresh); err != nil {
		return domain.Connection{}, fmt.Errorf("decrypt refresh token: %w", err)
	}
	if expiry != "" {
		if c.Token.Expiry, err = sqlutil.ParseTime(expiry); err != nil {
			return domain.Connection{}, err
		}
	}
	if c.ConnectedAt, err = sqlutil.ParseTime(connected); err != nil {
		return domain.Connection{}, err
	}
	c.Status = domain.Status(status)
	return c, nil
}

const connColumns = `account_id, service, access_enc, refresh_enc, expiry, status, connected_at`

// Get implements usecase.Tokens.
func (s *Store) Get(ctx context.Context, accountID, service string) (domain.Connection, bool, error) {
	c, err := s.scanConn(s.db.QueryRowContext(ctx, `SELECT `+connColumns+` FROM connection_tokens WHERE account_id = ? AND service = ?`, accountID, service))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Connection{}, false, nil
	}
	if err != nil {
		return domain.Connection{}, false, err
	}
	return c, true, nil
}

// ListFor implements usecase.Tokens.
func (s *Store) ListFor(ctx context.Context, accountID string) ([]domain.Connection, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+connColumns+` FROM connection_tokens WHERE account_id = ? ORDER BY service`, accountID)
	if err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
	}
	defer rows.Close()
	var out []domain.Connection
	for rows.Next() {
		c, err := s.scanConn(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// AccountsWith implements usecase.Tokens.
func (s *Store) AccountsWith(ctx context.Context, service string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT account_id FROM connection_tokens WHERE service = ? ORDER BY account_id`, service)
	if err != nil {
		return nil, fmt.Errorf("list connected accounts: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// Save implements usecase.Tokens.
func (s *Store) Save(ctx context.Context, c domain.Connection) error {
	access, err := s.enc(c.Token.Access)
	if err != nil {
		return fmt.Errorf("encrypt access token: %w", err)
	}
	refresh, err := s.enc(c.Token.Refresh)
	if err != nil {
		return fmt.Errorf("encrypt refresh token: %w", err)
	}
	expiry := ""
	if !c.Token.Expiry.IsZero() {
		expiry = sqlutil.FormatTime(c.Token.Expiry)
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO connection_tokens (`+connColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (account_id, service) DO UPDATE SET access_enc = excluded.access_enc, refresh_enc = excluded.refresh_enc,
		   expiry = excluded.expiry, status = excluded.status, connected_at = excluded.connected_at`,
		c.AccountID, c.Service, access, refresh, expiry, string(c.Status), sqlutil.FormatTime(c.ConnectedAt))
	if err != nil {
		return fmt.Errorf("save connection: %w", err)
	}
	return nil
}

// Delete implements usecase.Tokens.
func (s *Store) Delete(ctx context.Context, accountID, service string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM connection_tokens WHERE account_id = ? AND service = ?`, accountID, service); err != nil {
		return fmt.Errorf("delete connection: %w", err)
	}
	return nil
}

// DeleteAccount implements usecase.Tokens.
func (s *Store) DeleteAccount(ctx context.Context, accountID string) error {
	if _, err := s.db.ExecContext(ctx, `DELETE FROM connection_tokens WHERE account_id = ?`, accountID); err != nil {
		return fmt.Errorf("delete connections: %w", err)
	}
	return nil
}

// Clients returns the store as usecase.Clients (Get collides with Tokens).
func (s *Store) Clients() *ClientStore { return &ClientStore{s} }

// ClientStore implements usecase.Clients.
type ClientStore struct{ s *Store }

// Get implements usecase.Clients.
func (c *ClientStore) Get(ctx context.Context, service string) (domain.Client, bool, error) {
	var out domain.Client
	var secret string
	err := c.s.db.QueryRowContext(ctx, `SELECT service, client_id, secret_enc, tenant FROM connection_clients WHERE service = ?`, service).
		Scan(&out.Service, &out.ClientID, &secret, &out.Tenant)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Client{}, false, nil
	}
	if err != nil {
		return domain.Client{}, false, fmt.Errorf("read client: %w", err)
	}
	if out.Secret, err = c.s.dec(secret); err != nil {
		return domain.Client{}, false, fmt.Errorf("decrypt client secret: %w", err)
	}
	return out, true, nil
}

// List implements usecase.Clients.
func (c *ClientStore) List(ctx context.Context) ([]domain.Client, error) {
	services, err := c.services(ctx)
	if err != nil {
		return nil, err
	}
	var out []domain.Client
	for _, s := range services {
		cl, _, err := c.Get(ctx, s)
		if err != nil {
			return nil, err
		}
		out = append(out, cl)
	}
	return out, nil
}

func (c *ClientStore) services(ctx context.Context) ([]string, error) {
	rows, err := c.s.db.QueryContext(ctx, `SELECT service FROM connection_clients ORDER BY service`)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}
	defer rows.Close()
	var services []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, rows.Err()
}

// Save implements usecase.Clients.
func (c *ClientStore) Save(ctx context.Context, cl domain.Client) error {
	secret, err := c.s.enc(cl.Secret)
	if err != nil {
		return fmt.Errorf("encrypt client secret: %w", err)
	}
	_, err = c.s.db.ExecContext(ctx,
		`INSERT INTO connection_clients (service, client_id, secret_enc, tenant) VALUES (?, ?, ?, ?)
		 ON CONFLICT (service) DO UPDATE SET client_id = excluded.client_id, secret_enc = excluded.secret_enc, tenant = excluded.tenant`,
		cl.Service, cl.ClientID, secret, cl.Tenant)
	if err != nil {
		return fmt.Errorf("save client: %w", err)
	}
	return nil
}
