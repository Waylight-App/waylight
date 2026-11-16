# Connected accounts

With a licence key that includes `connected-accounts` ([licence keys](licence-keys.md)),
each member can connect their own account at Microsoft 365, Confluence
(Atlassian Cloud), Notion or Slack. The team integrations, such as wiki and chat
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
  and `offline_access`. `ChannelMessage.Read.All` (Teams channel messages)
  needs admin consent in most tenants: grant it for the organisation in the
  app registration before members connect, or they are asked for an
  admin's approval. Enter the tenant ID or domain together with the client
  ID and secret.
- **Confluence:** in the Atlassian developer console, create an OAuth 2.0
  (3LO) app, add the Confluence API with `read:confluence-content.all`,
  `search:confluence` and `read:confluence-space.summary`, and set the
  callback URL above.
- **Notion:** create a public integration at notion.so/my-integrations with
  the redirect URI above. Notion tokens do not expire and have no refresh
  token.
- **Slack:** create an internal app in your workspace at api.slack.com/apps
  (*Create New App → From a manifest*) with the manifest below, and copy its
  client ID and secret. Slack's Real-time Search API is open only to internal
  apps and apps from the Slack Marketplace, so each workspace uses its own
  app. Under *Agents & AI Apps* in the app settings, switch the feature on;
  Slack requires it for the search API. The workspace needs a plan that
  includes the Real-time Search API.

  ```yaml
  display_information:
    name: trisse-home search
  oauth_config:
    redirect_urls:
      - https://home.example.org/api/connections/callback
    scopes:
      user:
        - search:read.public
        - search:read.private
        - search:read.mpim
        - search:read.im
  settings:
    org_deploy_enabled: false
  ```

  trisse-home asks for exactly these four user scopes and never for a bot
  token. The search runs as the member, so private channels and direct
  messages are found only where the member is a participant. A token that
  lacks the private scopes still searches public channels.

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
