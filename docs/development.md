# Development setup

This page gets a development copy of trisse-home running and explains the
everyday commands, the repository layout and the test suites. Plan about ten
minutes for the setup. The rules for pull requests are in
[CONTRIBUTING.md](https://github.com/itadori26/trisse-home/blob/main/CONTRIBUTING.md).

## Requirements

- Go in the version named in `go.mod`
- Node.js 22 or newer (the current LTS is recommended)
- `make`
- [golangci-lint](https://golangci-lint.run) v2 for `make lint`
- Docker, only if you want to build the container image

`make deps` installs the remaining tools, such as the TypeScript type
generator.

## First run

```sh
git clone https://github.com/itadori26/trisse-home.git
cd trisse-home
make deps
cp .env.example .env
sed -i "s|^SECRET_KEY=.*|SECRET_KEY=$(openssl rand -base64 32)|" .env
make run                  # terminal 1: backend on :8090, data in ./data
cd web && npm run dev     # terminal 2: Vite with hot reload
```

Open the address Vite prints. It forwards `/api` to the backend on port 8090.
Enter the setup token from the backend log and choose a password; the sample
data fills the dashboard. To skip the setup altogether, add `DEMO_MODE=true`
to `.env` ([demo mode](demo.md)).

## Everyday commands

| Command | What it does |
| --- | --- |
| `make generate` | Generates the TypeScript DTO types from the Go DTOs and the catalogue of server messages. Run it after changing a DTO and commit the result. |
| `make lint` | golangci-lint (including the architecture rules in depguard), ESLint, Prettier and svelte-check. |
| `make test` | Go tests with the race detector and the Vitest suite. |
| `make smoke` | Builds the binary and runs the Playwright end-to-end suite against a fresh instance. Run `cd web && npx playwright install chromium` once first. |
| `make build` | Builds the frontend and the single binary into `bin/`. |
| `make openapi` | Merges the OpenAPI fragments into `dist/openapi.yaml`. |
| `make new-module NAME=quotes` | Scaffolds a new module with all rings, tests and registrations. See [adding a module](adding-a-module.md). |
| `make ci` | Everything CI runs, apart from the end-to-end suite. Run it before you open a pull request. |
| `make help` | Lists every target. |

## Repository layout

```
cmd/server/              composition root: wiring only, one file per module
internal/
  platform/              frameworks and drivers: config, database, HTTP server, middleware, scheduler, crypto
  shared/                kernel (error kinds, Clock, IDGenerator) and adapter helpers
  modules/<name>/
    domain/              entities, value objects, rules              (ring 1)
    usecase/             use cases, input and output ports, models   (ring 2)
    adapter/controller/  HTTP handlers                               (ring 3)
    adapter/presenter/   response DTOs
    adapter/gateway/     SQLite repositories, API clients, migrations
    module.go            wires the module together
web/src/
  app/                   frontend composition root and widget registry
  features/<name>/       domain, usecase, adapter/api and ui per feature
  shared/                API client, generated DTO types, theme, UI components
web/e2e/                 Playwright suite: setup, general flows, features/<name>.spec.ts, two-factor last
api/                     OpenAPI description, one fragment per module
docs/                    this documentation
```

Source dependencies only point inwards. Linters and an architecture test
enforce this; [architecture](architecture.md) explains the rings and the
request flow.

## Tests

| Layer | How it is tested |
| --- | --- |
| Domain | Table-driven unit tests without dependencies. |
| Use cases | Fakes for every output port; no database, no HTTP. |
| Gateways | Integration tests against a temporary SQLite database or an `httptest` server. |
| Controllers | `httptest` with fake use cases. |
| API end to end | `cmd/server`: all modules wired together, with fake upstream services. Each module registers its own check. |
| Architecture | `internal/archtest` and depguard. |
| Frontend | Vitest for the domain and use cases of every feature. |
| Browser | Playwright: setup, sign-in, editing, every widget kind, accessibility checks with axe-core. |

## Budgets

The container image, the binary, the memory use at idle and the JavaScript
loaded on the first page view have upper limits in `budgets.json`. CI fails
when a change exceeds one. If a feature really needs more, raise the budget
in the same pull request and explain why
([ADR 0011](adr/0011-lean-by-default.md)).

## Documentation

Every module documents itself in `docs/modules/<name>.md` and its endpoints
in `api/modules/<name>.yaml`; a test fails when either is missing. Pages for
users and operators live in `docs/` and are listed in `mkdocs.yml`, which is
the navigation of the future documentation site. Preview it with
`pip install zensical && zensical serve`; `zensical build --strict` fails on
broken links. Write them in plain
English, for a reader with a task in mind.
