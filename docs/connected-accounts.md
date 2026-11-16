# Connected accounts

With a licence key that includes `connected-accounts` ([licence keys](licence-keys.md)),
each member can connect their own account at Microsoft 365, Confluence
(Atlassian Cloud) or Notion. The team integrations, such as wiki and chat
search, then act on the member's behalf and only find what that person may
see. This needs a server built with the `ee` features and `PUBLIC_URL`.

## Register trisse-home at a service (admins)

Each service needs an app registration of your organisation. Its redirect URI
is shown in the settings and is always `<PUBLIC_URL>/api/connections/callback`,
for example `https://home.example.org/api/connections/callback`. The address
only has to be reachable from your members' browsers, not from the internet.

- **Microsoft 365:** in the Entra admin center, *App registrations → New
  registration*, platform *Web*, the redirect URI above. Add a client secret.
  The delegated permissions trisse-home asks for are `User.Read`,
  `Sites.Read.All`, `Files.Read.All`, `ChannelMessage.Read.All`, `Chat.Read`
  and `offline_access`; some need admin consent. Enter the tenant ID or
  domain together with the client ID and secret.
- **Confluence:** in the Atlassian developer console, create an OAuth 2.0
  (3LO) app, add the Confluence API with `read:confluence-content.all`,
  `search:confluence` and `read:confluence-space.summary`, and set the
  callback URL above.
- **Notion:** create a public integration at notion.so/my-integrations with
  the redirect URI above. Notion tokens do not expire and have no refresh
  token.

Admins enter the client ID and secret per service at
`PUT /api/connections/clients/{service}`. The secret is stored encrypted and
never shown again.

## Connect an account (members)

Under *Settings → Connected accounts*, a member chooses *Connect*, signs in at
the service and is sent back. From then on the connection shows as connected.
When the service no longer accepts it (revoked, or the refresh token ran
out), it shows as expired with a link to connect again.

*Disconnect* revokes the token at the service where it supports that, and
deletes it here. Deleting a member deletes their tokens.

## What stays on the server

Access and refresh tokens are encrypted with `SECRET_KEY` and never sent to
the browser; the interface only sees whether a service is connected. Tokens
are renewed shortly before they expire. Connecting and disconnecting are
recorded in the [audit log](audit-log.md).
