-- Connected external to-do lists (#169). The password is encrypted with
-- SECRET_KEY; the table is not part of data exports.
CREATE TABLE todos_sources (
    id                 TEXT PRIMARY KEY,
    name               TEXT NOT NULL,
    kind               TEXT NOT NULL,
    url                TEXT NOT NULL,
    username           TEXT NOT NULL,
    password_encrypted TEXT NOT NULL,
    created_at         TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
