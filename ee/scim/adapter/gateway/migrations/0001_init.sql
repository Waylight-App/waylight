-- What a directory said about members and groups that the auth module does
-- not keep (#396): its user names, its own IDs and display names.
CREATE TABLE scim_users (
    account_id   TEXT PRIMARY KEY,
    user_name    TEXT NOT NULL,
    external_id  TEXT NOT NULL DEFAULT '',
    display_name TEXT NOT NULL DEFAULT ''
);
CREATE TABLE scim_groups (
    group_id    TEXT PRIMARY KEY,
    external_id TEXT NOT NULL DEFAULT ''
);
