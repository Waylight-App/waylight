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

## Frontend

The screens of `ee/` live in [`web/`](web/), a package of their own that the
web app loads as a plugin ([ADR 0021](../docs/adr/0021-ee-frontend.md)):

```sh
cd ee/web && npm ci && npm run build   # writes ee/web/dist/plugin.js
```

The server embeds `ee/web/dist` in builds with the tag `ee` and offers it
through `GET /api/plugins`. Features without a licence stay listed under
*Settings → Features for organisations*.

## Status

- `licence/`: licence keys ([#307](https://github.com/itadori26/trisse-home/issues/307)).
  Other modules here implement `modules.Gated` with `licence.Gate`, so the
  core hides them and refuses their routes without a valid key.

A server built with `go build -tags ee ./cmd/server` contains this
directory; the published builds do not, until the licence files are final
([#319](https://github.com/itadori26/trisse-home/issues/319)).
