-- Provider subjects and the members they sign in as (#302).
CREATE TABLE sso_links (
    subject         TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL,
    groups          TEXT NOT NULL DEFAULT '[]', -- provider groups synced last time
    disabled_by_sso INTEGER NOT NULL DEFAULT 0,
    updated_at      TEXT NOT NULL
);
CREATE INDEX sso_links_account ON sso_links (account_id);
