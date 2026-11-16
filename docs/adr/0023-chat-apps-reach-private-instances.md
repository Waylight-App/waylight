# ADR 0023: How chat apps reach a private instance

- Status: Proposed
- Date: 2026-11-16
- Issues: [#333](https://github.com/itadori26/trisse-home/issues/333), [#334](https://github.com/itadori26/trisse-home/issues/334), part of [#335](https://github.com/itadori26/trisse-home/issues/335)

## Context

A Slack slash command (#333) and a search in the Teams command box (#334)
are requests that start at Slack or Microsoft and must reach the instance.
Until now every request to trisse-home comes from a browser in the
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

1. **The instance accepts no requests from the internet for chat apps.**
   No public endpoint, no relay run by the project, no change to the
   default that an instance is private.
2. **Slack uses Socket Mode** (#333). The instance connects out to Slack
   with an app-level token (`SLACK_APP_TOKEN`, scope `connections:write`)
   and answers slash commands over that connection. Without the token,
   nothing connects; it is off by default and needs a licence key with
   `workplace-search`.
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
3. **Teams gets the tab, not the command box search** (#334). A Teams app
   package with a tab that shows the start page is offered; it works for
   private instances because the Teams client on the person's device loads
   it. The command box search needs a public bot endpoint and is not built
   for self-hosted instances. It can be revisited for an instance that is
   reachable from the internet on purpose, in a new ADR.

## Consequences

- Slack commands work for an instance inside a private network, as long
  as it can open an outbound connection to `wss-primary.slack.com`. The
  outbound requests page lists it.
- One long-lived outbound connection per instance while Slack is
  configured; it reconnects when Slack asks it to.
- The tab needs the start page to allow being framed by Teams
  (`frame-ancestors` for the Teams domains) and a sign-in that works inside
  Teams; that is part of #334.
- The acceptance item "requests are verified with Slack's signing secret"
  of #333 is replaced by the authenticated outbound connection.
