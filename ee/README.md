# ee/ – team features

Everything in this directory is licensed under the
[Fair Core License 1.0 with MIT as future licence](LICENSE.md), not under the
MIT licence that covers the rest of the repository. The decision and its
reasons are in [ADR 0017](../docs/adr/0017-open-core.md).

## What that means

- You may read, run, change and share this code for any purpose except
  offering it to others as a commercial product or service that competes with
  trisse-home.
- Features in this directory sit behind a licence key. The licence does not
  allow removing or bypassing that check.
- Each released version becomes MIT two years after its release.

The licence text decides; this summary does not replace it.

## What belongs here

The features for teams tracked in
[#292](https://github.com/itadori26/trisse-home/issues/292). The start page
for one person or one household stays in the core under MIT, and nothing that
exists today moves here.

## Rules

1. Core code never depends on `ee/`. Go packages below `internal/` must not
   import it; the architecture test fails otherwise.
2. `ee/` is wired in the composition root only, in files with the build
   constraint `//go:build ee`. A build without that tag contains no code from
   this directory and is MIT only.
3. Code in `ee/` follows the same architecture as the core: one module per
   feature, the same four rings.
4. Pull requests from outside that touch `ee/` need an agreement with the
   maintainer first, see [CONTRIBUTING.md](../CONTRIBUTING.md#licence-of-contributions).

## Status

Empty. The licence key check comes with
[#307](https://github.com/itadori26/trisse-home/issues/307).
