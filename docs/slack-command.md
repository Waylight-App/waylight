# Slack command

With a licence key that includes `chat-apps` ([licence keys](licence-keys.md)),
people in your Slack workspace can search the start page from any
conversation: `/home vpn` answers with the matching services, bookmarks and
wiki pages, visible only to the person who asked. This needs a server built
with the `ee` features.

The instance connects out to Slack over Socket Mode
([ADR 0023](adr/0023-chat-apps-reach-private-instances.md)): Slack never
needs to reach it, so it also works for an instance that is only reachable
inside your network, as long as it can open an outbound connection to
Slack. Each workspace uses its own app, set up by its admins; nothing goes
through anyone else.

## Set it up (admins)

1. Use the workspace's internal app from
   [connected accounts](connected-accounts.md), or create one the same way,
   and add to its manifest:

   ```yaml
   features:
     slash_commands:
       - command: /home
         description: Search the start page
         usage_hint: vpn
   settings:
     socket_mode_enabled: true
   ```

2. Under *Basic Information → App-Level Tokens*, create a token with the
   scope `connections:write`. It starts with `xapp-`.
3. Set it on the server and restart:

   ```env
   SLACK_APP_TOKEN=xapp-…
   ```

Without the token nothing connects.

## How people use it

Everyone first connects their own Slack account under *Settings → Connected
accounts* ([connected accounts](connected-accounts.md)). That tells
Waylight which member is behind a Slack user; someone who has not
connected gets a short answer with a link to do so, and no results.

- `/home vpn` lists up to eight results: services whose name or
  description matches, the person's bookmarks and those shared with them,
  and pages from the wikis, including the team wikis they connected.
- The search runs with that member's rights; it finds nothing they could
  not find on the start page. Bookmarks shared with one of their groups are
  not searched from Slack yet.
- Slack waits three seconds for the answer. Sources that are slower are
  left out of that answer.
- Nothing from the search is stored, and the command never posts into the
  conversation for others.

## Outbound requests

While `SLACK_APP_TOKEN` is set: one long-lived WebSocket connection to
Slack (`wss-primary.slack.com` and its siblings, renewed hourly), plus
`apps.connections.open` and `auth.test` at `slack.com` to open the
connection and to learn which Slack user each connected account belongs to.
