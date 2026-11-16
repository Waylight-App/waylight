# Shopping list

**What it does:** a shopping list on the start page. Keep the list on
trisse-home and share it with the people you shop for, or connect a CalDAV
task list that everyone already uses in their task app.

## Using it

Add the *Shopping list* widget from the *Information* category. Without a list
yet, the widget offers to create one: type a name (or leave the suggested
one) and press *Create list*.

Type an item and press Enter to add it. Tick an item off when it is in the
basket; ticked items are struck through and listed after the open ones. *Remove n done* clears
them all at once, and × next to an item removes just that one.

When you have more than one list, or a list and a CalDAV list, a choice at
the top of the widget switches between them. trisse-home remembers the
choice per account, so every device shows the same list.

## Lists on trisse-home

*Settings → Shopping list* manages the lists kept on trisse-home. An account
can have several, for example *Groceries* and *Hardware store*.

- **Rename** a list by changing its name and pressing *Rename*.
- **Share** a list with one of your groups or with everyone, like notes and
  to-dos. Everyone who can see the list can add, tick off and remove items;
  renaming, sharing and deleting stay with you. Lists others share with you
  show who they come from.
- **Export** a list *as text* (one item per line) or *as .ics*. The `.ics`
  file holds one task (VTODO) per item, ticked ones marked as completed, and
  imports into most task apps.
- **Import from CalDAV** copies the open items of your connected CalDAV list
  into a list on trisse-home, once. Use it to move over from CalDAV; the
  CalDAV list itself is not changed.
- **Delete** removes the list with all its items.

Lists and their items are part of backups and of the data export.

## CalDAV task list

If your household already keeps its shopping list in a task app, connect
that list instead: *Settings → Shopping list → CalDAV task list*, with the
URL of the task list, a user name and an app password. Items are tasks
(VTODO) in that list, so changes show up in Nextcloud Tasks, Apple
Reminders or Tasks.org (with DAVx⁵) and on the start page alike. Create a
list called "shopping" in Nextcloud Tasks and use its URL.

Accounts that only have a CalDAV list keep seeing it in the widget; lists on
trisse-home are an addition, not a replacement.

## Outbound requests

Only with a CalDAV list: CalDAV `REPORT`, `PUT` and `DELETE` requests to the
configured task list when the widget loads (every minute while it is open)
and when items change, with the stored user name and app password. The
password is encrypted in the database and left out of exports. Lists on
trisse-home make no outside requests.

## API

Lists on trisse-home: `GET` and `POST /api/shopping/lists`, `PUT` and
`DELETE /api/shopping/lists/{id}`, `PUT /api/shopping/current`, the items
under `/api/shopping/lists/{id}/items`, plus
`POST /api/shopping/lists/{id}/clear-completed`,
`POST /api/shopping/lists/{id}/import` and
`GET /api/shopping/lists/{id}/export?format=text|ics`.

CalDAV list: `GET` and `PUT /api/shopping/settings`, the items under
`/api/shopping/items` and `POST /api/shopping/clear-completed`. See
`api/modules/shopping.yaml`.
