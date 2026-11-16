-- Connected accounts (#326). Tokens and client secrets are encrypted per
-- value with the instance's SECRET_KEY; nothing here is readable as is.
CREATE TABLE connection_clients (
    service    TEXT PRIMARY KEY,
    client_id  TEXT NOT NULL,
    secret_enc TEXT NOT NULL DEFAULT '',
    tenant     TEXT NOT NULL DEFAULT ''
);
CREATE TABLE connection_tokens (
    account_id   TEXT NOT NULL,
    service      TEXT NOT NULL,
    access_enc   TEXT NOT NULL,
    refresh_enc  TEXT NOT NULL DEFAULT '',
    expiry       TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL,
    connected_at TEXT NOT NULL,
    PRIMARY KEY (account_id, service)
);
