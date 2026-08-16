# UI

This directory contains the SvelteKit frontend for the application. It is a Svelte 5 project powered by [`sv`](https://github.com/sveltejs/cli) and built with Vite.

## Prerequisites

- [Node.js](https://nodejs.org/) and npm installed

## Installation

From the `ui/` directory:

```sh
npm install
```

## Developing

Start the development server:

```sh
npm run dev
```

Or start the server and open the app in a new browser tab:

```sh
npm run dev -- --open
```

## Building

Create a production version of the app:

```sh
npm run build
```

Preview the production build:

```sh
npm run preview
```

## Checking

Run type checking and Svelte validation:

```sh
npm run check
```

## Testing

Playwright is used for end-to-end (UI) automation. Tests live in the `e2e/` directory.

Most UI tests need the Go API as well as the UI, so Playwright starts both servers itself:

- the Vite dev server on `http://127.0.0.1:5173`
- the Go backend on `http://127.0.0.1:8080`, built to `../tmp/intraclub-e2e` and run with
  `--dev-token` so that dev data is seeded and login tokens are returned in API responses

The dev server proxies `/api` to the backend (see `vite.config.ts`), so the app's relative
`fetch('/api/...')` calls work unchanged under test.

First-time setup requires Go on your `PATH` and the Playwright browser binaries:

```sh
npx playwright install chromium
```

Run the E2E tests (this starts both servers automatically):

```sh
npm run test:e2e
```

Or open the interactive Playwright UI:

```sh
npm run test:e2e:ui
```

### Visual regression baselines

`e2e/visual.spec.ts` and `e2e/mobile.spec.ts` assert `toHaveScreenshot()` on a
small, high-signal set of pages — the signed-out hero, the signed-in empty
dashboard, the facilities list (seeded row), the new-facility form, the
organization detail tabs, and the mobile nav sheet / list / landing page — each
in light **and** dark, with the theme driven by Playwright's `colorScheme`
emulation rather than the theme toggle.

Baseline PNGs are **committed** next to their spec in
`e2e/visual.spec.ts-snapshots/` and `e2e/mobile.spec.ts-snapshots/`. They are
**platform-specific**: they are only valid for the CI browser build (Linux +
bundled Chromium), so regenerate them on the same platform CI runs on:

```sh
# Regenerate all baselines (writes missing/changed PNGs, starts both servers):
npm run test:e2e -- --update-snapshots
```

To regenerate just one project (e.g. after touching only mobile styles):

```sh
npm run test:e2e -- --project="Mobile Chrome" --update-snapshots
npm run test:e2e -- --project=visual-baselines --update-snapshots
```

Notes:

- The `visual-baselines` project runs **first** (the `chromium` project depends
  on it), against the DB freshly wiped by `make e2e` — that is what makes the
  signed-in dashboard baseline the genuine empty state. Don't reorder it.
- Report (`playwright-report/`) and result (`test-results/`) artefacts are
  gitignored; the baseline directories are deliberately committed.
- When you intentionally restyle a covered page, update its baselines in the
  same PR — never "fix" a failing visual test by deleting baselines.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.

---

## Backend prerequisites

The UI talks to a Go/Gin backend that lives at the repository root (`/api/...` routes). Start it before exercising the UI.

### 1. Prerequisites

- [Go](https://go.dev/dl/) installed and available in your `PATH`

No database service is required — the backend uses the default SQLite provider
and stores data in a single `intraclub.db` file.

### 2. Run the backend server

From the repository **root** (not `ui/`):

```sh
# build a binary named `main` and run it
make standard
./main --dev-token
```

The `--dev-token` flag enables development token mode and seeds sample data, which is useful while developing the UI.

Alternatively, run it directly with Go:

```sh
go run main.go --dev-token
```

The API server listens on `http://127.0.0.1:8080`.

### 3. Run the UI

Back in the `ui/` directory:

```sh
npm run dev
```

> Note: the UI fetches `/api/...` paths relative to its own origin, so you may need a dev proxy or to configure the API base URL to reach the backend on port `8080`.
