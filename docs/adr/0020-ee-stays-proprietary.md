# ADR 0020: `ee/` stays proprietary, and only organisation features go there

- Status: Proposed
- Date: 2026-11-15
- Issue: [#394](https://github.com/itadori26/trisse-home/issues/394), part of [#292](https://github.com/itadori26/trisse-home/issues/292)
- Supersedes in part: [ADR 0017](0017-open-core.md) (the licence of `ee/`
  and the rule for what goes there; the technical separation stays)

## Context

ADR 0017 split the repository into an MIT core and an `ee/` directory under
the Fair Core License with MIT as future licence (FCL-1.0-MIT). Under that
licence every version of a team feature becomes MIT two years after its
release. The maintainer does not want this. The intent is a clear line: the
core is open source and stays open source, the team features are a commercial
product and stay one. A licence that turns them into MIT after two years
blurs that line, and it makes every feature in `ee/` a countdown instead of
something the project can rely on to fund the work.

ADR 0017 also left the line between core and `ee/` to each issue. That has
already been argued once: members and groups (#294) were first counted as a
team feature, then moved to the core on 2026-11-15, because a family needs
several accounts just as much as an office does. Without a written rule the
same question comes back with every new feature.

`ee/` is still empty. Nothing has been released under FCL-1.0-MIT, so the
licence can change now without taking anything away from anyone.

## Decision

### Licence of `ee/`

- **`ee/` gets no future licence.** Its code never becomes MIT or any other
  open-source licence by the passage of time. It is source-available: anyone
  can read it, but it is not open source and the project does not call it
  that.
- **Recommended licence: Elastic License 2.0 (ELv2).** It is short, widely
  used and already known to legal departments. It allows use, change and
  redistribution, and forbids three things: offering the software to others
  as a hosted or managed service, moving or circumventing the licence key,
  and removing licence notices. That matches the licence key from
  [#307](https://github.com/itadori26/trisse-home/issues/307) without a
  custom text. Alternatives are listed at the end.
- The final choice of text is made together with the legal review in
  [#319](https://github.com/itadori26/trisse-home/issues/319), before the
  first feature in `ee/` is sold. If the review recommends another licence
  without a future grant, that licence replaces ELv2 without a new ADR.

### What goes into `ee/`

A feature may go into `ee/` only if all three hold:

1. It is only useful to an organisation with someone who administers it,
   not to a single person or a household.
2. It is about running the instance for many people: central control,
   compliance, or connecting the instance to systems that organisations use
   (directories, company wikis, company chat).
3. The core works fully without it. Nothing in the core waits for `ee/`, and
   no core feature is held back to make `ee/` more attractive.

As of today this covers:

| Issue | Feature |
| ----- | ------- |
| [#295](https://github.com/itadori26/trisse-home/issues/295) | Team tabs maintained by admins |
| [#302](https://github.com/itadori26/trisse-home/issues/302) | Single sign-on for all accounts with group mapping |
| [#303](https://github.com/itadori26/trisse-home/issues/303) | Audit log with export and retention |
| [#304](https://github.com/itadori26/trisse-home/issues/304) | Admin policies (required second factor, session length, allowed modules) |
| [#305](https://github.com/itadori26/trisse-home/issues/305) | Branding: own name, logo and colours |
| [#307](https://github.com/itadori26/trisse-home/issues/307) | The licence key itself |
| [#326](https://github.com/itadori26/trisse-home/issues/326) | Connected accounts per member |
| [#327](https://github.com/itadori26/trisse-home/issues/327)–[#331](https://github.com/itadori26/trisse-home/issues/331) | Workplace search: Confluence, SharePoint/OneDrive, Notion, Slack, Teams |

These stay in the MIT core, and are named here because they came close:

- members, roles and groups (#294) and invitation links and password reset
  without shell access (#296);
- data ownership and sharing with a group or with everyone
  ([ADR 0018](0018-data-ownership.md), #301);
- single sign-on for one owner (ADR 0013), two-factor authentication and
  passkeys (ADR 0009, ADR 0014);
- every module that exists today, and every widget a team uses but a
  household could use too (announcements, phone list, absences, link
  directory: #297–#300).

### Nothing moves backwards

A feature that has been released in the core is never moved into `ee/` and
never put behind the licence key. A feature in `ee/` may move to the core at
any time; that is the only direction allowed.

### How the project describes itself

"trisse-home is open source (MIT). Optional features for organisations are
available under a commercial licence." The README and the website lead with
the open-source core and name `ee/` plainly as not open source. They do not
say "fully open source" or put an OSI badge on the whole repository.

### What stays from ADR 0017

The technical rules are unchanged: the core never imports `ee/`, `ee/` is
wired only in the composition root behind the build tag `ee`, a build without
the tag is MIT only, and `internal/archtest` enforces both. Contributions
outside `ee/` stay MIT inbound and outbound with nothing to sign;
contributions to `ee/` need an agreement with the maintainer.

## Consequences

- The open-source label of the core is not affected: the MIT part is as open
  as before. The protection of `ee/` no longer runs out, so a competitor
  cannot wait two years and then sell the team features.
- `ee/` is now plainly proprietary. Some people in the self-hosting community
  will see that as a step away from open source, even though the core is
  unchanged. The list of what stays in the core and the "nothing moves
  backwards" rule are the answer to that, and both have to be kept.
- Outside contributions to `ee/` will be rare. That is acceptable: the
  features there are paid work.
- When this ADR is accepted, a follow-up change, done together with #319,
  replaces `ee/LICENSE.md` and updates `LICENSE`, `ee/README.md`,
  `CONTRIBUTING.md`, `README.md` and `docs/faq.md`, and marks ADR 0017 as
  superseded in part. Until then the files still name FCL-1.0-MIT, which is
  harmless because `ee/` contains no code.
- Whether published images are built with the `ee` tag is still decided with
  the first feature in `ee/`, as ADR 0017 says.
- If the project later moves to a new repository, the licence files move
  with the code; nothing in this decision is tied to the current repository.

## Alternatives considered

- **Keep FCL-1.0-MIT.** Rejected by the maintainer: every feature becomes MIT
  after two years.
- **FCL-1.0 with Apache 2.0 as future licence.** Same objection; only the
  future licence differs.
- **Business Source License 1.1.** Requires a change date and a change
  licence; a far-away date is possible but works against the intent and
  surprises readers who know BSL as "open source later".
- **A custom commercial licence** in the style of other open-core projects
  (production use only with a subscription). Gives the most control, but a
  custom text needs more legal work and is harder for customers to assess
  than a known licence.
- **Keeping `ee/` in a private repository.** Rejected in ADR 0017 for the
  same reasons as before: two code bases, and paying organisations could not
  inspect the code they run.
