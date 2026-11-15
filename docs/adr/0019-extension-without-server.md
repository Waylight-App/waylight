# ADR 0019: The browser extension without a server

- Status: Accepted (2026-11-15)
- Date: 2026-11-15
- Issue: [#291](https://github.com/itadori26/trisse-home/issues/291), part of [#280](https://github.com/itadori26/trisse-home/issues/280)

## Context

The browser extension (#65) forwards the new tab to a running trisse-home
server. Someone without a server cannot try trisse-home at all, and the
extension is about to be listed on addons.mozilla.org (#290), where most
visitors will not have one.

#291 asks for a "Use without a server" mode: the web app runs inside the
extension against local storage. In scope are bookmarks, notes, to-dos,
countdowns, the focus timer, habits, the clock and the weather; everything
that needs background jobs or secrets (services, Docker, parcels, notifications
and the like) stays out. Data must be exportable in a form the server imports,
so moving to a server later is an import.

Two facts shape the decision:

- Every frontend feature already talks to its backend through a port in its
  use case ring (`NotesGateway`, `TodosGateway`, …) implemented by an
  `adapter/api` class, and the container in `web/src/app` chooses the
  implementation. The offline editing of to-dos and notes (#80) already
  wraps these gateways with an IndexedDB outbox.
- The extension is deliberately tiny today (10 files, no dependencies, three
  permissions; `scripts/check.mjs` enforces that).

## Options

### A. A local HTTP adapter

An `HttpClient` implementation that answers `/api/...` requests inside the
browser from IndexedDB, mirroring the server's endpoints for the modules in
scope. The rest of the app stays unchanged.

- One place to implement, and the use cases do not notice.
- It has to imitate HTTP: routes, status codes, the error format, DTO shapes
  and the server's validation messages. Every server change can silently break
  it, and nothing in the type system notices.

### B. A local gateway per feature

Each feature in scope gets a second implementation of its existing port,
for example `LocalNotesGateway implements NotesGateway`, backed by IndexedDB.
The container chooses the local or the API gateways once, at start, from the
mode.

- Uses the seams the architecture already has; the compiler checks each
  local gateway against its port, so a changed port fails the build.
- Validation that today lives only in Go has to exist in TypeScript too for
  the features in scope. Most of it is already in the frontend domains for
  instant form feedback (#80 needed the same).
- More files: one local gateway per feature, plus the dashboard, appearance
  and the modules list.

### C. A WebAssembly build of the Go server

Compile the server to WebAssembly and run it in the extension.

- Exact behaviour, no duplicated rules.
- Several megabytes of WebAssembly, SQLite in the browser, no background jobs
  in a Manifest V3 service worker that is stopped when idle; far beyond the
  lean budget of ADR 0011.

## Decision

**Option B, local gateways per feature**, with these rules:

- A new composition root `web/src/app/localContainer.ts` builds the container
  with local gateways for: dashboard (tabs, widgets, layout), appearance,
  modules (a fixed list of the modules in scope), bookmarks, notes, to-dos,
  countdowns, focus, habits. Auth is skipped: the extension is a single
  person's browser profile.
- **Weather** calls Open-Meteo directly from the extension page, so the
  extension gains exactly one host permission, `https://api.open-meteo.com/*`,
  requested only when the user chooses local mode (`optional_host_permissions`).
  `check.mjs` allows this one host and nothing else.
- Storage is **IndexedDB**, one object store per feature, in the shape of the
  server's tables (IDs, timestamps), so that the export (below) is a direct
  mapping.
- **Export and moving to a server:** the local mode writes the JSON export
  format of the `transfer` module (format `trisse-home-export`, sections per
  module with table rows). When the user later enters a server address, the
  extension offers to import that file into the server through the existing
  import endpoint, then switches to forwarding mode.
- **Modules out of scope are hidden**, not broken: the local modules list
  contains only the modules in scope, and the widget picker, settings and
  search already follow that list.
- **Pointer to the full version:** local mode says plainly what it leaves
  out. Where a module is hidden (widget picker, settings, modules list) and in
  a short "What you get with a server" section in the settings, the extension
  names what a server or the Raspberry Pi image adds (services, Docker,
  parcels, notifications, several devices and accounts, sync) and links to
  the installation guides. The hint is static text bundled with the
  extension, can be dismissed, and causes no outbound request.
- **Build:** a second Vite entry `web/src/extension.ts` builds the app into
  `extension/app/` for the extension zip. The forwarding mode stays as it is;
  the local app is only loaded when the user chooses local mode.
- **Budget:** a new entry in `budgets.json`, *extension zip*, starting at the
  measured size + 10 % (expected around 1.5 MB, mostly the app bundle).
- **Tests:** each local gateway runs the same use case tests as the API
  gateway (shared test cases per port), and one Playwright test loads the
  extension in local mode, adds a note and a to-do, exports, and imports the
  file into a test server.

## Consequences

- The extension stops being a 10-file package: it carries the app bundle and
  becomes subject to a size budget and the store review of a larger add-on.
- Validation rules of the features in scope exist twice (Go and TypeScript);
  the shared test cases per port keep them in step.
- New features are server-only by default; adding one to the local mode is a
  deliberate step with its own local gateway.
- The export format of the `transfer` module becomes a contract with the
  extension; changing a table's columns needs a matching change in the local
  export (a test round-trips an export through the server import).

## Follow-up work

1. Vite entry and `localContainer.ts`; extension packaging and budget.
2. Local gateways: dashboard, appearance, modules list.
3. Local gateways: bookmarks, notes, to-dos, countdowns, focus, habits; weather
   with the optional host permission.
4. Export in the transfer format; import into a server when an address is
   entered.
5. Options page: choose "Use without a server" or enter a server; store
   listing text updated (#290).
6. Pointer to the full version in the widget picker, the modules list and
   the settings, with links to the server and Raspberry Pi guides.
