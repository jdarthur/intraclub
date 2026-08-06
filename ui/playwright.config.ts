import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: './e2e',
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	workers: process.env.CI ? 1 : undefined,
	reporter: 'html',
	use: {
		baseURL: 'http://127.0.0.1:5173',
		trace: 'on-first-retry'
	},
	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] }
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
