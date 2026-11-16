# ADR 0021: Frontend code of `ee/` and the two builds

- Status: Proposed
- Date: 2026-11-16
- Issue: [#307](https://github.com/itadori26/trisse-home/issues/307), part of [#292](https://github.com/itadori26/trisse-home/issues/292)
- Builds on: [ADR 0017](0017-open-core.md), [ADR 0020](0020-ee-stays-proprietary.md)

## Context

ADR 0017 keeps the Go code of `ee/` out of the core: the core never imports
it, it is wired only in the composition root behind the build tag `ee`, and a
build without the tag is MIT only. It says nothing about the frontend. The
first `ee/` features (#307, #297, #305, #295) all have screens, so the web
app needs the same separation.

## Decision

### Where the code lives

- Frontend code of `ee/` lives in `ee/web/`, one folder per feature with the
  same layout as `web/src/features/<name>` (`domain/`, `usecase/`,
  `adapter/api/`, `ui/`, `i18n/`).
- `ee/web/register.ts` exports what the features contribute, in the same
  shapes as the core's registration files: widget definitions, settings
  sections, services and search contributions.

### How it gets into a build

- The core imports `ee` code from exactly one place, the virtual module
  `virtual:ee` in `web/src/app/`. A small Vite plugin resolves it to
  `ee/web/register.ts` when the build runs with `EE=1`, and to an empty
  registration otherwise.
- `npm run build` (and therefore `make build`, the Docker image and every
  release) stays MIT only: no file from `ee/web/` is in the bundle.
- `npm run build:ee` and `make build-ee` build the frontend with `EE=1` and
  the server with `-tags ee`. CI builds and tests both variants.

### What shows up at runtime

- `ee/` backend modules register only in builds with the tag `ee`.
- The server leaves an `ee/` module out of `GET /api/modules` while it is
  not licensed (#307). The frontend already hides widgets, settings and
  search answers of modules that are not listed, so it needs no licence
  logic of its own. Only the licence settings section asks
  `GET /api/licence`.

### Rules, checked by tests

- Nothing under `web/src/` imports from `ee/web/` other than through
  `virtual:ee`; `ee/web/` may import from `web/src/shared/`.
- Vitest runs the tests in `ee/web/**` as well.
- Texts of `ee/web/` features stay in their own `i18n/` folders.

## Consequences

- One codebase, two bundles. The MIT bundle does not grow with `ee/`
  features.
- `ee/` features reuse the core's widget, settings and search registration
  unchanged; adding one does not touch shared lists.
- CI time grows by one frontend build and one Go build with `-tags ee`.

## Alternatives considered

- **`ee/` frontend inside `web/src/features/`:** simplest, but it puts code
  under another licence into the MIT tree and into every bundle.
- **A separate frontend package for `ee/`:** a second app and dependency set
  to maintain for a handful of screens.
- **Runtime plugins loaded from the server:** a loader, versioning and a
  content security policy exception for a problem a build flag solves.
