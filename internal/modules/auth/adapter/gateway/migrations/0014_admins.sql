-- Several admins (#434): the role 'admin' next to the one owner, and where a
-- role came from ('' by hand, 'sso' or 'directory'), so single sign-on only
-- takes back what it gave. SQLite cannot change a CHECK constraint, so
-- auth_accounts is rebuilt with the same columns, as in 0008 and 0010.
CREATE TABLE auth_accounts_v4 (
    id               TEXT PRIMARY KEY,
    role             TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'guest', 'kiosk', 'member')),
    name             TEXT NOT NULL UNIQUE,
    password_hash    TEXT NOT NULL,
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL,
    totp_secret_enc  TEXT NOT NULL DEFAULT '',
    totp_pending_enc TEXT NOT NULL DEFAULT '',
    totp_last_step   INTEGER NOT NULL DEFAULT 0,
    disabled         INTEGER NOT NULL DEFAULT 0,
    tabs             TEXT NOT NULL DEFAULT '[]',
    role_source      TEXT NOT NULL DEFAULT ''
);
INSERT INTO auth_accounts_v4 (id, role, name, password_hash, created_at, updated_at, totp_secret_enc, totp_pending_enc, totp_last_step, disabled, tabs)
SELECT id, role, name, password_hash, created_at, updated_at, totp_secret_enc, totp_pending_enc, totp_last_step, disabled, tabs FROM auth_accounts;
DROP TABLE auth_accounts;
ALTER TABLE auth_accounts_v4 RENAME TO auth_accounts;
CREATE UNIQUE INDEX auth_accounts_one_owner ON auth_accounts (role) WHERE role = 'owner';
