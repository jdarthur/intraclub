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

test('users: list shows seeded users and detail page is read-only', async ({ page }) => {
	await login(page);

	// List page shows the seeded dev users.
	await page.goto('/users');
	await waitForHydration(page);
	await expect(page.getByRole('heading', { name: 'Users' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'JD Arthur' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Avery Chen' })).toBeVisible();
	await expect(page.getByRole('cell', { name: 'jdarthur@gatech.edu' })).toBeVisible();

	// Detail page renders the user's fields without edit/delete actions.
	await page.getByRole('link', { name: 'JD Arthur' }).click();
	await expect(page).toHaveURL(/\/users\/[0-9a-f]{16}$/);
	await expect(page.getByRole('heading', { name: 'JD Arthur' })).toBeVisible();
	await expect(page.getByText('jdarthur@gatech.edu')).toBeVisible();
	await expect(page.getByRole('button', { name: /Save|Delete/ })).toHaveCount(0);
});
