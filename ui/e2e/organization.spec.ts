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

	// Update the name (Details tab is active by default)
	await page.getByLabel('Name').fill(`Test Org ${unique} Updated`);
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByRole('heading', { name: `Test Org ${unique} Updated` })).toBeVisible();

	// Manage membership in the Members tab via the transfer list.
	await page.getByRole('tab', { name: 'Members' }).click();

	// Add a member: select the seeded dev user Avery Chen, move to the pool, save.
	await page.getByRole('button', { name: 'Avery Chen' }).click();
	await page.getByRole('button', { name: 'Move selected to pool' }).click();
	await page.getByRole('button', { name: 'Save members' }).click();
	await expect(page.getByText(/1 member/)).toBeVisible();

	// Reload to build a fresh transfer list from the server. This avoids racing
	// the add's async reload with the mutation below (a stale core can
	// overwrite an in-flight change and make the save no-op).
	await page.reload();
	await waitForHydration(page);
	await page.getByRole('tab', { name: 'Members' }).click();

	// Remove the member: select Avery in the target, move back to source, save.
	const averyTarget = page.getByRole('button', { name: 'Avery Chen' });
	await averyTarget.getByRole('checkbox').click();
	await expect(averyTarget.getByRole('checkbox')).toHaveAttribute('aria-checked', 'true');
	await page.getByRole('button', { name: 'Move selected out of pool' }).click();
	await page.getByRole('button', { name: 'Save members' }).click();
	await expect(page.getByText(/0 members/)).toBeVisible();

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
