CREATE TABLE grafana_settings (
    id              INTEGER PRIMARY KEY CHECK (id = 1),
    url             TEXT NOT NULL,
    token_encrypted TEXT NOT NULL
);
