-- Push devices of the mobile app (#564, ADR 0025): one per app session.
CREATE TABLE notification_app_devices (
    session     TEXT PRIMARY KEY,      -- public ID of the app session
    account_id  TEXT NOT NULL,
    name        TEXT NOT NULL,
    platform    TEXT NOT NULL,         -- fcm or apns
    token_enc   TEXT NOT NULL,         -- encrypted push token
    public_key  BLOB NOT NULL,         -- X25519, 32 bytes
    updated_at  TEXT NOT NULL
);

-- The instance's Ed25519 key for the push relay, created on first use.
CREATE TABLE notification_push_key (
    id       INTEGER PRIMARY KEY CHECK (id = 1),
    seed_enc TEXT NOT NULL
);
