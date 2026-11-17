# ADR 0020: `ee/` stays proprietary, and only organisation features go there

- Status: Accepted (2026-11-15, by the maintainer)
- Date: 2026-11-15
- Issue: [#394](https://github.com/itadori26/trisse-home/issues/394), part of [#292](https://github.com/itadori26/trisse-home/issues/292)
- Supersedes in part: [ADR 0017](0017-open-core.md) (the licence of `ee/`
  and the rule for what goes there; the technical separation stays)

## Context

ADR 0017 put `ee/` under the Fair Core License with MIT as future licence
(FCL-1.0-MIT), so every version of a feature in `ee/` would become MIT two
years after its release. It also left the line between core and `ee/` to
each issue. `ee/` is still empty, so nothing has been released under that
licence yet.

## Decision

### Licence of `ee/`

- **`ee/` has no future licence.** Its code does not become MIT or any other
  open-source licence over time. It is source-available: anyone can read it,
  but it is not open source and the project does not call it that.
- **Licence: Elastic License 2.0 (ELv2)**, unless the review in
  [#319](https://github.com/itadori26/trisse-home/issues/319) leads to
  another licence without a future grant. ELv2 allows use, change and
  redistribution, and forbids offering the software as a hosted or managed
  service, moving or circumventing the licence key
  ([#307](https://github.com/itadori26/trisse-home/issues/307)), and removing
  licence notices.

### What goes into `ee/`

A feature may go into `ee/` only if all three hold:

1. It is only useful to an organisation with someone who administers it,
   not to a single person or a household.
2. It is about running the instance for many people: central control of
   accounts and sign-in, compliance, organisation-wide communication, or
   connecting the instance to systems organisations use (directories,
   company wikis, company chat).
3. The core works fully without it.

As of today this covers:

| Issue | Feature |
| ----- | ------- |
| [#295](https://github.com/itadori26/trisse-home/issues/295) | Team tabs maintained by admins |
| [#297](https://github.com/itadori26/trisse-home/issues/297) | Announcements |
| [#299](https://github.com/itadori26/trisse-home/issues/299) | Absences and team calendar |
| [#302](https://github.com/itadori26/trisse-home/issues/302) | Single sign-on for all accounts with group mapping |
| [#303](https://github.com/itadori26/trisse-home/issues/303) | Audit log with export and retention |
| [#304](https://github.com/itadori26/trisse-home/issues/304) | Admin policies (required second factor, session length, allowed modules) |
| [#305](https://github.com/itadori26/trisse-home/issues/305) | Branding: own name, logo and colours |
| [#307](https://github.com/itadori26/trisse-home/issues/307) | The licence key itself |
| [#326](https://github.com/itadori26/trisse-home/issues/326) | Connected accounts per member |
| [#327](https://github.com/itadori26/trisse-home/issues/327)–[#331](https://github.com/itadori26/trisse-home/issues/331) | Workplace search: Confluence, SharePoint/OneDrive, Notion, Slack, Teams |
| [#332](https://github.com/itadori26/trisse-home/issues/332)–[#334](https://github.com/itadori26/trisse-home/issues/334) | Announcements to Slack/Teams channels, Slack app, Teams app |
| [#346](https://github.com/itadori26/trisse-home/issues/346) | Live presence widget |
| [#396](https://github.com/itadori26/trisse-home/issues/396) | Directory sync: LDAP, Active Directory and SCIM |

These stay in the MIT core:

- members, roles and groups (#294), invitation links and password reset
  without shell access (#296);
- data ownership and sharing with a group or with everyone
  ([ADR 0018](0018-data-ownership.md), #301);
- single sign-on for the owner (ADR 0013), two-factor authentication and
  passkeys (ADR 0009, ADR 0014);
- the link directory (#298) and the phone list (#300);
- every module that exists today.

### Nothing moves backwards

A feature released in the core is never moved into `ee/` and never put
behind the licence key. A feature in `ee/` may move to the core at any time.

### How the project describes itself

"trisse-home is open source (MIT). Optional features for organisations are
available under a commercial licence." The documentation names `ee/` plainly
as not open source and does not describe the whole repository as open
source.

### What stays from ADR 0017

The core never imports `ee/`, `ee/` is wired only in the composition root
behind the build tag `ee`, a build without the tag is MIT only, and
`internal/archtest` enforces both. Contributions outside `ee/` are MIT
inbound and outbound with nothing to sign; contributions to `ee/` need an
agreement with the maintainer.

## Consequences

- The MIT core is unchanged. Features in `ee/` stay under their licence
  without an end date.
- `ee/LICENSE.md` holds the ELv2 text with Tristan Leoluca Farinella as
  licensor; `LICENSE`, `ee/README.md`, `CONTRIBUTING.md`, `README.md` and
  `docs/faq.md` name ELv2.
- Whether published images are built with the `ee` tag is decided with the
  first feature in `ee/`, as in ADR 0017.

## Alternatives considered

- **FCL-1.0-MIT or FCL-1.0-ALv2:** both turn each version into open source
  after two years.
- **Business Source License 1.1:** requires a change date and a change
  licence.
- **A custom commercial licence:** a new text instead of a known one.
- **A private repository for `ee/`:** rejected in ADR 0017.
