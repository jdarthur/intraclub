# intraclub

A web application for running and managing an intra-club recreational sports
league. The club organizes players into teams through a **draft**, then runs a
**season** of weekly head-to-head matches, tracks availability, lineups, and
scoring, and manages the club's rules through commissioner proposals.

> **Note:** this is a work in progress. Many features are partially implemented;
> see [`creation-sequence.md`](./creation-sequence.md) for a running checklist
> of what's done and what's still open.

## Overview

The system models the full lifecycle of an intra-club season:

1. **Users & auth** — self-registration with email verification, "magic link"
   one-time-password login, and JWT-based authentication.
2. **League setup** — facilities, rating types, formats (how drafted players are
   grouped by skill), and playoff structures.
3. **Draft** — a commissioner starts a draft, captains take turns selecting
   players, and players can be pre-graded. The completed draft seeds a season.
4. **Season** — made up of weeks with schedules and team matchups, a playoff
   structure, availability input, weekly lineups, and recorded individual
   matches with scoring.
5. **Club governance** — rulesets, rule amendments, and commissioner proposals.

The domain is described in detail in [`creation-sequence.md`](./creation-sequence.md),
and the authorization model is documented in [`api/authorization.md`](./api/authorization.md).

## Tech stack

**Backend**

- [Go](https://go.dev/) with the [Gin](https://github.com/gin-gonic/gin) web
  framework — the REST API lives under `/api`.
- [JWT](https://github.com/golang-jwt/jwt) (ECDSA P-521 / ES512 keypair) for
  stateless auth, with email-based one-time-password ("magic link") login via
  [go-msgauth](https://github.com/emersion/go-msgauth).
- Generic, type-safe CRUD abstraction over a `database.Provider` interface
  (`database/database_provider.go`), with pluggable backends:
  - **SQLite** via [modernc.org/sqlite](https://gitlab.com/cznic/sqlite)
    (pure-Go, no CGO, compiles into the static binary) with a migration runner
    (`database/migrations/`). This is the default and recommended backend —
    the whole database is a single file, so backup is just copying it.
  - An in-memory provider used for tests and ephemeral local runs.
- Provider selection is configurable via the `--db` / `--db-path` flags or the
  `INTRACLUB_DB_PATH` environment variable.

**Frontend**

- [SvelteKit](https://kit.svelte.dev/) (Svelte 5) built with
  [Vite](https://vitejs.dev/), TypeScript, and [Playwright](https://playwright.dev/)
  for end-to-end tests — all in the [`ui/`](./ui) directory.

## Repository layout

```
api/          Gin route handlers, auth, and generic CRUD route wiring
database/     Database provider abstraction, migrations, SQLite & access control
model/        Domain models (draft, season, schedule, team, user, ...) + tests
route/        Non-CRUD HTTP handlers (CSV import, self-registration, verify email)
mailer/       Email sending for verification / login tokens
ui/           SvelteKit frontend (see ui/README.md)
main.go       Server entrypoint & flag parsing
```

## Getting started

### Prerequisites

- [Go](https://go.dev/dl/) (1.26)
- [Node.js](https://nodejs.org/) and npm (for the UI)

No database service is required — the default SQLite provider stores everything
in a single local file.

### Backend

From the repository root:

```sh
# build a binary named `main` and run it in dev mode
make standard
./main --dev-token
```

The `--dev-token` flag enables development token mode (loopback-only, bypasses
auth gating) and seeds sample data. The API listens on `http://127.0.0.1:8080`.

### Simulating a slow connection

Local testing on `127.0.0.1` makes every database call appear instantaneous.
`--slow-mode` injects artificial latency into every API request to simulate a
slow / high-RTT (edge-to-cloud) connection, which is useful for surfacing UX
issues (missing skeletons/loading spinners), finding query chains that should
be composite endpoints, and shaking out timing assumptions in the e2e suite:

```sh
./main --dev-token --slow-mode                       # +500ms per request (default)
./main --dev-token --slow-mode --slow-mode-latency 2s
INTRACLUB_SLOW_MODE=1 INTRACLUB_SLOW_MODE_LATENCY=1s ./main --dev-token
```

The `--slow-mode` flag / `INTRACLUB_SLOW_MODE` env var enable the delay, and
`--slow-mode-latency` / `INTRACLUB_SLOW_MODE_LATENCY` set the per-request delay
(default 500ms). Flags take precedence over environment variables.

### Database: SQLite (default)

The server defaults to the **SQLite** provider, which stores everything in a
single file (default `intraclub.db` in the working directory). No external
database service is required.

```sh
./main                          # uses ./intraclub.db
./main --db sqlite --db-path /path/to/intraclub.db
INTRACLUB_DB_PATH=/path/to/intraclub.db ./main --db sqlite
```

For an ephemeral run that keeps data only in memory:

```sh
./main --db memory
```

**Migrations** run automatically at startup: ordered SQL scripts in
`database/migrations/` are applied against the database, and applied versions
are tracked in a `schema_migrations` table, so each deployment only runs the new
ones. No manual step is needed to create or upgrade the schema.

**Backup & restore** is simply copying the `.db` file:

```sh
cp intraclub.db intraclub.backup.db    # backup
cp intraclub.backup.db intraclub.db    # restore
```

### Frontend

See [`ui/README.md`](./ui/README.md) for full instructions. In short:

```sh
cd ui
npm install
npm run dev
```

The UI proxies `/api` to the backend on port `8080` in development (see
`ui/vite.config.ts`).

## Common commands

The [`Makefile`](./Makefile) wraps the common tasks:

| Command            | Description                                            |
| ------------------ | ------------------------------------------------------ |
| `make standard`    | Build the backend binary (`main`)                      |
| `make tests`       | Run Go unit tests and `go vet ./...`                   |
| `make watch`       | Live-reload the backend with `air`                     |
| `make e2e`         | Run Playwright end-to-end tests (starts both servers)  |
| `make e2e-ui`      | Run Playwright e2e tests in interactive UI mode        |
| `make clean`       | Remove build artifacts                                  |

## Testing

- **Backend unit tests:** `make tests` (or `go test ./...`) — extensive `_test.go`
  coverage alongside each model.
- **Frontend e2e:** Playwright tests in [`ui/e2e/`](./ui/e2e), run via `make e2e`.
