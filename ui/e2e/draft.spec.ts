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

// SeedDevData creates a format named "Men's Intraclub" owned by the seed user.
const SEED_FORMAT = "Men's Intraclub";

test('draft create navigates to its setup page and appears in the list', async ({ page }) => {
	await login(page);

	const unique = Date.now();
	const draftName = `Test Draft ${unique}`;

	// Create
	await page.goto('/drafts');
	await expect(page.getByRole('heading', { name: 'Drafts' })).toBeVisible();
	await page.getByRole('link', { name: 'New draft' }).click();
	await expect(page).toHaveURL(/\/drafts\/new$/);
	await waitForHydration(page);

	await page.getByLabel('Name').fill(draftName);
	await page.getByLabel('Format').selectOption({ label: SEED_FORMAT });
	await page.getByRole('button', { name: 'Create draft' }).click();

	// Lands on the setup page (/drafts/[id]).
	await expect(page).toHaveURL(/\/drafts\/[0-9a-f]+$/);
	await expect(page.getByRole('heading', { name: draftName })).toBeVisible();

	// The list reflects the new draft.
	await page.goto('/drafts');
	const row = page.getByRole('link', { name: draftName });
	await expect(row).toBeVisible();

	// Clean up the draft so the table doesn't accumulate rows across runs.
	const draftId = (await row.getAttribute('href'))?.split('/').pop();
	expect(draftId).toBeTruthy();
	const res = await page.request.delete(`http://127.0.0.1:8080/api/draft/${draftId}`, {
		headers: { 'X-INTRACLUB-TOKEN': await page.evaluate(() => localStorage.getItem('intraclub_jwt')) }
	});
	expect(res.ok()).toBeTruthy();
});

test('a created draft can be deleted by its owner from the setup page', async ({ page }) => {
	await login(page);

	const unique = Date.now();
	const draftName = `Delete Me Draft ${unique}`;

	// Create
	await page.goto('/drafts');
	await page.getByRole('link', { name: 'New draft' }).click();
	await waitForHydration(page);
	await page.getByLabel('Name').fill(draftName);
	await page.getByLabel('Format').selectOption({ label: SEED_FORMAT });
	await page.getByRole('button', { name: 'Create draft' }).click();

	// Lands on the setup page, which shows the delete control for the owner.
	await expect(page).toHaveURL(/\/drafts\/[0-9a-f]+$/);
	await expect(page.getByRole('heading', { name: draftName })).toBeVisible();
	const deleteTrigger = page.getByRole('button', { name: 'Delete draft' });
	await expect(deleteTrigger).toBeVisible();

	// Confirm deletion; the page navigates back to the list.
	await deleteTrigger.click();
	await page.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(page).toHaveURL(/\/drafts$/);
	await expect(page.getByRole('link', { name: draftName })).toHaveCount(0);
});
