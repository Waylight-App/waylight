-- Admin policies (#304): one row of instance-wide rules.
CREATE TABLE policies (
    id                    INTEGER PRIMARY KEY CHECK (id = 1),
    require_second_factor INTEGER NOT NULL DEFAULT 0,
    min_password_length   INTEGER NOT NULL DEFAULT 0,
    max_session_hours     INTEGER NOT NULL DEFAULT 0,
    member_modules        TEXT -- JSON list; NULL = all modules
);
