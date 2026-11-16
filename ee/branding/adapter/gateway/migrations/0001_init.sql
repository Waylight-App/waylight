-- Branding (#305): one row of settings and the uploaded images, re-encoded
-- as PNG. Kept when the licence expires; applied again with a valid key.
CREATE TABLE branding_settings (
    id              INTEGER PRIMARY KEY CHECK (id = 1),
    name            TEXT NOT NULL DEFAULT '',
    accent          TEXT NOT NULL DEFAULT '',
    secondary       TEXT NOT NULL DEFAULT '',
    hide_powered_by INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE branding_assets (
    kind         TEXT PRIMARY KEY CHECK (kind IN ('logo', 'logo-dark', 'favicon')),
    data         BLOB NOT NULL,
    content_type TEXT NOT NULL,
    updated_at   TEXT NOT NULL
);
