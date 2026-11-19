# ADR 0026: Smart search answers on the server

- Status: Accepted
- Date: 2026-11-20
- Issues: [#575](https://github.com/Waylight-App/waylight/issues/575), [waylight-mobile#82](https://github.com/Waylight-App/waylight-mobile/issues/82)

## Context

The smart search (#155 to #162) answers questions from the data of the
modules: "cpu", "is the nas reachable", "which containers are down". The
intents, the tolerant matching and the answers live in the web interface,
which fetches the module data and computes the answers in the browser. The
mobile app needs the same answers. Porting the rules to Dart would keep two
copies of every intent, term and text in sync by hand.

## Decision

1. **The server answers.** `GET /api/search/answers?q=…&lang=…` returns up
   to three answers. They are data in five kinds (metric, status, list,
   text, actions), as in the web interface, and already in the client's
   language (`lang`, else `Accept-Language`; English and German).
2. **One engine in `internal/shared/answers`.** Folding, tokenising, the
   tolerant word match (prefix while typing, Damerau-Levenshtein for
   typos), intents as data with slots, and the ranking are a port of the web
   interface's `smart/` domain. The web interface's regression table runs
   against it.
3. **Sources stay in their modules.** A module implements
   `answers.Provider` and returns a `Source` from its adapter ring
   (`adapter/smartsearch`): its intents, the names it knows (containers,
   services) and the answers, with its own texts in both languages. The
   composition root hands the sources of the enabled modules to the search
   module, so no module imports another.
4. **Same rights as the module's API.** A source answers only when the
   account may `GET /api/<module>` (members, guests, the admin's policies);
   guests cannot call the endpoint.
5. **Step by step.** The first sources are system, docker (overview and
   status) and services. The other sources follow module by module. The web
   interface keeps its own answers until the server covers them, then
   switches to the endpoint. Actions (restarting a container) and answers
   that need the browser (calculator, time zones) stay in the clients for
   now.

## Consequences

- One copy of intents, terms and texts for all clients; the mobile app only
  renders the five kinds.
- Every search sends a request to the server; sources read cached module
  data, and one search is limited to five seconds.
- The server has texts of its own for the answers; they are kept in the
  source files next to the intents.
