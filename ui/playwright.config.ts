import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: './e2e',
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	workers: process.env.CI ? 1 : undefined,
	reporter: 'html',
	// Visual baselines (#201) are committed next to their spec
	// (e2e/*.spec.ts-snapshots/) and are platform-specific — they are only
	// valid for the CI browser build (see ui/README.md). The default template
	// appends `-{projectName}`; these projects each own their spec file, so
	// drop it to keep baseline names stable.
	snapshotPathTemplate: '{testDir}/{testFileDir}/{testFileName}-snapshots/{arg}{ext}',
	use: {
		baseURL: 'http://127.0.0.1:5173',
		trace: 'on-first-retry'
	},
	projects: [
		{
			// Visual baselines run FIRST: the chromium project depends on this
			// one, so these tests execute against the freshly-wiped DB that
			// `make e2e` leaves behind, before any spec has created a season.
			// That is what makes the signed-in dashboard baseline the genuine
			// empty state (#201).
			name: 'visual-baselines',
			testMatch: /visual\.spec\.ts/,
			use: { ...devices['Desktop Chrome'] }
		},
		{
			name: 'chromium',
			testIgnore: /(visual|mobile)\.spec\.ts/,
			dependencies: ['visual-baselines'],
			use: { ...devices['Desktop Chrome'] }
		},
		{
			// Mobile subset (#201): a dedicated spec covering the nav sheet, one
			// list page and the landing page at 375px — not all 27 specs, most of
			// which are wide-table / draft-board flows that would not pass at
			// this width (the mobile-nav spec from #200 stays on the chromium
			// project with its own viewport).
			name: 'Mobile Chrome',
			testMatch: /mobile\.spec\.ts/,
			use: { ...devices['Pixel 5'] }
		}
	],
	webServer: [
		{
			// Build to tmp/ (gitignored, same convention as .air.toml) rather than using
			// `go run`, which leaves the compiled child holding :8080 on teardown.
			// --dev-token seeds dev data and returns login tokens in API responses.
			command: 'go build -o ./tmp/intraclub-e2e . && ./tmp/intraclub-e2e --dev-token',
			cwd: '..',
			// Every route is mounted under /api, so `/` 404s and never reads as ready.
			url: 'http://127.0.0.1:8080/api/score_counting_types',
			reuseExistingServer: !process.env.CI,
			timeout: 120_000
		},
		{
			command: 'npm run dev -- --host 127.0.0.1',
			url: 'http://127.0.0.1:5173',
			reuseExistingServer: !process.env.CI,
			timeout: 120_000
		}
	]
});
