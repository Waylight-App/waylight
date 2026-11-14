-- Birthdays (#287): entered by hand, imported from vCard or synced from CardDAV.
CREATE TABLE birthdays (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    month      INTEGER NOT NULL,
    day        INTEGER NOT NULL,
    year       INTEGER NOT NULL DEFAULT 0,
    source     TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- One row: reminder and the optional CardDAV address book.
CREATE TABLE birthday_settings (
    id                 INTEGER PRIMARY KEY CHECK (id = 1),
    remind_enabled     INTEGER NOT NULL DEFAULT 0,
    remind_at          TEXT NOT NULL DEFAULT '08:00',
    last_reminded      TEXT NOT NULL DEFAULT '',
    carddav_url        TEXT NOT NULL DEFAULT '',
    username           TEXT NOT NULL DEFAULT '',
    password_encrypted TEXT NOT NULL DEFAULT ''
);
