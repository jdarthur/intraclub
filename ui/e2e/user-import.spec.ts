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

test('user CSV import: creates users and reports already-existing ones', async ({ page }) => {
	await login(page);

	const unique = Date.now();
	const csv = [
		'First Name, Last Name, Email',
		`Alice, Import${unique}, alice${unique}@example.com`,
		`Bob, Import${unique}, bob${unique}@example.com`
	].join('\n');

	// First import: both users are new.
	await page.goto('/users/import');
	await waitForHydration(page);
	await page.getByLabel('CSV contents').fill(csv);
	await page.getByRole('button', { name: 'Import users' }).click();
	await expect(page.getByText(`Created (2)`)).toBeVisible();
	await expect(page.getByText(`Alice Import${unique}`)).toBeVisible();
	await expect(page.getByText(`Bob Import${unique}`)).toBeVisible();

	// Re-import the same CSV: both emails already exist.
	await page.goto('/users/import');
	await waitForHydration(page);
	await page.getByLabel('CSV contents').fill(csv);
	await page.getByRole('button', { name: 'Import users' }).click();
	await expect(page.getByText(`Created (0)`)).toBeVisible();
	await expect(page.getByText(`Already existing (2)`)).toBeVisible();
	await expect(page.getByText(`Alice Import${unique}`)).toBeVisible();

	// Invalid CSV (missing required header) surfaces the backend error.
	await page.goto('/users/import');
	await waitForHydration(page);
	await page.getByLabel('CSV contents').fill('First Name\nOnlyFirstName');
	await page.getByRole('button', { name: 'Import users' }).click();
	await expect(page.getByText(/expected 3 headers|not in the expected headers/i)).toBeVisible();
});
