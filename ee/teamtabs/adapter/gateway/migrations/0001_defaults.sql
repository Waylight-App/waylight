-- The team tab that opens by default for a group (#295).
CREATE TABLE teamtabs_defaults (
    group_id TEXT PRIMARY KEY,
    tab_id   TEXT NOT NULL
);
