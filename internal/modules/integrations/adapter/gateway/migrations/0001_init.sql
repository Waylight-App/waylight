-- Named connections to a service (#610): saved once by an admin and chosen
-- by the widgets that use the service. The credential is a JSON object of
-- its fields, encrypted with SECRET_KEY; it never leaves the server.
CREATE TABLE integrations_connections (
    id                   TEXT PRIMARY KEY,
    service              TEXT NOT NULL,
    name                 TEXT NOT NULL,
    url                  TEXT NOT NULL,
    credential_encrypted TEXT NOT NULL,
    created_at           TEXT NOT NULL
);
CREATE UNIQUE INDEX integrations_connections_name ON integrations_connections (service, name COLLATE NOCASE);
