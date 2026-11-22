-- Meter readings (#689). Personal content (ADR 0018): new meters are shared
-- with everyone, since meters belong to the household. A reading belongs to
-- its meter and has no owner of its own.
CREATE TABLE meters_items (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    kind       TEXT NOT NULL,
    unit       TEXT NOT NULL,
    owner_id   TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'everyone',
    created_at TEXT NOT NULL
);
CREATE INDEX meters_items_owner ON meters_items (owner_id);

CREATE TABLE meters_readings (
    meter_id TEXT NOT NULL REFERENCES meters_items (id) ON DELETE CASCADE,
    day      TEXT NOT NULL,  -- YYYY-MM-DD, one reading per day
    value    REAL NOT NULL,
    PRIMARY KEY (meter_id, day)
);
