-- Connected wikis (#321). The API token is encrypted with SECRET_KEY; page
-- content is never stored, the search asks the wikis directly.
CREATE TABLE wiki_connections (
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    kind            TEXT NOT NULL,
    url             TEXT NOT NULL,
    token_encrypted TEXT NOT NULL,
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
