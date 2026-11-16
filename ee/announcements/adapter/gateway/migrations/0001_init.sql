-- Announcements (#297): notices written by admins or a writer group.
CREATE TABLE announcements (
    id         TEXT PRIMARY KEY,
    title      TEXT NOT NULL,
    body       TEXT NOT NULL DEFAULT '',
    pinned     INTEGER NOT NULL DEFAULT 0,
    author_id  TEXT NOT NULL,
    expires_at TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Who read which notice; removed with the notice or the account.
CREATE TABLE announcement_reads (
    announcement_id TEXT NOT NULL REFERENCES announcements (id) ON DELETE CASCADE,
    account_id      TEXT NOT NULL,
    read_at         TEXT NOT NULL,
    PRIMARY KEY (announcement_id, account_id)
);
CREATE INDEX announcement_reads_account ON announcement_reads (account_id);

-- One row: who besides admins may write ('' = admins only).
CREATE TABLE announcement_settings (
    id           INTEGER PRIMARY KEY CHECK (id = 1),
    writer_group TEXT NOT NULL DEFAULT ''
);
