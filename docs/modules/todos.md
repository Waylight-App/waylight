# To-dos

**What it does:** A to-do list with due dates, overdue highlighting, recurring to-dos and reminders. A to-do widget shows either the built-in list or a **connected list**: a CalDAV task list (#169: Nextcloud Tasks, Synology Calendar, Radicale, Baïkal, Fastmail, …) a **Todoist** project (#170), or a list in **Microsoft To Do**, **Google Tasks** or **TickTick** (#171). Tasks you already keep elsewhere don't need a second list.

## Connecting a task list

1. Create an app password in your cloud (Nextcloud: _Settings → Security → Devices & sessions_).
2. In Waylight open **Settings → To-do lists**, enter the server's DAV address (Nextcloud: `https://cloud.example.com/remote.php/dav`), user name and app password, and click **Find lists**. You can also paste the address of one list directly.
3. Pick the list, give it a name and click **Connect**. Waylight reads the list once to test the connection.
4. In a to-do widget's settings (⚙ in edit mode), choose the list under **List**. Several widgets can show different lists.

The app password is stored encrypted with `SECRET_KEY` and never sent back to the browser. Disconnecting a list deletes nothing on the server.

## Connecting Todoist

1. In Todoist open _Settings → Integrations → Developer_ and copy the **API token**.
2. In Waylight open **Settings → To-do lists**, choose **Todoist**, paste the token and click **Find lists**.
3. Pick the project, give it a name and click **Connect**. Then choose it in a to-do widget as above.

The token is stored encrypted with `SECRET_KEY` and never sent back to the browser. It gives full access to your Todoist account, as every Todoist API token does; disconnecting the list removes it from Waylight (revoke it in Todoist if you no longer need it).

How it syncs: the first request loads the active tasks of the account through Todoist's sync endpoint; later requests only fetch what changed since then (sync token). Several widgets share one request, and Todoist is asked at most every 15 seconds. When Todoist answers _429 Too Many Requests_, Waylight pauses for the time Todoist names and shows the last known list meanwhile. Writes use the REST endpoints and send only what changed: a title change leaves a (possibly recurring) Todoist due date untouched. Completing a recurring Todoist task moves it to its next date, as in the Todoist apps. Completed tasks appear once they were completed after the first sync.

## Connecting Microsoft To Do, Google Tasks or TickTick

These services need a sign-in (OAuth). Waylight uses **an app you register yourself**, so no company sits in between and your tokens never leave your server.

Requirements: `PUBLIC_URL` must be set to the https:// address of Waylight (see [configuration](../configuration.md)). The settings show the **redirect address** to register: `PUBLIC_URL` + `/api/todos/settings/oauth/callback`.

**Microsoft To Do**
1. In the [Microsoft Entra admin center](https://entra.microsoft.com/#view/Microsoft_AAD_RegisteredApps/ApplicationsListBlade) choose _App registrations → New registration_. Supported account types: _personal Microsoft accounts and accounts in any organizational directory_.
2. Redirect URI: platform _Web_, the address shown in Waylight.
3. Under _Certificates & secrets_ create a client secret. Under _API permissions_ add Microsoft Graph _Tasks.ReadWrite_ (delegated); `offline_access` is requested at sign-in.
4. Copy the _Application (client) ID_ and the secret value.

**Google Tasks**
1. In the [Google Cloud console](https://console.cloud.google.com/apis/credentials) create a project and enable the _Google Tasks API_.
2. Configure the OAuth consent screen (_External_, add yourself as test user, scope `…/auth/tasks`).
3. _Create credentials → OAuth client ID_, type _Web application_, authorized redirect URI: the address shown in Waylight.
4. Copy client ID and client secret. While the app is in "Testing", Google ends refresh tokens after 7 days; publish the app (no review is needed for your own use with the warning screen) to keep the sign-in.

**TickTick**
1. In the [TickTick developer center](https://developer.ticktick.com/manage) create an app and set the OAuth redirect URL to the address shown in Waylight.
2. Copy client ID and client secret.

Then in Waylight: **Settings → To-do lists**, choose the service, paste client ID and secret, click **Sign in**, confirm in the window that opens, pick the list and click **Connect**.

What is stored: client ID, client secret and the tokens, encrypted with `SECRET_KEY`. Tokens are renewed automatically shortly before they expire. If a service no longer accepts the sign-in (revoked, or the refresh token ended), the widget says so; disconnect and connect again. Disconnecting revokes the token at Google; Microsoft and TickTick have no revocation endpoint, so remove the app's access in your account if you want to be sure.

Behaviour per service:
- A list is loaded at most every 30 seconds; several widgets share one request.
- Microsoft To Do: edits are sent with the task's ETag (`If-Match`), so a change made in To Do meanwhile wins. A due date that did not change is not sent, so a time or reminder set in To Do stays.
- Google Tasks: Google keeps only the date of a due date. Google has no creation time, so undated tasks are ordered by their last change.
- TickTick: the Open API lists only open tasks, so a completed task leaves the list. Reopening sends the task with status "open".

## How syncing works (CalDAV)

- A widget with a connected list checks for changes **every minute** while it is shown. The server asks for the list's CTag first; only when it changed does it compare the ETags and download the changed to-dos (`calendar-multiget`). Servers without CTag get the ETag comparison every time.
- Adding, editing, completing and deleting go straight to the server. Waylight changes only the title, due date and completion of a to-do; priorities, categories, alarms and other properties set by other apps stay as they are.
- **Changed in both places:** the version of a to-do follows its ETag. If the other app changed it after you opened it, your change is refused, the widget says so and shows the server's version (the server wins).
- Offline: changes wait in the outbox and are sent when the connection is back, as for the built-in list.

## Limits

- Repeating to-dos, reminders and manual ordering are only available in the built-in list; for connected lists the edit dialog hides them. Connected lists are sorted by due date (to-dos without a date last), then by creation.
- Due dates are dates without a time. A due time set in another app is kept as long as you don't change the date.
- The daily briefing, the "due today" hint in the header and to-do reminders use the built-in list only.
- Tested against a CalDAV server that follows Nextcloud's responses. Other servers that follow RFC 4791 should work; please report one that doesn't.
- Todoist, Microsoft To Do, Google Tasks and TickTick: tested against fakes that follow the published API responses, not against real accounts.
- Todoist: Subtasks are shown as ordinary to-dos of the project; sections, labels and priorities are kept but not shown.

## API

`GET/POST /api/todos`, `PUT/DELETE /api/todos/{id}`, `POST /api/todos/clear-completed`, each with an optional `?source=<list id>`; managing connected lists under `/api/todos/settings/sources` (not available to guests) – see `api/modules/todos.yaml`.
