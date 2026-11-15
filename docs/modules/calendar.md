# Calendar

**What it does:** Shows your appointments as an agenda, a month or a split view with event details, and lets you add events from the widget. Events come from **CalDAV** calendars (Nextcloud, Radicale, Baïkal, Fastmail, …), read-only **ICS subscriptions**, and **Google Calendar** or **Microsoft 365** accounts (#282). All sources end up in one agenda.

## Configuration

Everything lives under **Settings → Calendar**: the CalDAV addresses with user name and app password, the ICS subscriptions, and the connected Google and Microsoft accounts.

## Connecting Google Calendar or Microsoft 365

trisse-home signs in with an app registration of your own, the same way as for the [to-do lists](todos.md#connecting-microsoft-to-do-google-tasks-or-ticktick). There is no shared app run by the project, so no third party sits between you and your calendar.

Requirements: `PUBLIC_URL` must be set to the https:// address of trisse-home (see [configuration](../configuration.md)), because the provider sends the browser back to it after signing in. The settings show the **redirect address** to register: `PUBLIC_URL` + `/api/calendar/accounts/oauth/callback`. A registration you already made for the to-do lists can be reused if you add this second redirect address and the calendar permission.

**Google Calendar**
1. In the [Google Cloud console](https://console.cloud.google.com/apis/credentials) create a project and enable the _Google Calendar API_.
2. Configure the OAuth consent screen (type _External_ for a personal account) and add yourself as a test user.
3. _Create credentials → OAuth client ID_, type _Web application_, authorized redirect URI: the address shown in trisse-home.
4. Copy client ID and client secret into trisse-home and sign in. While the app is in "Testing", Google ends refresh tokens after 7 days; publish it to keep the sign-in.

**Microsoft 365 / Outlook.com**
1. In the [Microsoft Entra admin center](https://entra.microsoft.com/#view/Microsoft_AAD_RegisteredApps/ApplicationsListBlade) choose _App registrations → New registration_; supported account types: _personal Microsoft accounts and accounts in any organizational directory_.
2. Redirect URI: platform _Web_, the address shown in trisse-home.
3. Under _Certificates & secrets_ create a client secret. Under _API permissions_ add Microsoft Graph _Calendars.ReadWrite_ and _User.Read_ (delegated); `offline_access` is requested at sign-in.

After signing in, all calendars of the account are switched on. Untick the ones you do not want to see. Read-only calendars (holidays, shared calendars) are marked; new events can only be added to writable ones.

What is stored: client ID, client secret and the tokens, encrypted with `SECRET_KEY`. Tokens are renewed shortly before they expire. If the provider no longer accepts the sign-in, the agenda says so; disconnect and connect again. **Disconnect** revokes the token at Google; Microsoft has no revocation endpoint, so remove the app's access in your Microsoft account if you want to be sure.

## Outbound requests

- CalDAV and ICS: as described in [outbound requests](../outbound-requests.md).
- Google (`accounts.google.com`, `oauth2.googleapis.com`, `www.googleapis.com`) and Microsoft (`login.microsoftonline.com`, `graph.microsoft.com`): only after you connect an account. Then once when signing in, every 15 minutes for the agenda (one request per shown calendar), when a month outside the cached range is opened, and when you add an event.

## Limits

- Up to 5 accounts. Each shown calendar costs one request per refresh.
- Only creating events is supported, not editing or deleting them, the same as for CalDAV.
- Google and Microsoft: tested against fakes that follow the published API responses, not against real accounts.

## API

`GET /api/calendar/events`, `POST /api/calendar/events`, `GET /api/calendar/writable`, `GET /api/calendar/accounts`, `POST /api/calendar/accounts/oauth/start`, `PUT`/`DELETE /api/calendar/accounts/{id}` – see `api/modules/calendar.yaml`.
