import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	// The sveltekit plugin provides the `$lib` / `$app` aliases and compiles
	// `.svelte.ts` modules, so unit tests can import from the real API client
	// modules (e.g. src/lib/week.ts) rather than only from `$lib`-free files.
	plugins: [sveltekit()],
	test: {
		// Only pick up unit tests under src/ — the Playwright suites live in
		// e2e/*.spec.ts and run via `npm run test:e2e`.
		include: ['src/**/*.test.ts'],
		environment: 'node'
	}
});
