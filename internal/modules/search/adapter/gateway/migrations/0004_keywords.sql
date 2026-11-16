-- Own keywords per account (#478): "nas" opens the NAS tile. Personal like
-- the shortcuts (ADR 0018); target is an id, a web address or a search.
CREATE TABLE search_keywords (
    account_id TEXT NOT NULL DEFAULT 'owner',
    keyword    TEXT NOT NULL,
    kind       TEXT NOT NULL,
    target     TEXT NOT NULL,
    label      TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (account_id, keyword)
);
