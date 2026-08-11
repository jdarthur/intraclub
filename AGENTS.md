# AGENTS.md

Intra-club recreational sports league web app: draft players into teams, run a
season of weekly head-to-head matches, track availability/lineups/scoring, and
manage club rules via commissioner proposals. Go/Gin backend + SvelteKit UI.
Work in progress — see `creation-sequence.md` for a feature checklist.

## Project
- Backend: Go 1.26 + [Gin](https://github.com/gin-gonic/gin) REST API under `/api`.
- Auth: JWT (ES512, keypair in `token.crt`/`token.key`) + email magic-link OTP.
- Storage: generic `database.Provider` with pluggable backends — **SQLite**
  (`modernc.org/sqlite`, migrations in `database/migrations/`) and in-memory
  (used by tests). Default is SQLite (single `intraclub.db` file); select via
  `--db` / `--db-path` / `INTRACLUB_DB_PATH`.
- Frontend: SvelteKit (Svelte 5) / Vite / TypeScript in `ui/`; proxies `/api` to
  backend on port 8080.
- Entry point: `main.go` (flag parsing, provider selection, route wiring).
- Docs: `api/authorization.md` (authz model), `creation-sequence.md` (build status).

## Commands
```sh
make standard        # build backend binary `main`
make tests           # go test ./... && go vet ./...
make watch           # live-reload via ./air
make e2e             # Playwright e2e from ui/ (starts backend + UI itself)
make e2e-ui          # Playwright e2e in interactive UI mode
make clean           # remove build artifacts

./main --dev-token   # run backend, dev mode (loopback-only, seeds sample data) on :8080
cd ui && npm install && npm run dev   # frontend dev server (port 5173)
cd ui && npm run check                # svelte-check + typecheck
```
Verify: `go build ./...` succeeds; Go toolchain is go1.26.

## Verification for UI changes
- **Always run `make e2e` before marking any UI change as finished.** The
  full Playwright suite runs headless with `fullyParallel` (see `ui/playwright.config.ts`)
  and covers the CRUD pages, login, and smoke flows end-to-end against a real
  backend + SQLite DB. A change that only passes `npm run check` (svelte-check/
  typecheck) is not enough — UI work is only "done" once the full parallel e2e
  suite is green.
- Run `make e2e` a few times if a change touches navigation/forms: the suite is
  parallel, so hydration/race flakes only surface across repeated full runs.

## Architecture
- `database/` — the core. `Provider` interface (`database_provider.go`) with
  generic helpers `GetOneById`/`GetAll`/`GetAllWhere`/`CreateOne`/`UpdateOne`/
  `DeleteOneById`. `CrudRecord` = `Type()` + `RecordId` + `Authorizable` +
  `DatabaseValidatable` + `NewRecord()`. SQLite impl in `sqlite_db.go`,
  access-control wrapper in `db_with_access_control.go`, migrations in `migrations/`.
- `model/` — one domain struct per file (facility, draft, season, team, user, …),
  each implementing the `CrudRecord` interface. `NewX()` constructor pattern.
- `api/` — Gin wiring. `CrudCommon[T]` auto-generates CRUD routes from a model;
  `RouteFamily[T]` for custom routes. `crud_common.go`, `api_route.go`,
  `api_auth.go` (JWT/token).
- `route/` — non-CRUD handlers: `route/user/` (self-register, verify email, whoami),
  CSV import.
- `mailer/` — email for verification / login tokens.
- `ui/` — SvelteKit frontend; Playwright e2e in `ui/e2e/`.

## Conventions
- **Every domain model** must implement the full `CrudRecord` interface: `Type()`
  (singular table name), typed `XxxId` wrappers with `RecordId()`, `GetId`/`SetId`,
  `StaticallyValid` (field-level checks, trims values) + `DynamicallyValid`
  (DB-backed existence checks), `EditableBy`/`AccessibleTo` (authz), `GetOwner`/
  `SetOwner`, `NewRecord()`, `NewX()` constructor. Model one domain type per file.
- **Authz is per-record**: `EditableBy`/`AccessibleTo` return `[]database.UserId`;
  use helpers `database.SysAdminAndUsers(...)` / `database.AccessibleToEveryone` /
  `database.EveryoneUserId`. See `api/authorization.md`.
- **Validation lifecycle**: `CreateOne`/`UpdateOne` run `StaticallyValid` then
  `DynamicallyValid` then `ValidateUniqueConstraint`, and fire optional
  `PreCreate`/`PostCreate`/`PreUpdate`/`PostUpdate`/`PreDelete`/`PostDelete` hooks.
- **Use generic DB helpers**, not raw provider calls: `database.CreateOne(ctx, db, rec)`,
  `database.GetAllWhere[T](ctx, db, filter)`, `database.ExistsById(ctx, db, &Type{}, id)`.
- **Uniqueness**: implement `UniquenessEquivalent(other) error` (e.g. `User`, `Facility`).
- Tests: `_test.go` beside each model; SQLite round-trip tests in `database/`.
  Uses `github.com/stretchr/testify`.
- Don't commit secrets; `token.crt`/`token.key` are the local dev JWT keypair.

## Notes
- Tickets/issues are tracked **upstream on GitHub** (`jdarthur/intraclub`), not in this repo.
- All changes must land on `main` **through a pull request** — no direct pushes to `main`.
- `gh` CLI (v2.97.0) is available and authorized for read/write against `origin` (`git@github.com:jdarthur/intraclub.git`) — use it for issue/PR workflows.

