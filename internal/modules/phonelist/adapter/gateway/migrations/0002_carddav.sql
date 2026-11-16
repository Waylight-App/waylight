-- Entries from a CardDAV address book (#300). source tells them apart from
-- entries typed in by an admin; a sync replaces all "carddav" rows.
ALTER TABLE phonelist_entries ADD COLUMN source TEXT NOT NULL DEFAULT 'manual';

-- The address book of the instance (one row). The password is encrypted
-- with the instance key and never exported.
CREATE TABLE phonelist_settings (
    id                 INTEGER PRIMARY KEY CHECK (id = 1),
    carddav_url        TEXT NOT NULL DEFAULT '',
    username           TEXT NOT NULL DEFAULT '',
    password_encrypted TEXT NOT NULL DEFAULT '',
    last_sync          TEXT NOT NULL DEFAULT '',
    last_error         TEXT NOT NULL DEFAULT ''
);
