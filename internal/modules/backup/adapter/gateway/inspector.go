package gateway

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"waylight/internal/modules/backup/domain"
)

// Inspector implements usecase.BackupInspector. It attaches the candidate
// file to a dedicated connection of the live database, so no second driver
// connection is opened. The file is attached read-only and immutable:
// SQLite neither writes to it nor creates journal files next to it, which
// matters for stored backups that are inspected in place.
type Inspector struct{ db *sql.DB }

// NewInspector returns an inspector using db.
func NewInspector(db *sql.DB) *Inspector { return &Inspector{db: db} }

// Inspect implements usecase.BackupInspector.
func (i *Inspector) Inspect(ctx context.Context, path string) (domain.BackupInfo, error) {
	conn, err := i.db.Conn(ctx)
	if err != nil {
		return domain.BackupInfo{}, fmt.Errorf("get connection: %w", err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, `ATTACH DATABASE ? AS restore_check`, readOnlyURI(path)); err != nil {
		return domain.BackupInfo{}, fmt.Errorf("attach backup: %w", err)
	}
	defer func() { _, _ = conn.ExecContext(context.WithoutCancel(ctx), `DETACH DATABASE restore_check`) }()
	var info domain.BackupInfo
	// The first row is "ok" for a sound database.
	if err := conn.QueryRowContext(ctx, `PRAGMA restore_check.integrity_check`).Scan(&info.Integrity); err != nil {
		return domain.BackupInfo{}, fmt.Errorf("integrity check: %w", err)
	}
	var tables int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM restore_check.sqlite_master WHERE type = 'table' AND name = 'schema_migrations'`).Scan(&tables); err != nil {
		return domain.BackupInfo{}, fmt.Errorf("read schema: %w", err)
	}
	if tables == 0 {
		return info, nil
	}
	if info.Migrations, err = migrations(ctx, conn, `SELECT module, version FROM restore_check.schema_migrations`); err != nil {
		return domain.BackupInfo{}, err
	}
	if info.Tabs, err = countRows(ctx, conn, "dashboard_tabs"); err != nil {
		return domain.BackupInfo{}, err
	}
	if info.Widgets, err = countRows(ctx, conn, "dashboard_widgets"); err != nil {
		return domain.BackupInfo{}, err
	}
	return info, nil
}

// readOnlyURI turns a file path into an SQLite URI that opens it read-only
// and immutable.
func readOnlyURI(path string) string {
	return (&url.URL{Scheme: "file", Path: path, RawQuery: "mode=ro&immutable=1"}).String()
}

// countRows counts the rows of an attached table; zero when the backup has
// no such table. table is one of the constant names above, never input.
func countRows(ctx context.Context, conn *sql.Conn, table string) (int, error) {
	var exists int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM restore_check.sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&exists); err != nil {
		return 0, fmt.Errorf("read schema: %w", err)
	}
	if exists == 0 {
		return 0, nil
	}
	var n int
	if err := conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM restore_check.`+table).Scan(&n); err != nil {
		return 0, fmt.Errorf("count %s: %w", table, err)
	}
	return n, nil
}

// Current implements usecase.BackupInspector.
func (i *Inspector) Current(ctx context.Context) ([]domain.Migration, error) {
	conn, err := i.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}
	defer conn.Close()
	return migrations(ctx, conn, `SELECT module, version FROM main.schema_migrations`)
}

func migrations(ctx context.Context, conn *sql.Conn, query string) ([]domain.Migration, error) {
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	defer rows.Close()
	var out []domain.Migration
	for rows.Next() {
		var m domain.Migration
		if err := rows.Scan(&m.Module, &m.Version); err != nil {
			return nil, fmt.Errorf("scan migration: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
