-- An optional e-mail address per account, set by the account itself, for
-- password reset links (#296). Kept apart from auth_accounts so the account
-- table does not need another rebuild.
CREATE TABLE auth_account_emails (
    account_id TEXT PRIMARY KEY,
    email      TEXT NOT NULL
);
