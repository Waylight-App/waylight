-- Login names of directory users and the members they sign in as (#396).
CREATE TABLE ldap_links (
    name            TEXT PRIMARY KEY,
    account_id      TEXT NOT NULL,
    groups          TEXT NOT NULL DEFAULT '[]', -- directory groups synced last time
    disabled_by_dir INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX ldap_links_account ON ldap_links (account_id);
