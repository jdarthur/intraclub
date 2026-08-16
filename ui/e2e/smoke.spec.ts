import { expect, test } from '@playwright/test';

test('homepage renders a landing page with nav and login form', async ({ page }) => {
	await page.goto('/');

	// Nav-bar is always shown.
	await expect(page.getByRole('navigation')).toBeVisible();
	await expect(page.getByRole('button', { name: 'Settings' })).toBeVisible();
	await expect(page.getByRole('button', { name: 'Seasons' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Teams' })).toBeVisible();

	// Not logged in, so the login form is displayed on the root route.
	await expect(page.getByRole('heading', { name: 'IntraClub' })).toBeVisible();
	await expect(page.getByLabel('Email')).toBeVisible();

	// The signed-out hero explains the app lifecycle with a how-it-works strip.
	await expect(page.getByRole('heading', { name: 'How it works' })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'Draft', exact: true })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'Season', exact: true })).toBeVisible();
	await expect(page.getByRole('heading', { name: 'Score', exact: true })).toBeVisible();
});
