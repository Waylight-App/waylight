-- The audit log (#303): who changed what, when; never the content. Entries
-- are only appended and aged out by the retention job.
CREATE TABLE audit_entries (
    id         TEXT PRIMARY KEY,
    at         TEXT NOT NULL,
    account_id TEXT NOT NULL,
    role       TEXT NOT NULL,
    module     TEXT NOT NULL,
    route      TEXT NOT NULL,
    object_id  TEXT NOT NULL DEFAULT ''
);
CREATE INDEX audit_entries_at ON audit_entries (at);
CREATE TABLE audit_settings (
    id             INTEGER PRIMARY KEY CHECK (id = 1),
    retention_days INTEGER NOT NULL
);
