# Architecture Decision Records

Each record documents one significant decision: its context, the decision and
its consequences. Records are immutable; a changed decision gets a new record
that supersedes the old one.

| No.                                         | Title                                         | Status   |
| ------------------------------------------- | --------------------------------------------- | -------- |
| [0001](0001-clean-architecture.md)          | Strict Clean Architecture for every module    | Accepted |
| [0002](0002-go-backend.md)                  | Go with the standard library for the backend  | Accepted |
| [0003](0003-sqlite.md)                      | SQLite via a pure-Go driver                   | Accepted |
| [0004](0004-svelte-spa.md)                  | Svelte 5 SPA embedded into the binary         | Accepted |
| [0005](0005-single-user-auth.md)            | Single-user auth with setup token and sessions| Accepted |
| [0006](0006-background-jobs-cache-aside.md) | Background jobs and cache-aside for integrations | Accepted |
| [0007](0007-own-dashboard-grid.md)          | Own grid implementation for the dashboard     | Accepted |
| [0009](0009-two-factor-authentication.md)   | Two-factor authentication with TOTP           | Accepted |
| [0008](0008-sample-data-on-setup.md)        | Sample data after the first setup             | Accepted |
| [0010](0010-internationalisation.md)        | Internationalisation without a framework      | Accepted |
| [0011](0011-lean-by-default.md)             | Lean by default: dependency policy and resource budgets | Accepted |
| [0012](0012-owner-and-guest-roles.md)       | Owner and read-only guest accounts (supersedes the single-user scope of 0005) | Accepted |
| [0013](0013-single-sign-on.md)              | Single sign-on with OpenID Connect and trusted proxy headers | Accepted |
| [0014](0014-passkeys.md)                    | Passkeys as second factor with a minimal WebAuthn implementation | Accepted |
| [0015](0015-raspberry-pi-image.md)          | Raspberry Pi appliance image: native systemd service on Raspberry Pi OS Lite | Proposed |
| [0016](0016-cross-module-read-models.md)    | Cross-module read models: consumer-owned ports, adapters in the composition root | Accepted |
| [0017](0017-open-core.md)                   | Open core: MIT core, team features in `ee/` under the Fair Core License | Accepted, superseded in part by 0020 |
| [0018](0018-data-ownership.md)              | Data ownership for several full accounts: owner and visibility per record for personal content | Accepted |
| [0019](0019-extension-without-server.md)    | The browser extension without a server: local gateways per feature | Accepted |
| [0020](0020-ee-stays-proprietary.md)        | `ee/` stays proprietary, and only organisation features go there (supersedes part of 0017) | Accepted |
| [0021](0021-ee-frontend.md)                | Frontend of `ee/` as a separate plugin package | Accepted |
| [0022](0022-several-admins.md)              | Several admins next to the owner: by hand or from single sign-on and directory groups | Accepted |
| [0023](0023-chat-apps-reach-private-instances.md) | How chat apps reach a private instance: Slack over Socket Mode, a Teams tab, the Teams command box search as an opt-in | Accepted |
| [0024](0024-app-sessions.md)                | App sessions for the mobile app: sessions with a device name and a bearer token, paired by QR code | Proposed |
| [0025](0025-push-relay.md)                  | Push relay: signed requests, tokens bound to the first instance, end-to-end encrypted payloads; FCM parts superseded by 0027 | Accepted |
| [0026](0026-smart-search-answers-on-the-server.md) | Smart search answers on the server: one engine, sources in their modules, same rights as the module API | Accepted |
| [0027](0027-push-over-server-sent-events.md) | Push to Android over Server-Sent Events from the relay instead of FCM; device keys, mailbox, one stream per device | Accepted |
