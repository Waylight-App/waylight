# Adding a module

A module is a feature with its own backend rings, frontend feature folder,
settings, documentation and tests. This page is the reference for what a
module consists of. If you would rather learn by building one, follow the
[developer guide](developer-guide.md), which writes a small widget step by
step in about 30 minutes.

## Start with the generator

```sh
make new-module NAME=quotes TITLE="Quotes"
```

The generator creates every file listed below with working placeholder code
and tests, and registers the module through files of its own. Nothing in a
shared list needs editing, so two modules developed in parallel do not
conflict. [Modules](modules/README.md) explains the convention.

The steps below use a `quotes` module that shows a random quote as the
example.

## Backend

1. **Domain**, `internal/modules/quotes/domain/`. The entity `Quote` with a
   constructor that enforces its invariants, and domain errors that wrap the
   kinds from `internal/shared/kernel`, such as
   `kernel.NewPublicError(kernel.ErrNotFound, "…")`. Only the standard
   library and the kernel may be imported here. Cover the rules with table
   tests.

2. **Use cases**, `internal/modules/quotes/usecase/`. One input-port
   interface per use case (`RandomQuote`), the output ports it needs
   (`QuoteRepository`, `kernel.Clock`), request and response models, and the
   implementations. Use cases never return entities; they map them to
   response models. Test them with fakes for the ports.

3. **Gateway**, `adapter/gateway/`. Implementations of the output ports:
   SQLite through `database/sql`, HTTP clients injected from outside.
   Migrations go into `adapter/gateway/migrations/0001_init.sql`, embedded
   and exposed through `Migrations() fs.FS`. Map rows to entities and back,
   and write integration tests with `dbtest.Open`.

4. **Presenter**, `adapter/presenter/`. JSON DTOs and the functions that map
   response models to them.

5. **Controller**, `adapter/controller/`. Request DTOs in `dto.go` and
   handlers that decode with `httpapi.DecodeJSON`, call the input port and
   answer with `httpapi.WriteJSON` or `httpapi.WriteDomainError`. Test them
   with `httptest` and fake use cases.

6. **Module**, `internal/modules/quotes/module.go`. Embed `modules.Base`,
   build gateway, use cases and handlers in `New(...)` and implement `Name`,
   `DataKind`, and `Migrations`, `RegisterRoutes`, `Jobs` and `WidgetKinds`
   as needed. `DataKind` is required and says who owns the module's data
   ([ADR 0018](adr/0018-data-ownership.md)): `InstanceData` for what admins
   manage for everyone, `PersonalSettings` for one row per account,
   `PersonalContent` for records people create themselves. A
   personal-content module adds `owner_id` and `visibility` to its tables and
   filters every query with `sqlutil.VisibleTo` (reads) or `sqlutil.OwnedBy`
   (changes); `internal/archtest` refuses a query that forgets it. Add the
   module to the list in `cmd/server/datakind_test.go` and the table in the
   ADR.
   Register everything that needs a session through `r.Protected`. For
   sample data, implement `modules.Seeder` with an idempotent use case.

7. **Composition root**, `cmd/server/module_quotes.go`. Registers the module
   with `registerModule` in `init()`. Add the module name to the list in
   `.env.example`.

8. **End-to-end check**, `cmd/server/api_quotes_test.go`. Registers a check
   with `registerModuleE2E`: a fake upstream, a configuration override and a
   request against the fully wired server.

## Frontend

9. **Types.** Run `make generate`. The generator finds the presenter and
   controller DTO packages by convention and writes the TypeScript types.

10. **Feature folder**, `web/src/features/quotes/`, with `domain/`,
    `usecase/` (the port interface and the use case class), `adapter/api/`
    (implements the port and maps DTOs to domain types), `ui/` (the widget
    and an optional settings component), `i18n/en.ts` and `i18n/de.ts`, and
    an `index.ts` that exports the public API.

11. **Registration**, each in a file of its own:
    `web/src/app/services/quotes.ts` (the service in the container),
    `web/src/app/widgets/quotes.ts` (the widget, its picker category and
    icon) and, if the module has settings, `web/src/app/settings/quotes.ts`.
    A settings section without an `area` lands in *Connections & services*;
    set `area` (`account`, `appearance`, `notifications`, `dashboard`,
    `people`, `data` or `system`) when the section is about something else,
    for example `dashboard` for settings that only change what a widget
    shows. `keywords` (English and German synonyms) and `fields` (message
    keys of headings or labels inside the section) make the section and its
    settings findable in the settings search. Answers in the quick search go into `web/src/app/search/quotes.ts`.

12. **Browser test**, `web/e2e/features/quotes.spec.ts`, if the widget does
    something worth clicking through. Import `account` and `password` from
    `../support` and log in at the start; `openSettings(page, area, section)`
    opens a settings section by the names on screen; the account already exists,
    because the `setup` project runs first. Leave shared state as you found
    it, or work on a tab of your own, since all specs use one server. The
    file is new, so it never conflicts with other open pull requests.

## Documentation and API

13. **Module page**, `docs/modules/quotes.md`: what the module does, its
    configuration, the outside requests it makes and its endpoints. Add a
    line to the [feature overview](features.md) as well.

14. **OpenAPI**, `api/modules/quotes.yaml`: tags, paths and schemas of the
    module. `make openapi` merges and checks the fragments.

15. **Decision record.** If the module introduces a notable decision, such
    as a new kind of outside access, write an ADR in `docs/adr/`.

Run `make ci` and `make smoke` before you open the pull request. The
checklist for integrations with outside services, covering timeouts,
secrets and outbound requests, is at the end of the
[developer guide](developer-guide.md#checklist-for-a-new-integration).
