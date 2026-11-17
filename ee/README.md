# ee/ – features for organisations

This directory is **not open source**. It is source-available under the
[Elastic License 2.0](https://www.elastic.co/licensing/elastic-license)
(ELv2), with no future licence: its code does not become MIT or any other
open-source licence over time. The rest of the repository stays under the
MIT licence. The decision is [ADR 0020](../docs/adr/0020-ee-stays-proprietary.md),
which replaces the licence part of [ADR 0017](../docs/adr/0017-open-core.md).

The licensor is Tristan L. Farinella; the full text is in
[`LICENSE.md`](LICENSE.md).

## What the licence means

- You may read, use, change and redistribute this code.
- You may not offer it to others as a hosted or managed service that gives
  them access to its features.
- You may not move, change, disable or get around the licence key check, or
  remove features that the licence key protects.
- You may not remove or hide licence, copyright or other notices.

The licence text decides; this summary does not replace it.

## What belongs here

A feature goes into `ee/` only if all three hold
([ADR 0020](../docs/adr/0020-ee-stays-proprietary.md#what-goes-into-ee)):

1. It is only useful to an organisation with someone who administers it, not
   to a single person or a household.
2. It is about running the instance for many people: central control of
   accounts and sign-in, compliance, organisation-wide communication, or
   connecting the instance to systems organisations use.
3. The core works fully without it.

Members, roles and groups, invitations, sharing, the owner's single sign-on,
two-factor authentication, passkeys, the link directory, the phone list and
every module of the core stay under MIT. A feature released in the core never
moves into `ee/`; a feature in `ee/` may move to the core at any time.

## Rules

1. The core never depends on `ee/`. Go packages below `internal/` must not
   import it, and the web app under `web/` must not import `ee/web/`; the
   architecture tests fail otherwise.
2. `ee/` is wired in the composition root only, in files with the build
   constraint `//go:build ee`. A build without that tag contains no code from
   this directory and is MIT only.
3. Code in `ee/` follows the same architecture as the core: one module per
   feature, the same four rings.
4. Pull requests from outside that touch `ee/` need a written agreement with
   the maintainer first; please open an issue before starting, see
   [CONTRIBUTING.md](../CONTRIBUTING.md#licence-of-contributions).

## Modules

Every module checks the licence key. Modules with their own screens and
routes implement `modules.Gated` with `licence.Gate`, so the core hides them
and refuses their routes without a valid key. Modules that hook into the
core or answer outside requests (`sso/`, `ldap/`, `teamwiki/`, `slackcmd/`,
`teamsbot/`) ask the licence for their feature on every use instead.

| Directory | Feature | Docs |
| --- | --- | --- |
| `licence/` | Licence keys ([#307](https://github.com/itadori26/waylight/issues/307)) | [licence-keys.md](../docs/licence-keys.md) |
| `teamtabs/` | Dashboard tabs shared by admins ([#295](https://github.com/itadori26/waylight/issues/295)) | [team-tabs.md](../docs/team-tabs.md) |
| `announcements/` | Notices for the whole instance ([#297](https://github.com/itadori26/waylight/issues/297)) | [announcements.md](../docs/announcements.md) |
| `absences/` | Absences ([#299](https://github.com/itadori26/waylight/issues/299)) | [absences.md](../docs/absences.md) |
| `teamcalendar/` | Calendars shared with a group ([#299](https://github.com/itadori26/waylight/issues/299)) | [team-calendar.md](../docs/team-calendar.md) |
| `sso/` | Single sign-on for members with group mapping ([#302](https://github.com/itadori26/waylight/issues/302)) | [single-sign-on-members.md](../docs/single-sign-on-members.md) |
| `audit/` | Audit log ([#303](https://github.com/itadori26/waylight/issues/303)) | [audit-log.md](../docs/audit-log.md) |
| `policies/` | Admin policies ([#304](https://github.com/itadori26/waylight/issues/304)) | [admin-policies.md](../docs/admin-policies.md) |
| `branding/` | Own name, logos and colours ([#305](https://github.com/itadori26/waylight/issues/305)) | [branding.md](../docs/branding.md) |
| `connections/` | Connected accounts per member ([#326](https://github.com/itadori26/waylight/issues/326)) | [connected-accounts.md](../docs/connected-accounts.md) |
| `teamwiki/` | Confluence, SharePoint/OneDrive, Notion, Slack and Teams in the wiki search ([#327](https://github.com/itadori26/waylight/issues/327)–[#331](https://github.com/itadori26/waylight/issues/331)) | [connected-accounts.md](../docs/connected-accounts.md) |
| `slackcmd/` | Slack slash command ([#333](https://github.com/itadori26/waylight/issues/333)) | [slack-command.md](../docs/slack-command.md) |
| `teamsbot/` | Search from the Microsoft Teams command box ([#334](https://github.com/itadori26/waylight/issues/334)) | [teams.md](../docs/teams.md) |
| `ldap/`, `scim/` | Directory sync: LDAP, Active Directory and SCIM ([#396](https://github.com/itadori26/waylight/issues/396)) | [directory-sync.md](../docs/directory-sync.md) |

## Frontend

The screens of `ee/` live in [`web/`](web/), a package of their own that the
web app loads as a plugin ([ADR 0021](../docs/adr/0021-ee-frontend.md)). The
server embeds `ee/web/dist` in builds with the tag `ee` and offers it through
`GET /api/plugins`. Features without a licence stay listed under
*Settings → Features for organisations*.

## Building and testing

```sh
cd ee/web && npm ci && npm run build      # writes ee/web/dist/plugin.js
cd ee/web && npm run format               # Prettier, checked in CI
go build -tags ee ./cmd/server            # a server with this directory
go test -tags ee ./cmd/server/... ./ee/...
```

Build the plugin before the server: the `ee` build embeds whatever is in
`ee/web/dist`.
