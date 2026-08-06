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

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.

---

## Backend prerequisites

The UI talks to a Go/Gin backend that lives at the repository root (`/api/...` routes). Start it before exercising the UI.

### 1. Start MongoDB

The backend stores data in MongoDB, provided via Docker:

```sh
docker compose up -d
```

This starts `mongo-findash` and maps the container's MongoDB to port `27018` on your host.

### 2. Prerequisites

- [Go](https://go.dev/dl/) installed and available in your `PATH`
- [Docker](https://docs.docker.com/engine/install/) (for MongoDB)

### 3. Run the backend server

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

### 4. Run the UI

Back in the `ui/` directory:

```sh
npm run dev
```

> Note: the UI fetches `/api/...` paths relative to its own origin, so you may need a dev proxy or to configure the API base URL to reach the backend on port `8080`.
