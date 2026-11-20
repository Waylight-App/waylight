-- Mailbox of the Server-Sent Events delivery (ADR 0027): sealed payloads
-- per device until the device acknowledges them or they expire.
CREATE TABLE mailbox (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    device  TEXT NOT NULL,
    payload BLOB NOT NULL,
    created INTEGER NOT NULL
) STRICT;

CREATE INDEX mailbox_device ON mailbox (device, id);
CREATE INDEX mailbox_created ON mailbox (created);
