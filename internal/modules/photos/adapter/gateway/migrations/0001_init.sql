-- Photo frame (#286): picture metadata; the JPEGs are kept apart so that
-- listing an album never reads the blobs.
CREATE TABLE photos (
    id         TEXT PRIMARY KEY,
    album      TEXT NOT NULL,
    position   INTEGER NOT NULL,
    width      INTEGER NOT NULL,
    height     INTEGER NOT NULL,
    size       INTEGER NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX photos_album ON photos (album, position);

CREATE TABLE photo_files (
    photo_id TEXT PRIMARY KEY REFERENCES photos (id) ON DELETE CASCADE,
    image    BLOB NOT NULL,
    thumb    BLOB NOT NULL
);
