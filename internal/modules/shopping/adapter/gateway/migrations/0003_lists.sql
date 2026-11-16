-- Shopping lists kept on trisse-home (#471), next to the connected CalDAV
-- list. A list belongs to its owner and is shared like a note (ADR 0018);
-- everyone who sees it may add, tick off and remove items.
CREATE TABLE shopping_lists (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    owner_id   TEXT NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'private',
    created_at TEXT NOT NULL
);
CREATE INDEX shopping_lists_owner ON shopping_lists (owner_id);

CREATE TABLE shopping_items (
    id           TEXT PRIMARY KEY,
    list_id      TEXT NOT NULL REFERENCES shopping_lists (id) ON DELETE CASCADE,
    summary      TEXT NOT NULL,
    done         INTEGER NOT NULL DEFAULT 0,
    created_at   TEXT NOT NULL,
    completed_at TEXT NOT NULL DEFAULT ''
);
CREATE INDEX shopping_items_list ON shopping_items (list_id);

-- Which list the widget shows for an account: a list ID or 'caldav'.
CREATE TABLE shopping_current (
    account_id TEXT PRIMARY KEY,
    list_id    TEXT NOT NULL
);
