-- Google and Microsoft accounts (#282). The client secret and both tokens
-- are encrypted with SECRET_KEY; the calendars are a small JSON list.
CREATE TABLE calendar_accounts (
    id                TEXT PRIMARY KEY,
    provider          TEXT NOT NULL,
    client_id         TEXT NOT NULL,
    name              TEXT NOT NULL,
    secret_encrypted  TEXT NOT NULL,
    access_encrypted  TEXT NOT NULL,
    refresh_encrypted TEXT NOT NULL,
    expiry            TEXT NOT NULL,
    calendars         TEXT NOT NULL,
    created_at        TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
