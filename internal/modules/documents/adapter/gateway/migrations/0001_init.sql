-- Warranty and documents (#688). Personal content (ADR 0018): private for
-- members and admins by default, shareable like a note. The attached files
-- are kept apart so that listing never reads the blobs.
CREATE TABLE documents_items (
    id           TEXT PRIMARY KEY,
    title        TEXT NOT NULL,
    purchased_on TEXT NOT NULL DEFAULT '',  -- YYYY-MM-DD or empty
    warranty_end TEXT NOT NULL DEFAULT '',  -- YYYY-MM-DD or empty
    note         TEXT NOT NULL DEFAULT '',
    file_name    TEXT NOT NULL DEFAULT '',  -- empty: no file attached
    content_type TEXT NOT NULL DEFAULT '',
    size         INTEGER NOT NULL DEFAULT 0,
    owner_id     TEXT NOT NULL,
    visibility   TEXT NOT NULL DEFAULT 'private',
    created_at   TEXT NOT NULL
);
CREATE INDEX documents_items_owner ON documents_items (owner_id);

CREATE TABLE documents_files (
    item_id TEXT PRIMARY KEY REFERENCES documents_items (id) ON DELETE CASCADE,
    data    BLOB NOT NULL
);
