-- Invitation and password reset links (#296). Only the token hash is stored.
CREATE TABLE auth_account_links (
    id         TEXT PRIMARY KEY,
    token_hash TEXT NOT NULL UNIQUE,
    purpose    TEXT NOT NULL CHECK (purpose IN ('invite', 'reset')),
    account_id TEXT NOT NULL DEFAULT '',
    name       TEXT NOT NULL,
    role       TEXT NOT NULL,
    tabs       TEXT NOT NULL DEFAULT '[]',
    created_at TEXT NOT NULL,
    expires_at TEXT NOT NULL
);
CREATE INDEX auth_account_links_account ON auth_account_links (account_id);
