-- Followed artists per account (ADR 0018, #294). The artist ID comes from
-- MusicBrainz, so two accounts can follow the same artist: the key becomes
-- (owner_id, id). SQLite cannot change a primary key, so the table is
-- rebuilt. Existing artists belong to the owner, visible to everyone.
CREATE TABLE releases_artists_v2 (
    owner_id   TEXT NOT NULL DEFAULT 'owner',
    id         TEXT NOT NULL,
    name       TEXT NOT NULL,
    position   INTEGER NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'everyone',
    PRIMARY KEY (owner_id, id)
);
INSERT INTO releases_artists_v2 (owner_id, id, name, position) SELECT 'owner', id, name, position FROM releases_artists;
DROP TABLE releases_artists;
ALTER TABLE releases_artists_v2 RENAME TO releases_artists;
