-- Sign-ins at the OpenID Connect provider linked to an account (#527). One
-- identity (issuer + subject) belongs to one account at most. Links of a
-- deleted account are removed by the account deletion, not by a foreign key:
-- auth_accounts has been rebuilt by migrations before.
CREATE TABLE auth_identities (
    issuer     TEXT NOT NULL,
    subject    TEXT NOT NULL,
    account_id TEXT NOT NULL,
    email      TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    PRIMARY KEY (issuer, subject)
);
CREATE INDEX auth_identities_account ON auth_identities (account_id);
