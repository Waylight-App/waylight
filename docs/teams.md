# Microsoft Teams

With a licence key that includes `chat-apps` ([licence keys](licence-keys.md)),
Waylight works with Microsoft Teams in two ways, both set up by your own
admins with your own Microsoft registrations
([ADR 0023](adr/0023-chat-apps-reach-private-instances.md)). This needs a
server built with the `ee` features.

- **Search from the command box** (opt-in, below): type in the Teams
  message box and pick a service, bookmark or wiki page from Waylight.
- **The start page as a tab**: a tab with your instance's name and a
  button that opens the start page in the browser (below).

## The tab

The tab shows your instance's name and a button that opens the start page
in the browser, where people are signed in as usual. Signing in inside
Teams is deliberately not offered: the session cookie stays strict, so no
other site can make the browser send it along.

It works for an instance inside your network: the Teams client on each
person's device loads the page, so it only has to be reachable from there
over HTTPS with a certificate the device trusts. It needs `PUBLIC_URL` and
a licence key with `chat-apps`; without them `/teams/tab` answers 404.

Add a static tab to your Teams app package (the same package as the command
box search, or one of its own):

```json
"staticTabs": [{
  "entityId": "start",
  "name": "Start page",
  "contentUrl": "https://home.example.org/teams/tab",
  "websiteUrl": "https://home.example.org/",
  "scopes": ["personal"]
}],
"validDomains": ["home.example.org"]
```

Only this page may be shown inside Teams, Outlook and the Microsoft 365
app; the rest of the instance still refuses to be framed. The page loads
Microsoft's Teams library (version 2.58.0) from `res.cdn.office.net`, pinned
with a Subresource Integrity hash, because Teams shows a tab only after that
library has started.

## Search from the command box

This is the one chat feature that needs Microsoft to reach your instance
over the internet. It is off unless you switch it on.

### What becomes reachable

Only one route: `POST /api/teams/messages`. Make just that path reachable
from the internet, for example with a rule in your reverse proxy, and keep
the rest of the instance private. If you use `ALLOWED_CIDRS`, the route is
covered by it too; Microsoft's addresses change, so a proxy rule by path is
easier to keep right.

Every request must carry a token signed by Microsoft's Bot Connector for
your bot. Waylight checks it before anything else happens: issuer,
audience (your bot's app ID), validity with five minutes of clock skew, the
signature against Microsoft's published keys (refreshed daily), the service
address and that the key is endorsed for Teams. Anything else is answered
with 403.

### Set it up (admins)

1. In the Azure portal, create an **Azure Bot** resource (single tenant is
   fine) and note its **Microsoft App ID**. Set its messaging endpoint to
   `https://<your public address>/api/teams/messages` and add the
   **Microsoft Teams** channel.
2. Set the app ID on the server and restart:

   ```env
   TEAMS_BOT_APP_ID=00000000-0000-0000-0000-000000000000
   ```

   Waylight needs no bot password: the answer goes back in the HTTP
   response.
3. Create a Teams app package with a search command for that bot and upload
   it in the Teams admin center. The relevant part of `manifest.json`:

   ```json
   "composeExtensions": [{
     "botId": "00000000-0000-0000-0000-000000000000",
     "commands": [{
       "id": "search", "type": "query", "title": "Search the start page",
       "parameters": [{ "name": "q", "title": "Search", "description": "Service, bookmark or wiki page" }]
     }]
   }]
   ```

### How people use it

Everyone first connects their Microsoft 365 account under *Settings →
Connected accounts* ([connected accounts](connected-accounts.md)); that tells
Waylight which member is behind a Teams user. Others get a short hint and
no results.

Results are what that member could find on the start page: services, their
bookmarks and those shared with them, wiki pages including team wikis. Teams
waits five seconds; slower sources are left out of that answer. Nothing is
stored.

### Outbound requests

`login.botframework.com` for the signing keys (at most daily, and at most
every five minutes for an unknown key) and `graph.microsoft.com/v1.0/me`
to match connected accounts to Teams users.
