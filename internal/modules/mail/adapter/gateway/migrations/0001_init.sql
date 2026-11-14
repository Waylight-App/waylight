-- IMAP accounts of the mail widget (#283). The password is encrypted with
-- SECRET_KEY; the table is not part of data exports.
CREATE TABLE mail_accounts (
    id                 TEXT PRIMARY KEY,
    name               TEXT NOT NULL,
    host               TEXT NOT NULL,
    port               INTEGER NOT NULL,
    username           TEXT NOT NULL,
    password_encrypted TEXT NOT NULL,
    folder             TEXT NOT NULL,
    webmail_url        TEXT NOT NULL DEFAULT '',
    created_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
