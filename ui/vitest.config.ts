import { defineConfig } from 'vitest/config';

export default defineConfig({
	test: {
		// Only pick up unit tests under src/ — the Playwright suites live in
		// e2e/*.spec.ts and run via `npm run test:e2e`.
		include: ['src/**/*.test.ts'],
		environment: 'node'
	}
});
