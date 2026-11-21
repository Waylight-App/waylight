-- Places entered once and chosen by widgets (#591). They belong to the
-- account that saved them and are visible to everyone by default, like
-- other personal content (ADR 0018).
CREATE TABLE places_places (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    label      TEXT NOT NULL DEFAULT '',
    lat        REAL NOT NULL,
    lon        REAL NOT NULL,
    created_at TEXT NOT NULL,
    owner_id   TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'everyone'
);
CREATE INDEX places_places_owner ON places_places (owner_id);
