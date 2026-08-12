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
	await page.getByRole('button', { name: 'Delete organization' }).click();
	await page.getByRole('button', { name: 'Delete', exact: true }).click();
}

test('organization CRUD: create, manage membership, update, delete', async ({ page }) => {
	await login(page);

	const unique = Date.now();

	// Create
	await page.goto('/organizations');
	await page.getByRole('link', { name: 'New organization' }).click();
	await expect(page).toHaveURL(/\/organizations\/new$/);
	await waitForHydration(page);
	await page.getByLabel('Name').fill(`Test Org ${unique}`);
	await page.getByRole('button', { name: 'Create organization' }).click();

	// View: lands on the detail page (owner can manage membership)
	await expect(page).toHaveURL(/\/organizations\/[0-9a-f]+$/);
	await expect(page.getByRole('heading', { name: `Test Org ${unique}` })).toBeVisible();

	// Add a member: pick the seeded dev user Avery Chen
	const averyOption = page.locator('select option', { hasText: 'Avery Chen' });
	const averyValue = await averyOption.getAttribute('value');
	await page.getByLabel('Member to add').selectOption(averyValue!);
	await page.getByRole('button', { name: 'Add' }).click();
	await expect(page.locator('ul').getByText('Avery Chen')).toBeVisible();

	// Remove the member
	await page.getByRole('button', { name: 'Remove' }).click();
	await expect(page.getByText(/No members yet/)).toBeVisible();

	// Update the name
	await page.getByLabel('Name').fill(`Test Org ${unique} Updated`);
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByRole('heading', { name: `Test Org ${unique} Updated` })).toBeVisible();

	// The list reflects the update
	await page.goto('/organizations');
	const row = page.getByRole('link', { name: `Test Org ${unique} Updated` });
	await expect(row).toBeVisible();

	// Delete (confirm via the popover)
	await row.click();
	await expect(page.getByRole('heading', { name: `Test Org ${unique} Updated` })).toBeVisible();
	await confirmDelete(page);
	await expect(page).toHaveURL('/organizations');
	await expect(page.getByRole('link', { name: `Test Org ${unique} Updated` })).toHaveCount(0);
});
