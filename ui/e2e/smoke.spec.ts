import { expect, test } from '@playwright/test';

test('homepage renders a landing page with nav and login form', async ({ page }) => {
	await page.goto('/');

	// Nav-bar is always shown.
	await expect(page.getByRole('navigation')).toBeVisible();
	await expect(page.getByRole('button', { name: 'Settings' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Drafts' })).toBeVisible();

	// Not logged in, so the login form is displayed on the root route.
	await expect(page.getByRole('heading', { name: 'IntraClub' })).toBeVisible();
	await expect(page.getByLabel('Email')).toBeVisible();
});
