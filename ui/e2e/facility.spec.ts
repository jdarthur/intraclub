import { expect, test, type Page } from '@playwright/test';
import { DatabaseSync } from 'node:sqlite';
import crypto from 'node:crypto';
import { fileURLToPath } from 'node:url';

// The Go backend runs with --dev-token from the repo root (see
// playwright.config.ts), so POST /api/one_time_password returns the magic-link
// token in the response body. Login with the user seeded by SeedDevData.
async function login(page: Page) {
	await page.goto('/login');
	await page.waitForLoadState('networkidle');
	await page.getByLabel('Email').fill('jdarthur@gatech.edu');
	await page.getByRole('button', { name: 'Send login link' }).click();
	const link = page.getByRole('link', { name: 'Log in' });
	await expect(link).toBeVisible();
	await link.click();
	await expect(page).toHaveURL('/');
}

// The Playwright test process runs with cwd = ui/, while the Go backend
// (spawned from the repo root) keeps its SQLite db at <repo>/intraclub.db.
const dbPath = fileURLToPath(new URL('../../intraclub.db', import.meta.url));

test('facility CRUD: create, view, update, delete', async ({ page }) => {
	await login(page);

	const unique = Date.now();

	// Create
	await page.goto('/facilities');
	await page.getByRole('link', { name: 'New facility' }).click();
	await expect(page).toHaveURL(/\/facilities\/new$/);
	await page.getByLabel('Name').fill(`Test Facility ${unique}`);
	await page.getByLabel('Address').fill(`${unique} Main St`);
	await page.getByLabel('Number of courts').fill('4');
	await page.getByRole('button', { name: 'Create facility' }).click();

	// View: lands on the detail page
	await expect(page).toHaveURL(/\/facilities\/[0-9a-f]+$/);
	await expect(page.getByRole('heading', { name: `Test Facility ${unique}` })).toBeVisible();

	// Update
	await page.getByLabel('Name').fill(`Test Facility ${unique} Updated`);
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByRole('heading', { name: `Test Facility ${unique} Updated` })).toBeVisible();

	// The list reflects the update
	await page.goto('/facilities');
	const row = page.getByRole('link', { name: `Test Facility ${unique} Updated` });
	await expect(row).toBeVisible();

	// Delete (accept the confirm dialog)
	await row.click();
	await expect(page.getByRole('heading', { name: `Test Facility ${unique} Updated` })).toBeVisible();
	page.on('dialog', (d) => d.accept());
	await page.getByRole('button', { name: 'Delete facility' }).click();
	await expect(page).toHaveURL('/facilities');
	await expect(page.getByRole('link', { name: `Test Facility ${unique} Updated` })).toHaveCount(0);
});

test('facility delete is blocked when the facility is assigned to a season', async ({ page }) => {
	await login(page);

	// Both delete attempts below use window.confirm; accept every dialog.
	page.on('dialog', (d) => d.accept());

	// Create a facility to attempt deletion of.
	const unique = Date.now();
	await page.goto('/facilities/new');
	await page.getByLabel('Name').fill(`In Use Facility ${unique}`);
	await page.getByLabel('Address').fill(`${unique} St`);
	await page.getByLabel('Number of courts').fill('2');
	await page.getByRole('button', { name: 'Create facility' }).click();
	await expect(page).toHaveURL(/\/facilities\/([0-9a-f]+)$/);
	const facilityId = page.url().match(/\/facilities\/([0-9a-f]+)$/)![1];

	// Insert a Season referencing this facility directly into the SQLite db
	// (there is no Season CRUD API yet). The row must deserialize cleanly, so
	// every ID column needs a valid 16-hex string.
	const seasonId = crypto.randomBytes(8).toString('hex');
	const db = new DatabaseSync(dbPath);
	// The Go backend holds this SQLite file open; wait out any transient lock
	// instead of failing immediately with "database is locked".
	db.exec('PRAGMA busy_timeout = 10000');
	try {
		db.prepare(
			`INSERT INTO season
				(id, name, facility, start_time, draft_id, schedule_id, playoff_structure, owner)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
		).run(
			seasonId,
			'In Use Test',
			facilityId,
			'2025-01-01T08:30:00Z',
			seasonId,
			'0000000000000000',
			'0000000000000000',
			'000000000000000a'
		);

		// Attempt to delete -> PreDelete blocks it and the page surfaces the error.
		await page.getByRole('button', { name: 'Delete facility' }).click();
		await expect(page.getByText(/is in use/)).toBeVisible();
	} finally {
		// Clean up the season row so it can't affect other tests.
		db.prepare('DELETE FROM season WHERE id = ?').run(seasonId);
		db.close();
	}

	// With the season gone, the facility can now be deleted.
	await page.getByRole('button', { name: 'Delete facility' }).click();
	await expect(page).toHaveURL('/facilities');
	await expect(page.getByRole('link', { name: `In Use Facility ${unique}` })).toHaveCount(0);
});
