import { expect, test, type Page } from '@playwright/test';
import { waitForHydration } from './helpers';

// The Go backend runs with --dev-token from the repo root (see
// playwright.config.ts), so POST /api/one_time_password returns the magic-link
// token in the response body. Login with the user seeded by SeedDevData.
async function login(page: Page) {
	await page.goto('/login');
	await page.waitForLoadState('networkidle');
	await page.getByLabel('Email').fill('jdarthur@gatech.edu');
	await page.getByRole('button', { name: 'Send login link' }).click();
	const link = page.getByRole('main').getByRole('link', { name: 'Log in' });
	await expect(link).toBeVisible();
	await link.click();
	await expect(page).toHaveURL('/');
}

// Delete now confirms through an in-app shadcn Popover (no native window.confirm):
// click the trigger, then the "Delete" button inside the popover.
async function confirmDelete(page: Page) {
	await page.getByRole('button', { name: 'Delete playoff structure' }).click();
	await page.getByRole('button', { name: 'Delete', exact: true }).click();
}

test('playoff structure CRUD: create, view, update, delete', async ({ page }) => {
	await login(page);

	// Create (8 teams, 0 byes -> 4 first-round matchups, a valid bracket)
	await page.goto('/playoff-structures');
	await page.getByRole('link', { name: 'New playoff structure' }).click();
	await expect(page).toHaveURL(/\/playoff-structures\/new$/);
	await waitForHydration(page);
	await page.getByLabel('Byes').fill('0');
	await page.getByLabel('Number of teams').fill('8');
	await page.getByRole('button', { name: 'Create playoff structure' }).click();

	// View: lands on the detail page with the submitted values reflected
	await expect(page).toHaveURL(/\/playoff-structures\/[0-9a-f]+$/);
	await expect(page.getByRole('heading', { name: 'Playoff Structure' })).toBeVisible();
	await expect(page.getByLabel('Byes')).toHaveValue('0');
	await expect(page.getByLabel('Number of teams')).toHaveValue('8');

	// The record's detail path uniquely identifies it (configs like "0 byes /
	// 8 teams" are not unique, so select the list row by href rather than text).
	const recordLink = page.locator(`a[href="${new URL(page.url()).pathname}"]`);

	// Update: 8 -> 4 teams keeps a valid bracket (2 first-round matchups)
	await page.getByLabel('Number of teams').fill('4');
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByLabel('Number of teams')).toHaveValue('4');

	// The list reflects the update
	await page.goto('/playoff-structures');
	await expect(recordLink).toHaveText('0 byes / 4 teams');

	// Delete (confirm via the popover)
	await recordLink.click();
	await expect(page).toHaveURL(/\/playoff-structures\/[0-9a-f]+$/);
	await confirmDelete(page);
	await expect(page).toHaveURL('/playoff-structures');
	await expect(recordLink).toHaveCount(0);
});
