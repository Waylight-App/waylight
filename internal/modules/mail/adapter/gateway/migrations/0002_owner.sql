-- Mail accounts belong to one account (ADR 0018, #294). Existing ones are
-- the owner's. Members never see the owner's mailboxes; guests and kiosk
-- displays keep seeing the owner's, as before.
ALTER TABLE mail_accounts ADD COLUMN owner_id TEXT NOT NULL DEFAULT 'owner';
ALTER TABLE mail_accounts ADD COLUMN visibility TEXT NOT NULL DEFAULT 'private';
CREATE INDEX mail_accounts_owner ON mail_accounts (owner_id);
