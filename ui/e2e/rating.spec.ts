import { expect, test, type Page } from '@playwright/test';
import { DatabaseSync } from 'node:sqlite';
import crypto from 'node:crypto';
import { fileURLToPath } from 'node:url';
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

// The Playwright test process runs with cwd = ui/, while the Go backend
// (spawned from the repo root) keeps its SQLite db at <repo>/intraclub.db.
const dbPath = fileURLToPath(new URL('../../intraclub.db', import.meta.url));

// Delete now confirms through an in-app shadcn Popover (no native window.confirm):
// click the trigger, then the "Delete" button inside the popover.
async function confirmDelete(page: Page) {
	await page.getByRole('button', { name: 'Delete rating' }).click();
	await page.getByRole('button', { name: 'Delete', exact: true }).click();
}

test('rating CRUD: create, view, update, delete', async ({ page }) => {
	await login(page);

	const unique = Date.now();

	// Create
	await page.goto('/ratings');
	await page.getByRole('link', { name: 'New rating' }).click();
	await expect(page).toHaveURL(/\/ratings\/new$/);
	await waitForHydration(page);
	await page.getByLabel('Name').fill(`Test Rating ${unique}`);
	await page.getByLabel('Description').fill(`Description ${unique}`);
	await page.getByRole('button', { name: 'Create rating' }).click();

	// View: lands on the detail page
	await expect(page).toHaveURL(/\/ratings\/[0-9a-f]+$/);
	await expect(page.getByRole('heading', { name: `Test Rating ${unique}` })).toBeVisible();
	await expect(page.getByLabel('Description')).toHaveValue(`Description ${unique}`);

	// Update
	await page.getByLabel('Name').fill(`Test Rating ${unique} Updated`);
	await page.getByLabel('Description').fill(`Description ${unique} Updated`);
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByRole('heading', { name: `Test Rating ${unique} Updated` })).toBeVisible();

	// The list reflects the update
	await page.goto('/ratings');
	const row = page.getByRole('link', { name: `Test Rating ${unique} Updated` });
	await expect(row).toBeVisible();

	// Delete (confirm via the popover)
	await row.click();
	await expect(page.getByRole('heading', { name: `Test Rating ${unique} Updated` })).toBeVisible();
	await confirmDelete(page);
	await expect(page).toHaveURL('/ratings');
	await expect(page.getByRole('link', { name: `Test Rating ${unique} Updated` })).toHaveCount(0);
});

test('rating delete is blocked when the rating is assigned to a format', async ({ page }) => {
	await login(page);

	// Create a rating to attempt deletion of.
	const unique = Date.now();
	await page.goto('/ratings/new');
	await waitForHydration(page);
	await page.getByLabel('Name').fill(`In Use Rating ${unique}`);
	await page.getByLabel('Description').fill(`Description ${unique}`);
	await page.getByRole('button', { name: 'Create rating' }).click();
	await expect(page).toHaveURL(/\/ratings\/([0-9a-f]+)$/);
	const ratingId = page.url().match(/\/ratings\/([0-9a-f]+)$/)![1];

	// Insert a FormatRating referencing this rating directly into the SQLite db
	// (there is no FormatRating CRUD API yet). The `format_id` below points at a
	// non-existent format: that's fine because Rating.PreDelete only counts
	// format_rating rows (it does not resolve each format), so this reliably
	// blocks the delete.
	const formatRatingId = crypto.randomBytes(8).toString('hex');
	const formatId = crypto.randomBytes(8).toString('hex');
	const db = new DatabaseSync(dbPath);
	// The Go backend holds this SQLite file open; wait out any transient lock
	// instead of failing immediately with "database is locked".
	db.exec('PRAGMA busy_timeout = 10000');
	try {
		db.prepare(
			`INSERT INTO format_rating (id, format_id, rating_id, rating_index) VALUES (?, ?, ?, ?)`
		).run(formatRatingId, formatId, ratingId, 0);

		// Attempt to delete -> PreDelete blocks it and the page surfaces the error.
		await confirmDelete(page);
		await expect(page.getByText(/in-use by/)).toBeVisible();
	} finally {
		// Clean up the format_rating row so it can't affect other tests.
		db.prepare('DELETE FROM format_rating WHERE id = ?').run(formatRatingId);
		db.close();
	}

	// With the format_rating gone, the rating can now be deleted.
	await confirmDelete(page);
	await expect(page).toHaveURL('/ratings');
	await expect(page.getByRole('link', { name: `In Use Rating ${unique}` })).toHaveCount(0);
});
