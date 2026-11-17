# ADR 0023: How chat apps reach a private instance

- Status: Accepted (2026-11-16, by the maintainer)
- Date: 2026-11-16
- Issues: [#333](https://github.com/itadori26/waylight/issues/333), [#334](https://github.com/itadori26/waylight/issues/334), part of [#335](https://github.com/itadori26/waylight/issues/335)

## Context

A Slack slash command (#333) and a search in the Teams command box (#334)
are requests that start at Slack or Microsoft and must reach the instance.
Until now every request to Waylight comes from a browser in the
organisation's network; the instance is private by default and nothing in
it expects traffic from the internet.

Slack offers two ways to deliver events to an app:

- **HTTP:** Slack sends each command to a public HTTPS URL, signed with the
  app's signing secret.
- **Socket Mode:** the app opens a WebSocket connection to Slack with an
  app-level token, and Slack delivers commands over it. No public URL is
  needed. It is meant for apps that are not distributed in the Slack
  Marketplace, which is the case here: each workspace creates its own
  internal app (as for the search in #330).

Microsoft Teams has no such option for message extensions: the Bot
Framework always calls a public HTTPS endpoint of the bot. A Teams tab, in
contrast, is a web page the Teams client loads in the person's own browser
or desktop app, so it only needs to be reachable from there, like the
start page itself.

## Decision

1. **Each organisation's admins run the connection to their chat
   service.** The project runs no relay and no shared app: every instance
   uses an app its own organisation registered at Slack or Microsoft, and
   nothing passes through a third party. By default the instance accepts
   no requests from the internet for chat apps.
2. **All of it is part of the `ee/` features** ([ADR 0020](0020-ee-stays-proprietary.md))
   and needs a licence key that includes `chat-apps`: the Slack command,
   the Teams tab and the Teams command box search. The core offers no chat
   app.
3. **Slack uses Socket Mode** (#333). The instance connects out to Slack
   with an app-level token (`SLACK_APP_TOKEN`, scope `connections:write`)
   and answers slash commands over that connection. Without the token,
   nothing connects; it is off by default.
   - The connection is authenticated by the app-level token over TLS to
     Slack, so there is no inbound request whose signature would need
     checking; the signing secret is not used.
   - The asking Slack user is matched to a member through the Slack
     account they connected (#326). Results are those that member may see
     (services, bookmarks, wiki pages and the team sources). Someone without
     a connection gets a short reply with a link to connect, and nothing
     else.
   - The answer is sent as the acknowledgement within Slack's three
     seconds, visible only to the person who asked.
4. **Teams: the tab for everyone, the command box search as an opt-in**
   (#334).
   - A Teams app package with a tab that shows the start page is offered;
     it works for private instances because the Teams client on the
     person's device loads it.
   - The command box search (a message extension) needs Microsoft to reach
     a bot endpoint of the instance over the internet. It is **off by
     default** and only for organisations whose admins make their instance
     reachable from the internet on purpose. It is switched on by
     configuring the organisation's own bot registration
     (`TEAMS_BOT_APP_ID`); without it the endpoint does not exist. The
     answer goes back in the HTTP response, so the instance needs no bot
     password.
   - The endpoint is a single route, `/api/teams/messages`, separate from
     the normal API and outside its session handling. Every request must
     carry a valid Bot Framework token (signed by Microsoft, issued for the
     organisation's bot app ID); anything else is refused before any
     search runs. The asking Teams user is matched to a member through the
     Microsoft 365 account they connected (#326), as for Slack.
   - The docs say plainly what making an instance public means and that
     only the bot route needs to be reachable (for example through the
     reverse proxy), not the whole instance.

## Consequences

- Slack commands work for an instance inside a private network, as long
  as it can open an outbound connection to `wss-primary.slack.com`. The
  outbound requests page lists it.
- One long-lived outbound connection per instance while Slack is
  configured; it reconnects when Slack asks it to.
- With the Teams command box search switched on, one route of the
  instance is reachable from the internet; its security rests on the Bot
  Framework token check, which is tested against forged, expired and
  wrong-audience tokens.
- The tab is a page of its own that Microsoft's hosts may frame
  (`frame-ancestors` for that page only). It does not sign anyone in inside
  Teams: it opens the start page in the browser (maintainer's decision on
  #334), so the session cookie keeps `SameSite=Strict`.
- The acceptance item "requests are verified with Slack's signing secret"
  of #333 is replaced by the authenticated outbound connection.
