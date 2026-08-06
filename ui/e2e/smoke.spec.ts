import { expect, test } from '@playwright/test';

test('homepage renders basic content', async ({ page }) => {
	await page.goto('/');

	await expect(page.getByRole('heading', { name: 'Welcome to SvelteKit' })).toBeVisible();
	await expect(page.getByRole('link', { name: /svelte.dev\/docs\/kit/i })).toBeVisible();
});
