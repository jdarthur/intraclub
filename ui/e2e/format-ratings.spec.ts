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

	// The Ratings tab hosts the transfer list.
	await page.getByRole('tab', { name: 'Ratings' }).click();

	// Fresh format: no ratings assigned yet; both ratings are available in the
	// source list. The target is empty, so saving is blocked.
	await expect(page.getByText('0 ratings')).toBeVisible();
	await expect(page.getByText('A format must keep at least one rating.')).toBeVisible();
	await expect(page.getByRole('button', { name: 'Save ratings' })).toBeDisabled();

	// Assign the first rating: select it in the source, move it, save.
	await page.getByRole('button', { name: ratingName1, exact: true }).click();
	await page.getByRole('button', { name: 'Move selected to pool' }).click();
	await expect(page.getByText('1 rating')).toBeVisible();
	await page.getByRole('button', { name: 'Save ratings' }).click();
	await expect(page.getByText('1 rating')).toBeVisible();

	// Assign the second rating.
	await page.getByRole('button', { name: ratingName2, exact: true }).click();
	await page.getByRole('button', { name: 'Move selected to pool' }).click();
	await page.getByRole('button', { name: 'Save ratings' }).click();
	await expect(page.getByText('2 ratings')).toBeVisible();

	// Remove the first rating: select it in the target, move it back out, save.
	const firstItem = page.getByRole('button', { name: ratingName1, exact: true });
	await firstItem.getByRole('checkbox').click();
	await expect(firstItem.getByRole('checkbox')).toHaveAttribute('aria-checked', 'true');
	await page.getByRole('button', { name: 'Move selected out of pool' }).click();
	await expect(page.getByText('1 rating')).toBeVisible();
	await page.getByRole('button', { name: 'Save ratings' }).click();
	await expect(page.getByText('1 rating')).toBeVisible();

	// The removed rating is available again and the remaining one is assigned.
	await expect(page.getByRole('button', { name: ratingName1, exact: true })).toBeVisible();
	await expect(page.getByRole('button', { name: ratingName2, exact: true })).toBeVisible();

	// A format must keep at least one rating: move the last one out and the
	// save button becomes disabled again.
	const secondItem = page.getByRole('button', { name: ratingName2, exact: true });
	await secondItem.getByRole('checkbox').click();
	await page.getByRole('button', { name: 'Move selected out of pool' }).click();
	await expect(page.getByText('0 ratings')).toBeVisible();
	await expect(page.getByText('A format must keep at least one rating.')).toBeVisible();
	await expect(page.getByRole('button', { name: 'Save ratings' })).toBeDisabled();

	// Clean up. Deleting a format now cascades to its format_rating / format_line
	// join rows, so the ratings are no longer "in-use" and can be deleted to
	// keep the shared dev db clean. Both deletes below confirm via an in-app
	// popover (no window.confirm). The delete button lives on the Details tab.
	await page.getByRole('tab', { name: 'Details' }).click();
	await page.getByRole('button', { name: 'Delete format' }).click();
	await page.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(page).toHaveURL('/formats');
	await expect(page.getByRole('link', { name: formatName })).toHaveCount(0);

	for (const ratingName of [ratingName1, ratingName2]) {
		await page.goto('/ratings');
		await page.getByRole('link', { name: ratingName }).click();
		await expect(page).toHaveURL(/\/ratings\/[0-9a-f]+$/);
		await page.getByRole('button', { name: 'Delete rating' }).click();
		await page.getByRole('button', { name: 'Delete', exact: true }).click();
		await expect(page).toHaveURL('/ratings');
		await expect(page.getByRole('link', { name: ratingName })).toHaveCount(0);
	}
});
