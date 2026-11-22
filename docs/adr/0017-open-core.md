# ADR 0017: Open core – MIT core, team features in `ee/` under the Fair Core License

- Status: Accepted, superseded in part by [ADR 0020](0020-ee-stays-proprietary.md) (licence of `ee/` and its scope)
- Date: 2026-11-14

## Context

Waylight is MIT licensed. Features for organisations
([#292](https://github.com/Waylight-App/waylight/issues/292)) get terms of
their own. Changing the licence of the whole project was considered and
rejected: the published versions stay MIT in any case, and a copyleft or
source-available core would add friction for every contributor.

## Decision

- **Everything outside `ee/` stays MIT.** That includes all modules that exist
  today and everything a single person or household needs.
- **`ee/` holds the team features** under the Fair Core License 1.0 with MIT
  as future licence (FCL-1.0-MIT): the source is public, use is free except
  for a competing commercial offer, the features sit behind a licence key
  that may not be bypassed, and each version becomes MIT two years after its
  release.
- **The core never depends on `ee/`.** Packages below `internal/` must not
  import it. `ee/` is wired only in the composition root, in files with the
  build constraint `//go:build ee`, so a build without the tag contains no
  code from `ee/` and is MIT only. `internal/archtest` enforces both rules.
- **Contributions** outside `ee/` are MIT, inbound as outbound, with nothing
  to sign. Contributions to `ee/` need an agreement with the maintainer.
- The licence key is verified offline
  ([#307](https://github.com/Waylight-App/waylight/issues/307)); the principle
  "no outbound request you did not configure" holds for licensed instances too.

## Consequences

- The repository carries two licences; `LICENSE` states which applies where.
  Licence scanners may report the repository as "other" instead of MIT.
- The MIT core may be used and redistributed by anyone under MIT. Only
  `ee/` has other terms.
- Features needed by both sides (for example the ownership of records) must
  live in the core. Where the line runs is decided per feature in its issue.
- Whether the published images and packages are built with the `ee` tag is
  decided with the first feature in `ee/`. Until then every build is MIT only
  and `packaging/nfpm.yaml` keeps `license: MIT`.
- Rejected alternatives: AGPL for the core (does not stop hosting by others
  and keeps some organisations away), FSL for everything (no longer open
  source), a private repository for the team features (two code bases, and
  the code could not be inspected by the organisations that run it).
