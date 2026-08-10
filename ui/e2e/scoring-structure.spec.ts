import { expect, test, type Page } from '@playwright/test';

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

test('scoring structure CRUD: create, view, update, delete', async ({ page }) => {
	await login(page);

	const unique = Date.now();

	// Create
	await page.goto('/scoring-structures');
	await page.getByRole('link', { name: 'New scoring structure' }).click();
	await expect(page).toHaveURL(/\/scoring-structures\/new$/);
	await page.getByLabel('Name').fill(`Test Scoring Structure ${unique}`);
	await page.getByLabel('Score counting type').selectOption({ label: 'Game' });
	await page.getByLabel('Win threshold', { exact: true }).fill('5');
	await page.getByLabel('Must win by', { exact: true }).fill('2');
	await page.getByLabel('Instant win threshold', { exact: true }).fill('0');
	await page.getByRole('button', { name: 'Create scoring structure' }).click();

	// View: lands on the detail page with the submitted values reflected
	await expect(page).toHaveURL(/\/scoring-structures\/[0-9a-f]+$/);
	await expect(
		page.getByRole('heading', { name: `Test Scoring Structure ${unique}` })
	).toBeVisible();
	await expect(page.getByLabel('Score counting type')).toHaveValue('1'); // Game
	await expect(page.getByLabel('Win threshold', { exact: true })).toHaveValue('5');
	await expect(page.getByLabel('Must win by', { exact: true })).toHaveValue('2');

	// Update
	await page.getByLabel('Name').fill(`Test Scoring Structure ${unique} Updated`);
	await page.getByLabel('Score counting type').selectOption({ label: 'Set' });
	await page.getByLabel('Win threshold', { exact: true }).fill('3');
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(
		page.getByRole('heading', { name: `Test Scoring Structure ${unique} Updated` })
	).toBeVisible();
	await expect(page.getByLabel('Score counting type')).toHaveValue('2'); // Set
	await expect(page.getByLabel('Win threshold', { exact: true })).toHaveValue('3');

	// The list reflects the update
	await page.goto('/scoring-structures');
	const row = page.getByRole('link', { name: `Test Scoring Structure ${unique} Updated` });
	await expect(row).toBeVisible();

	// Delete (accept the confirm dialog)
	await row.click();
	await expect(
		page.getByRole('heading', { name: `Test Scoring Structure ${unique} Updated` })
	).toBeVisible();
	page.on('dialog', (d) => d.accept());
	await page.getByRole('button', { name: 'Delete scoring structure' }).click();
	await expect(page).toHaveURL('/scoring-structures');
	await expect(
		page.getByRole('link', { name: `Test Scoring Structure ${unique} Updated` })
	).toHaveCount(0);
});
