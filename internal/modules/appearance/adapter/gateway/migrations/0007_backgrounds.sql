-- #277: pictures uploaded as backgrounds, already scaled and re-encoded.
-- They live in the database so backups contain them.
CREATE TABLE appearance_backgrounds (
    id         TEXT PRIMARY KEY,
    width      INTEGER NOT NULL,
    height     INTEGER NOT NULL,
    size       INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    image      BLOB NOT NULL,
    thumb      BLOB NOT NULL
);
