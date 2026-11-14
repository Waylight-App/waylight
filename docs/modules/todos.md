# To-dos

**What it does:** A to-do list with due dates, overdue highlighting, recurring to-dos and reminders. A to-do widget shows either the built-in list or a **connected CalDAV task list** (#169), so tasks you already keep in Nextcloud Tasks, Synology Calendar, Radicale, Baïkal, Fastmail or another CalDAV server don't need a second list.

## Connecting a task list

1. Create an app password in your cloud (Nextcloud: _Settings → Security → Devices & sessions_).
2. In trisse-home open **Settings → To-do lists**, enter the server's DAV address (Nextcloud: `https://cloud.example.com/remote.php/dav`), user name and app password, and click **Find lists**. You can also paste the address of one list directly.
3. Pick the list, give it a name and click **Connect**. trisse-home reads the list once to test the connection.
4. In a to-do widget's settings (⚙ in edit mode), choose the list under **List**. Several widgets can show different lists.

The app password is stored encrypted with `SECRET_KEY` and never sent back to the browser. Disconnecting a list deletes nothing on the server.

## How syncing works

- A widget with a connected list checks for changes **every minute** while it is shown. The server asks for the list's CTag first; only when it changed does it compare the ETags and download the changed to-dos (`calendar-multiget`). Servers without CTag get the ETag comparison every time.
- Adding, editing, completing and deleting go straight to the server. trisse-home changes only the title, due date and completion of a to-do; priorities, categories, alarms and other properties set by other apps stay as they are.
- **Changed in both places:** the version of a to-do follows its ETag. If the other app changed it after you opened it, your change is refused, the widget says so and shows the server's version (the server wins).
- Offline: changes wait in the outbox and are sent when the connection is back, as for the built-in list.

## Limits

- Repeating to-dos, reminders and manual ordering are only available in the built-in list; for connected lists the edit dialog hides them. Connected lists are sorted by due date (to-dos without a date last), then by creation.
- Due dates are dates without a time. A due time set in another app is kept as long as you don't change the date.
- The daily briefing, the "due today" hint in the header and to-do reminders use the built-in list only.
- Tested against a CalDAV server that follows Nextcloud's responses. Other servers that follow RFC 4791 should work; please report one that doesn't.

## API

`GET/POST /api/todos`, `PUT/DELETE /api/todos/{id}`, `POST /api/todos/clear-completed`, each with an optional `?source=<list id>`; managing connected lists under `/api/todos/settings/sources` (not available to guests) – see `api/modules/todos.yaml`.
