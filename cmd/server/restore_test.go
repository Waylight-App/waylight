package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"waylight/internal/platform/config"
	"waylight/internal/platform/database"
)

func (c *client) restore(fields map[string]string, file []byte, wantStatus int) []byte {
	c.t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for _, k := range []string{"password", "code", "confirm", "source", "name"} {
		if v, ok := fields[k]; ok {
			_ = mw.WriteField(k, v)
		}
	}
	if file != nil {
		fw, _ := mw.CreateFormFile("file", "backup.db")
		_, _ = fw.Write(file)
	}
	_ = mw.Close()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, c.base+"/api/backups/restore", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("X-CSRF-Token", c.csrf)
	res, err := c.http.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	if res.StatusCode != wantStatus {
		c.t.Fatalf("restore: status %d, want %d: %s", res.StatusCode, wantStatus, out)
	}
	return out
}

// TestRestoreEndToEnd runs the guided restore (#46, #677) against real
// SQLite files: the preview, the checks that leave everything untouched,
// the safety copy and the swap the next start performs.
func TestRestoreEndToEnd(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "restore.db")
	backupDir := filepath.Join(t.TempDir(), "backups")
	srv := newTestServer(t, "http://127.0.0.1:1", func(c *config.Config) { c.DatabasePath, c.BackupDir = dbPath, backupDir })
	c := newClient(t, srv.URL)
	var session obj
	c.json(http.MethodPost, "/api/auth/setup", obj{"setupToken": "e2e-setup-token-0123", "password": "e2e-password-123"}, http.StatusOK, &session)
	c.csrf, _ = session["csrfToken"].(string)
	c.do(http.MethodPost, "/api/notes", obj{"title": "Before the backup", "body": "kept"}, http.StatusCreated)
	var backupFile obj
	c.json(http.MethodPost, "/api/backups", nil, http.StatusCreated, &backupFile)
	name := backupFile["name"].(string)
	// Written after the backup: gone after the restore, kept in the safety copy.
	c.do(http.MethodPost, "/api/notes", obj{"title": "After the backup", "body": "replaced"}, http.StatusCreated)

	// The preview inspects the stored backup without changing it.
	var preview obj
	c.json(http.MethodGet, "/api/backups/"+name+"/preview", nil, http.StatusOK, &preview)
	if preview["restorable"] != true || preview["problem"] != nil || preview["updates"] != float64(0) || preview["modules"].(float64) < 2 {
		t.Fatalf("preview: %v", preview)
	}
	c.do(http.MethodGet, "/api/backups/waylight-20000101-000000.db/preview", nil, http.StatusNotFound)

	ok := map[string]string{"password": "e2e-password-123", "confirm": "replace", "source": "local", "name": name}
	with := func(k, v string) map[string]string {
		m := map[string]string{}
		for key, val := range ok {
			m[key] = val
		}
		m[k] = v
		return m
	}
	c.restore(with("confirm", ""), nil, http.StatusUnprocessableEntity)
	c.restore(with("password", "wrong-password"), nil, http.StatusUnauthorized)
	c.restore(with("source", "upload"), []byte("not a database, just text that is long enough"), http.StatusUnprocessableEntity)
	if _, err := os.Stat(dbPath + ".restore"); err == nil {
		t.Fatal("rejected restore must not be scheduled")
	}
	var list []obj
	c.json(http.MethodGet, "/api/backups", nil, http.StatusOK, &list)
	if len(list) != 1 {
		t.Fatalf("a rejected restore made a safety copy: %v", list)
	}

	// Drain a restart signal left by other tests.
	select {
	case <-restartRequests:
	default:
	}
	var restored obj
	if err := json.Unmarshal(c.restore(ok, nil, http.StatusAccepted), &restored); err != nil {
		t.Fatal(err)
	}
	safety, _ := restored["safetyCopy"].(map[string]any)
	safetyName, _ := safety["name"].(string)
	if restored["status"] != "restarting" || safetyName == "" || safetyName == name {
		t.Fatalf("restore answer: %v", restored)
	}
	select {
	case <-restartRequests:
	default:
		t.Fatal("no restart requested")
	}
	entries, _ := os.ReadDir(filepath.Dir(dbPath))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".restore-upload-") {
			t.Fatalf("staging file left: %s", e.Name())
		}
	}
	c.json(http.MethodGet, "/api/backups", nil, http.StatusOK, &list)
	if len(list) != 2 || list[0]["name"] != safetyName {
		t.Fatalf("the safety copy is not the newest backup: %v", list)
	}

	// What the next start does, on copies so the running server keeps its files.
	notes := func(path string) []string {
		t.Helper()
		db, err := database.Open(ctx, path)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		rows, err := db.QueryContext(ctx, `SELECT title FROM notes ORDER BY title`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var out []string
		for rows.Next() {
			var title string
			_ = rows.Scan(&title)
			out = append(out, title)
		}
		return out
	}
	next := filepath.Join(t.TempDir(), "waylight.db")
	copyFile(t, dbPath+".restore", next+".restore")
	copyFile(t, filepath.Join(backupDir, safetyName), next)
	if applied, err := database.ApplyPendingRestore(next); err != nil || !applied {
		t.Fatalf("apply: %v %v", applied, err)
	}
	if got := strings.Join(notes(next), ","); got != "Before the backup" {
		t.Fatalf("restored notes: %s", got)
	}
	if got := strings.Join(notes(next+".before-restore"), ","); got != "After the backup,Before the backup" {
		t.Fatalf("notes kept from before the restore: %s", got)
	}
	if got := strings.Join(notes(filepath.Join(backupDir, safetyName)), ","); got != "After the backup,Before the backup" {
		t.Fatalf("notes in the safety copy: %s", got)
	}
}

func copyFile(t *testing.T, from, to string) {
	t.Helper()
	b, err := os.ReadFile(from)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(to, b, 0o600); err != nil {
		t.Fatal(err)
	}
}
