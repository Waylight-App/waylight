-- Members and groups (#294, ADR 0018). SQLite cannot change a CHECK
-- constraint, so auth_accounts is rebuilt with the same columns, as in 0008.
CREATE TABLE auth_accounts_v3 (
    id               TEXT PRIMARY KEY,
    role             TEXT NOT NULL CHECK (role IN ('owner', 'guest', 'kiosk', 'member')),
    name             TEXT NOT NULL UNIQUE,
    password_hash    TEXT NOT NULL,
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL,
    totp_secret_enc  TEXT NOT NULL DEFAULT '',
    totp_pending_enc TEXT NOT NULL DEFAULT '',
    totp_last_step   INTEGER NOT NULL DEFAULT 0,
    disabled         INTEGER NOT NULL DEFAULT 0,
    tabs             TEXT NOT NULL DEFAULT '[]'
);
INSERT INTO auth_accounts_v3 (id, role, name, password_hash, created_at, updated_at, totp_secret_enc, totp_pending_enc, totp_last_step, disabled, tabs)
SELECT id, role, name, password_hash, created_at, updated_at, totp_secret_enc, totp_pending_enc, totp_last_step, disabled, tabs FROM auth_accounts;
DROP TABLE auth_accounts;
ALTER TABLE auth_accounts_v3 RENAME TO auth_accounts;
CREATE UNIQUE INDEX auth_accounts_one_owner ON auth_accounts (role) WHERE role = 'owner';

-- Groups that records can be shared with ("group:<id>").
CREATE TABLE auth_groups (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE COLLATE NOCASE,
    created_at TEXT NOT NULL
);

CREATE TABLE auth_group_accounts (
    group_id   TEXT NOT NULL,
    account_id TEXT NOT NULL,
    PRIMARY KEY (group_id, account_id)
);
CREATE INDEX auth_group_accounts_account ON auth_group_accounts (account_id);
