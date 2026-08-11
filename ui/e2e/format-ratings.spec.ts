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

test('assign ratings to a format: add, list, remove', async ({ page }) => {
	await login(page);

	const unique = Date.now();
	const ratingName1 = `Test Rating A ${unique}`;
	const ratingName2 = `Test Rating B ${unique}`;
	const formatName = `Test Format ${unique}`;

	// Create two ratings to assign.
	for (const ratingName of [ratingName1, ratingName2]) {
		await page.goto('/ratings/new');
		await waitForHydration(page);
		await page.getByLabel('Name').fill(ratingName);
		await page.getByLabel('Description').fill(`Description ${unique}`);
		await page.getByRole('button', { name: 'Create rating' }).click();
		await expect(page).toHaveURL(/\/ratings\/[0-9a-f]+$/);
	}

	// Create a fresh format with no ratings yet.
	await page.goto('/formats/new');
	await waitForHydration(page);
	await page.getByLabel('Name').fill(formatName);
	await page.getByRole('button', { name: 'Create format' }).click();
	await expect(page).toHaveURL(/\/formats\/[0-9a-f]+$/);

	// No ratings assigned yet.
	await expect(page.getByText('No ratings assigned yet.')).toBeVisible();

	const ratingNames = page.locator('.rating-name');

	// Assign the first rating.
	await page.getByLabel('Rating to assign').selectOption({ label: ratingName1 });
	await page.getByRole('button', { name: 'Add rating' }).click();
	await expect(ratingNames.getByText(ratingName1)).toBeVisible();

	// Assign the second rating.
	await page.getByLabel('Rating to assign').selectOption({ label: ratingName2 });
	await page.getByRole('button', { name: 'Add rating' }).click();
	await expect(ratingNames.getByText(ratingName1)).toBeVisible();
	await expect(ratingNames.getByText(ratingName2)).toBeVisible();

	// Remove the first rating; it disappears and becomes assignable again.
	const firstItem = page.locator('li', { hasText: ratingName1 });
	await firstItem.getByRole('button', { name: 'Remove' }).click();
	await expect(ratingNames.getByText(ratingName1)).toHaveCount(0);
	await expect(ratingNames.getByText(ratingName2)).toBeVisible();
	await expect(page.getByLabel('Rating to assign')).toContainText(ratingName1);

	// A format must keep at least one rating, so with only rating2 remaining the
	// Remove button is disabled.
	const secondItem = page.locator('li', { hasText: ratingName2 });
	await expect(secondItem.getByRole('button', { name: 'Remove' })).toBeDisabled();
	await expect(page.getByText('A format must keep at least one rating.')).toBeVisible();

	// Clean up. Deleting a format now cascades to its format_rating / format_line
	// join rows, so the ratings are no longer "in-use" and can be deleted to
	// keep the shared dev db clean.
	// The format delete confirms via an in-app popover (no window.confirm); the
	// dialog handler below is still needed for the ratings deletes further down.
	page.on('dialog', (d) => d.accept());
	await page.getByRole('button', { name: 'Delete format' }).click();
	await page.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(page).toHaveURL('/formats');
	await expect(page.getByRole('link', { name: formatName })).toHaveCount(0);

	for (const ratingName of [ratingName1, ratingName2]) {
		await page.goto('/ratings');
		await page.getByRole('link', { name: ratingName }).click();
		await expect(page).toHaveURL(/\/ratings\/[0-9a-f]+$/);
		await page.getByRole('button', { name: 'Delete rating' }).click();
		await expect(page).toHaveURL('/ratings');
		await expect(page.getByRole('link', { name: ratingName })).toHaveCount(0);
	}
});
