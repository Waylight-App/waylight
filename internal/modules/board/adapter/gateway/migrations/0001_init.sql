-- Family board (#685). Notes are personal content (ADR 0018), shared with
-- everyone unless their writer narrows it.
CREATE TABLE board_notes (
    id         TEXT PRIMARY KEY,
    text       TEXT NOT NULL,
    pinned     INTEGER NOT NULL DEFAULT 0,
    expires_on TEXT NOT NULL DEFAULT '',  -- YYYY-MM-DD, '' = never
    owner_id   TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'everyone',
    created_at TEXT NOT NULL
);
CREATE INDEX board_notes_owner ON board_notes (owner_id);
