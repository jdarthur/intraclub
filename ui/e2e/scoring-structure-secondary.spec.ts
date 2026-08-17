import { expect, test, type Page } from '@playwright/test';
import { waitForHydration } from './helpers';

// This spec drives a long UI flow (creating four structures, transfer-list
// edits, saves, reloads, then deleting each structure) that routinely exceeds
// Playwright's 30s default test timeout when the fullyParallel suite runs on a
// loaded machine, so give it a longer budget.
test.describe.configure({ timeout: 60_000 });

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

test('scoring structure secondaries: assign, save, reorder, remove', async ({ page }) => {
	await login(page);

	const unique = Date.now();
	const gameA = `Test Game A ${unique}`;
	const gameB = `Test Game B ${unique}`;
	const gameC = `Test Game C ${unique}`;
	const match = `Test Match ${unique}`;

	// Three game-based structures to use as secondaries (a set-based primary
	// uses game-based secondaries), and a set-based primary (best of 3 sets ->
	// plays at most 3 sets -> needs 3 secondaries).
	await createScoringStructure(page, gameA, 'Game', '11', '1', '0');
	await createScoringStructure(page, gameB, 'Game', '11', '1', '0');
	await createScoringStructure(page, gameC, 'Game', '11', '1', '0');
	await createScoringStructure(page, match, 'Set', '2', '1', '0');

	// No secondaries assigned yet, with the required-count hint.
	await expect(page.getByText('No secondary scoring structures assigned yet.')).toBeVisible();
	await expect(
		page.getByText(
			'This win condition can play at most 3 sets, so a composite structure needs exactly 3 secondary scoring structures (0 currently assigned).'
		)
	).toBeVisible();

	// The target list shows the assigned secondaries in tie-breaker order; the
	// source list shows the structures still available to assign.
	const targetItems = page.locator('.secondary-target .secondary-name');

	// Build the assignment by moving each secondary into the target in turn;
	// the target order follows the order they are moved in.
	for (const name of [gameA, gameB, gameC]) {
		await page.getByRole('button', { name, exact: true }).click();
		await page.getByRole('button', { name: 'Move selected to pool' }).click();
	}
	await expect(targetItems.nth(0)).toHaveText(gameA);
	await expect(targetItems.nth(1)).toHaveText(gameB);
	await expect(targetItems.nth(2)).toHaveText(gameC);
	await expect(
		page.getByText('All 3 required secondary scoring structures assigned.')
	).toBeVisible();

	// Commit the assignment once complete.
	await page.getByRole('button', { name: 'Save secondaries' }).click();
	await expect(
		page.getByText('All 3 required secondary scoring structures assigned.')
	).toBeVisible();

	// Reorder: move the first tie-breaker (gameA) to the end by removing it
	// and re-adding it; the saved order follows the target list order
	// (SecondaryIndex).
	await page
		.locator('.secondary-target')
		.getByRole('button', { name: gameA, exact: true })
		.getByRole('checkbox')
		.click();
	await page.getByRole('button', { name: 'Move selected out of pool' }).click();
	await page.getByRole('button', { name: gameA, exact: true }).click();
	await page.getByRole('button', { name: 'Move selected to pool' }).click();
	await expect(targetItems.nth(0)).toHaveText(gameB);
	await expect(targetItems.nth(1)).toHaveText(gameC);
	await expect(targetItems.nth(2)).toHaveText(gameA);
	await page.getByRole('button', { name: 'Save secondaries' }).click();

	// The reordered assignment survives a reload (SecondaryIndex order).
	await page.reload();
	await waitForHydration(page);
	await expect(targetItems.nth(0)).toHaveText(gameB);
	await expect(targetItems.nth(1)).toHaveText(gameC);
	await expect(targetItems.nth(2)).toHaveText(gameA);

	// Saving the primary succeeds now that the required number of secondaries
	// is assigned.
	await page.getByLabel('Name').fill(`${match} Updated`);
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByRole('heading', { name: `${match} Updated` })).toBeVisible();

	// Remove one secondary; the assignment is now under the exact count, so
	// the save is disabled until the exact required count is restored.
	await page
		.locator('.secondary-target')
		.getByRole('button', { name: gameC, exact: true })
		.getByRole('checkbox')
		.click();
	await page.getByRole('button', { name: 'Move selected out of pool' }).click();
	await expect(
		page.getByText(
			'This win condition can play at most 3 sets, so a composite structure needs exactly 3 secondary scoring structures (2 currently assigned).'
		)
	).toBeVisible();
	await expect(page.getByRole('button', { name: 'Save secondaries' })).toBeDisabled();

	// Re-add a secondary and commit; the draft is complete again.
	await page.getByRole('button', { name: gameC, exact: true }).click();
	await page.getByRole('button', { name: 'Move selected to pool' }).click();
	await expect(
		page.getByText('All 3 required secondary scoring structures assigned.')
	).toBeVisible();
	await page.getByRole('button', { name: 'Save secondaries' }).click();

	await page.getByLabel('Name').fill(`${match} Final`);
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByRole('heading', { name: `${match} Final` })).toBeVisible();

	// Clean up. Deleting the primary cascades to its secondary join rows, so
	// the three game structures are no longer referenced and can be deleted to
	// keep the shared dev db clean. Both deletes confirm via an in-app popover.
	await page.getByRole('button', { name: 'Delete scoring structure' }).click();
	await page.getByRole('button', { name: 'Delete', exact: true }).click();
	await expect(page).toHaveURL('/scoring-structures');
	await expect(page.getByRole('link', { name: `${match} Final` })).toHaveCount(0);

	for (const name of [gameA, gameB, gameC]) {
		await page.goto('/scoring-structures');
		await page.getByRole('link', { name: name }).click();
		await expect(page).toHaveURL(/\/scoring-structures\/[0-9a-f]+$/);
		await page.getByRole('button', { name: 'Delete scoring structure' }).click();
		await page.getByRole('button', { name: 'Delete', exact: true }).click();
		await expect(page).toHaveURL('/scoring-structures');
		await expect(page.getByRole('link', { name: name })).toHaveCount(0);
	}
});
