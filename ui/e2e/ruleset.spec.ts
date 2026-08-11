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

test('ruleset CRUD: create, view, amend (update), delete', async ({ page }) => {
	await login(page);

	const unique = Date.now();

	// Create
	await page.goto('/rulesets');
	await page.getByRole('link', { name: 'New ruleset' }).click();
	await expect(page).toHaveURL(/\/rulesets\/new$/);
	await waitForHydration(page);
	await page.getByLabel('Name').fill(`Test Ruleset ${unique}`);
	await page.getByRole('button', { name: 'Create ruleset' }).click();

	// View: lands on the detail page as revision 0
	await expect(page).toHaveURL(/\/rulesets\/[0-9a-f]+$/);
	await expect(page.getByRole('heading', { name: `Test Ruleset ${unique}` })).toBeVisible();
	await expect(page.getByText('0', { exact: true })).toBeVisible();
	await expect(page.getByText('— (current revision)')).toBeVisible();
	const originalId = page.url().match(/\/rulesets\/([0-9a-f]+)$/)![1];

	// Amend (edit): editing produces a new superseding revision with an
	// incremented revision number and navigates to the new revision.
	await page.getByLabel('Name').fill(`Test Ruleset ${unique} Amended`);
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByRole('heading', { name: `Test Ruleset ${unique} Amended` })).toBeVisible();
	await expect(page.getByText('1', { exact: true })).toBeVisible();
	await expect(page.getByText('— (current revision)')).toBeVisible();
	const amendedId = page.url().match(/\/rulesets\/([0-9a-f]+)$/)![1];

	// The original revision is now superseded by the amended revision.
	await page.goto(`/rulesets/${originalId}`);
	await expect(page.getByRole('heading', { name: `Test Ruleset ${unique}` })).toBeVisible();
	await expect(page.getByRole('link', { name: amendedId })).toBeVisible();

	// The list reflects the amendment
	await page.goto('/rulesets');
	const row = page.getByRole('link', { name: `Test Ruleset ${unique} Amended` });
	await expect(row).toBeVisible();

	// Delete (accept the confirm dialog)
	await row.click();
	await expect(page.getByRole('heading', { name: `Test Ruleset ${unique} Amended` })).toBeVisible();
	page.on('dialog', (d) => d.accept());
	await page.getByRole('button', { name: 'Delete ruleset' }).click();
	await expect(page).toHaveURL('/rulesets');
	await expect(page.getByRole('link', { name: `Test Ruleset ${unique} Amended` })).toHaveCount(0);
});
