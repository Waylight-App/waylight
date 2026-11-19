-- App sessions (#559): a session with a device name is used by the mobile
-- app as bearer token; browser sessions keep an empty name.
ALTER TABLE auth_sessions ADD COLUMN device TEXT NOT NULL DEFAULT '';

-- Pairing codes hand a session to the app through a QR code. Only the hash
-- is stored; a code is deleted when it is redeemed.
CREATE TABLE auth_pairing_codes (
    code_hash  TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    expires_at TEXT NOT NULL
);
