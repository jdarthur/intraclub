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

// createScoringStructure fills out the /scoring-structures/new form and
// creates a structure with the given values, landing on its detail page.
async function createScoringStructure(
	page: Page,
	name: string,
	countingType: string,
	winThreshold: string,
	mustWinBy: string,
	instantWinThreshold: string
) {
	await page.goto('/scoring-structures/new');
	await waitForHydration(page);
	await page.getByLabel('Name').fill(name);
	await page.getByLabel('Score counting type').selectOption({ label: countingType });
	await page.getByLabel('Win threshold', { exact: true }).fill(winThreshold);
	await page.getByLabel('Must win by', { exact: true }).fill(mustWinBy);
	await page.getByLabel('Instant win threshold', { exact: true }).fill(instantWinThreshold);
	await page.getByRole('button', { name: 'Create scoring structure' }).click();
	await expect(page).toHaveURL(/\/scoring-structures\/[0-9a-f]+$/);
}

test('scoring structure secondaries: add, reorder, remove, save', async ({ page }) => {
	await login(page);

	const unique = Date.now();
	const gameA = `Test Game A ${unique}`;
	const gameB = `Test Game B ${unique}`;
	const match = `Test Match ${unique}`;

	// Two game-based structures to use as secondaries (a set-based primary
	// uses game-based secondaries), and a set-based primary (best of 3 sets ->
	// plays at most 3 sets -> needs 3 secondaries).
	await createScoringStructure(page, gameA, 'Game', '11', '1', '0');
	await createScoringStructure(page, gameB, 'Game', '11', '1', '0');
	await createScoringStructure(page, match, 'Set', '2', '1', '0');

	// No secondaries assigned yet, with the required-count hint.
	await expect(page.getByText('No secondary scoring structures assigned yet.')).toBeVisible();
	await expect(
		page.getByText(
			'This win condition can play at most 3 sets, so a composite structure needs exactly 3 secondary scoring structures (0 currently assigned).'
		)
	).toBeVisible();

	const secondaryNames = page.locator('.secondary-name');

	// Add three secondaries (the same structure may be used at multiple
	// positions). Each add replaces the full list via a PUT, so wait for the
	// save to settle (the count hint updates from the response) before the
	// next add.
	const toAssign = [gameA, gameB, gameA];
	for (let i = 0; i < toAssign.length; i++) {
		await page
			.getByLabel('Secondary structure to assign')
			.selectOption({ label: toAssign[i] });
		await page.getByRole('button', { name: 'Add secondary' }).click();
		if (i < toAssign.length - 1) {
			await expect(page.getByText(`(${i + 1} currently assigned)`)).toBeVisible();
		}
	}
	await expect(secondaryNames.nth(0)).toHaveText(gameA);
	await expect(secondaryNames.nth(1)).toHaveText(gameB);
	await expect(secondaryNames.nth(2)).toHaveText(gameA);
	await expect(
		page.getByText('All 3 required secondary scoring structures assigned.')
	).toBeVisible();

	// Reorder: move the first row (gameA) down so gameB becomes the first
	// tie-breaker. The new order is preserved.
	await page
		.locator('li')
		.filter({ hasText: gameA })
		.first()
		.getByRole('button', { name: '↓' })
		.click();
	await expect(secondaryNames.nth(0)).toHaveText(gameB);
	await expect(secondaryNames.nth(1)).toHaveText(gameA);
	await expect(secondaryNames.nth(2)).toHaveText(gameA);

	// Saving the primary succeeds now that the required number of secondaries
	// is assigned.
	await page.getByLabel('Name').fill(`${match} Updated`);
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByRole('heading', { name: `${match} Updated` })).toBeVisible();

	// Remove one secondary; the structure is now under-assigned, so saving the
	// primary fails validation.
	await page
		.locator('li')
		.filter({ hasText: gameA })
		.last()
		.getByRole('button', { name: 'Remove' })
		.click();
	await expect(
		page.getByText(
			'This win condition can play at most 3 sets, so a composite structure needs exactly 3 secondary scoring structures (2 currently assigned).'
		)
	).toBeVisible();

	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(
		page.getByText('secondary scoring structures length is 2, but we can play 3 max sets')
	).toBeVisible();

	// Re-add a secondary; saving the primary succeeds again.
	await page.getByLabel('Secondary structure to assign').selectOption({ label: gameB });
	await page.getByRole('button', { name: 'Add secondary' }).click();
	await expect(
		page.getByText('All 3 required secondary scoring structures assigned.')
	).toBeVisible();

	await page.getByLabel('Name').fill(`${match} Final`);
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByRole('heading', { name: `${match} Final` })).toBeVisible();

	// Clean up. Deleting the primary cascades to its secondary join rows, so
	// the two game structures are no longer referenced and can be deleted to
	// keep the shared dev db clean. Both deletes confirm via an in-app popover.
	await page.getByRole('button', { name: 'Delete scoring structure' }).click();
	await page.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(page).toHaveURL('/scoring-structures');
	await expect(page.getByRole('link', { name: `${match} Final` })).toHaveCount(0);

	for (const name of [gameA, gameB]) {
		await page.goto('/scoring-structures');
		await page.getByRole('link', { name: name }).click();
		await expect(page).toHaveURL(/\/scoring-structures\/[0-9a-f]+$/);
		await page.getByRole('button', { name: 'Delete scoring structure' }).click();
		await page.getByRole('button', { name: 'Delete', exact: true }).click();
		await expect(page).toHaveURL('/scoring-structures');
		await expect(page.getByRole('link', { name: name })).toHaveCount(0);
	}
});
