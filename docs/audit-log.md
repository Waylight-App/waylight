# Audit log

With a licence key that includes `audit-log` ([licence keys](licence-keys.md)),
Waylight records who changed shared data and settings, when, and which
object. It never records the content: the log says that a service was
changed, not what it is called now.

This needs a server built with the `ee` features.

## What is recorded

Every successful change through the API (create, change, delete) to:

- instance data and settings: services, Docker, backups, the phone list,
  members and groups, appearance defaults and so on;
- sharing: making a note, to-do or other personal record visible to a group
  or to everyone.

Each entry has the time, the account and its role, the module, the kind of
change (for example `PUT /api/services/{id}`) and the object's ID.

Personal content of an account (their own notes, to-dos and settings) is not
recorded, so an admin cannot read a member's private activity from the log.
Sign-ins and account security are in the separate security log.

Changes from background jobs and the configuration file are not recorded;
only changes people make through the interface or an API token are.

## Read and export

Admins read the log at `GET /api/audit`, newest first, with optional filters
`account`, `module`, `from` and `to` (dates or RFC 3339 times) and `limit`
(up to 500). `GET /api/audit/export?format=csv` or `format=json` downloads
everything matching the filters.

Entries cannot be changed or deleted through the application. They only
leave the log when they are older than the retention period.

## Retention

Entries are kept for 365 days by default. Set another period with
`PUT /api/audit/settings` (`{"retentionDays": 90}`), from 30 to 3650 days, or
0 to keep entries for good. The change is itself recorded. Old entries are
removed once a day.

If the licence key expires or covers too few accounts, changes keep being
recorded and the log stays readable; only the retention cannot be changed.
Without a key nothing is recorded.
