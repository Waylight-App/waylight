# Announcements

With a licence key that includes `announcements` ([licence keys](licence-keys.md)),
admins post short notices for everyone on the instance, such as "The office
is closed on Friday". This needs a server built with the `ee` features.

## Who writes and who reads

- **Admins** write, edit, pin and delete every announcement.
- **A writer group** can write too: choose one under the announcement
  settings. Its members write announcements and change or delete their own.
  Guests and kiosk displays in that group still only read.
- **Everyone** who can see the widget reads the announcements: members,
  admins, and guests and kiosk displays whose shared tab has the widget.
- Members and admins mark announcements as read; unread ones are
  highlighted. Read marks belong to each person and are removed with the
  account.

## What an announcement has

- A title (up to 120 characters) and a text in Markdown (up to 10,000).
- **Pinned** announcements stay at the top; the rest are newest first.
- An optional **expiry**: afterwards the announcement leaves the widget and
  is kept in the archive (`GET /api/announcements?archive=1`) until someone
  deletes it.
- **Notify**: when set on publishing, the announcement goes once to the
  notification channels under the type *announcement*. Editing does not
  notify again unless the author sets it again.

At most 50 announcements can be current at the same time.

With an expired licence key or too few seats, announcements stay readable,
but nothing can be written. Writes are listed in the [audit log](audit-log.md).

## API

| Method | Path | |
| --- | --- | --- |
| GET | `/api/announcements` | Current announcements; `?archive=1` the expired ones. Each has `read` (for the asking person) and `editable`. |
| POST | `/api/announcements` | `{title, body, pinned, expiresAt, notify}` |
| PUT | `/api/announcements/{id}` | Same fields |
| DELETE | `/api/announcements/{id}` | |
| POST | `/api/announcements/{id}/read` | Marks it read for the signed-in member or admin |
| GET, PUT | `/api/announcements/settings` | `{writerGroup}`, a group ID or empty; admins only |
