# To-dos

**What it does:** A to-do list with due dates, overdue highlighting, recurring to-dos and reminders. A to-do widget shows either the built-in list or a **connected list**: a CalDAV task list (#169: Nextcloud Tasks, Synology Calendar, Radicale, Baïkal, Fastmail, …) or a **Todoist** project (#170). Tasks you already keep elsewhere don't need a second list.

## Connecting a task list

1. Create an app password in your cloud (Nextcloud: _Settings → Security → Devices & sessions_).
2. In trisse-home open **Settings → To-do lists**, enter the server's DAV address (Nextcloud: `https://cloud.example.com/remote.php/dav`), user name and app password, and click **Find lists**. You can also paste the address of one list directly.
3. Pick the list, give it a name and click **Connect**. trisse-home reads the list once to test the connection.
4. In a to-do widget's settings (⚙ in edit mode), choose the list under **List**. Several widgets can show different lists.

The app password is stored encrypted with `SECRET_KEY` and never sent back to the browser. Disconnecting a list deletes nothing on the server.

## Connecting Todoist

1. In Todoist open _Settings → Integrations → Developer_ and copy the **API token**.
2. In trisse-home open **Settings → To-do lists**, choose **Todoist**, paste the token and click **Find lists**.
3. Pick the project, give it a name and click **Connect**. Then choose it in a to-do widget as above.

The token is stored encrypted with `SECRET_KEY` and never sent back to the browser. It gives full access to your Todoist account, as every Todoist API token does; disconnecting the list removes it from trisse-home (revoke it in Todoist if you no longer need it).

How it syncs: the first request loads the active tasks of the account through Todoist's sync endpoint; later requests only fetch what changed since then (sync token). Several widgets share one request, and Todoist is asked at most every 15 seconds. When Todoist answers _429 Too Many Requests_, trisse-home pauses for the time Todoist names and shows the last known list meanwhile. Writes use the REST endpoints and send only what changed: a title change leaves a (possibly recurring) Todoist due date untouched. Completing a recurring Todoist task moves it to its next date, as in the Todoist apps. Completed tasks appear once they were completed after the first sync.

## How syncing works (CalDAV)

- A widget with a connected list checks for changes **every minute** while it is shown. The server asks for the list's CTag first; only when it changed does it compare the ETags and download the changed to-dos (`calendar-multiget`). Servers without CTag get the ETag comparison every time.
- Adding, editing, completing and deleting go straight to the server. trisse-home changes only the title, due date and completion of a to-do; priorities, categories, alarms and other properties set by other apps stay as they are.
- **Changed in both places:** the version of a to-do follows its ETag. If the other app changed it after you opened it, your change is refused, the widget says so and shows the server's version (the server wins).
- Offline: changes wait in the outbox and are sent when the connection is back, as for the built-in list.

## Limits

- Repeating to-dos, reminders and manual ordering are only available in the built-in list; for connected lists the edit dialog hides them. Connected lists are sorted by due date (to-dos without a date last), then by creation.
- Due dates are dates without a time. A due time set in another app is kept as long as you don't change the date.
- The daily briefing, the "due today" hint in the header and to-do reminders use the built-in list only.
- Tested against a CalDAV server that follows Nextcloud's responses. Other servers that follow RFC 4791 should work; please report one that doesn't.
- Todoist: tested against a fake that follows the responses of the Todoist API v1, not against a real account. Subtasks are shown as ordinary to-dos of the project; sections, labels and priorities are kept but not shown.

## API

`GET/POST /api/todos`, `PUT/DELETE /api/todos/{id}`, `POST /api/todos/clear-completed`, each with an optional `?source=<list id>`; managing connected lists under `/api/todos/settings/sources` (not available to guests) – see `api/modules/todos.yaml`.
