import type { Page } from '@playwright/test';

/**
 * Wait for SvelteKit hydration to finish after navigating to a page.
 *
 * Svelte 5 attaches a form's `onsubmit` handler through event delegation only
 * during hydration. If a test fills and clicks "Create" before hydration is
 * live, the `<form>` submits natively as a GET (`/new?`) instead of running the
 * enhanced submit handler's `goto(...)`, so the detail page never loads; a
 * `bind:value` input filled pre-hydration can also be reset. Under
 * `fullyParallel` the race is much more likely because Vite compiles route
 * chunks on demand, so each worker can reach the `/new` page before its JS is
 * ready — the flake tracked in issue #103.
 *
 * `networkidle` settles once the route's JS chunks have been fetched and the
 * app has hydrated — the same barrier the login flow already relies on.
 */
export async function waitForHydration(page: Page): Promise<void> {
	await page.waitForLoadState('networkidle');
}
