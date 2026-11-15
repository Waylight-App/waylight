-- Subscriptions per account (ADR 0018, #294). The feed URL was unique for
-- the instance; now each account can subscribe to the same feed, so the
-- constraint becomes UNIQUE (owner_id, url). Existing feeds belong to the
-- owner and stay visible to everyone.
--
-- SQLite cannot change a constraint, so the table is rebuilt. Foreign keys
-- are on during migrations, and dropping rss_feeds would cascade to every
-- item. Both tables are therefore copied first; the old items table goes
-- before the old feeds table, and the renames update the reference.
CREATE TABLE rss_feeds_v2 (
    id              TEXT PRIMARY KEY,
    url             TEXT NOT NULL,
    title           TEXT NOT NULL,
    site_url        TEXT NOT NULL DEFAULT '',
    last_fetched_at TEXT,
    last_error      TEXT NOT NULL DEFAULT '',
    created_at      TEXT NOT NULL,
    category        TEXT NOT NULL DEFAULT '',
    owner_id        TEXT NOT NULL DEFAULT 'owner',
    visibility      TEXT NOT NULL DEFAULT 'everyone',
    UNIQUE (owner_id, url)
);
INSERT INTO rss_feeds_v2 (id, url, title, site_url, last_fetched_at, last_error, created_at, category)
SELECT id, url, title, site_url, last_fetched_at, last_error, created_at, category FROM rss_feeds;

CREATE TABLE rss_items_v2 (
    id           TEXT PRIMARY KEY,
    feed_id      TEXT NOT NULL REFERENCES rss_feeds_v2 (id) ON DELETE CASCADE,
    guid         TEXT NOT NULL,
    title        TEXT NOT NULL,
    link         TEXT NOT NULL DEFAULT '',
    published_at TEXT NOT NULL,
    read         INTEGER NOT NULL DEFAULT 0,
    UNIQUE (feed_id, guid)
);
INSERT INTO rss_items_v2 (id, feed_id, guid, title, link, published_at, read)
SELECT id, feed_id, guid, title, link, published_at, read FROM rss_items;

DROP TABLE rss_items;
DROP TABLE rss_feeds;
ALTER TABLE rss_feeds_v2 RENAME TO rss_feeds;
ALTER TABLE rss_items_v2 RENAME TO rss_items;
CREATE INDEX rss_items_published ON rss_items (published_at DESC);
CREATE INDEX rss_feeds_owner ON rss_feeds (owner_id);
