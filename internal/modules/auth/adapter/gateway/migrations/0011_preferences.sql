-- Per-account preferences of the auth module (ADR 0018): whether an
-- account's own new records start private. No row means the default.
CREATE TABLE auth_preferences (
    account_id         TEXT PRIMARY KEY,
    private_by_default INTEGER NOT NULL DEFAULT 0
);
