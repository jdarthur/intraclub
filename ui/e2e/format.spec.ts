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
	await page.getByRole('button', { name: 'Delete format' }).click();
	await page.getByRole('button', { name: 'Delete', exact: true }).click();
}

test('format CRUD: create, view, update, delete', async ({ page }) => {
	await login(page);

	const unique = Date.now();

	// Create
	await page.goto('/formats');
	await page.getByRole('link', { name: 'New format' }).click();
	await expect(page).toHaveURL(/\/formats\/new$/);
	await waitForHydration(page);
	await page.getByLabel('Name').fill(`Test Format ${unique}`);
	await page.getByRole('button', { name: 'Create format' }).click();

	// View: lands on the detail page
	await expect(page).toHaveURL(/\/formats\/[0-9a-f]+$/);
	await expect(page.getByRole('heading', { name: `Test Format ${unique}` })).toBeVisible();

	// Update
	await page.getByLabel('Name').fill(`Test Format ${unique} Updated`);
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByRole('heading', { name: `Test Format ${unique} Updated` })).toBeVisible();

	// The list reflects the update
	await page.goto('/formats');
	const row = page.getByRole('link', { name: `Test Format ${unique} Updated` });
	await expect(row).toBeVisible();

	// Delete (confirm via the popover)
	await row.click();
	await expect(page.getByRole('heading', { name: `Test Format ${unique} Updated` })).toBeVisible();
	await confirmDelete(page);
	await expect(page).toHaveURL('/formats');
	await expect(page.getByRole('link', { name: `Test Format ${unique} Updated` })).toHaveCount(0);
});

test('format delete is blocked when the format is assigned to a draft', async ({ page }) => {
	await login(page);

	// Create a format to attempt deletion of.
	const unique = Date.now();
	await page.goto('/formats/new');
	await waitForHydration(page);
	await page.getByLabel('Name').fill(`In Use Format ${unique}`);
	await page.getByRole('button', { name: 'Create format' }).click();
	await expect(page).toHaveURL(/\/formats\/([0-9a-f]+)$/);
	const formatId = page.url().match(/\/formats\/([0-9a-f]+)$/)![1];

	// Insert a DraftFormat referencing this format directly into the SQLite db
	// (there is no Draft CRUD API yet). The `draft_id` below points at a
	// non-existent draft: that's fine because Format.PreDelete ->
	// CheckHasAssignedDrafts only counts draft_format rows (it does not resolve
	// each draft), so this reliably blocks the delete. If PreDelete ever switches
	// to resolving drafts via GetAssignedDrafts, a real draft row would be needed.
	const draftId = crypto.randomBytes(8).toString('hex');
	const draftFormatId = crypto.randomBytes(8).toString('hex');
	const db = new DatabaseSync(dbPath);
	// The Go backend holds this SQLite file open; wait out any transient lock
	// instead of failing immediately with "database is locked".
	db.exec('PRAGMA busy_timeout = 10000');
	try {
		db.prepare(`INSERT INTO draft_format (id, draft_id, format_id) VALUES (?, ?, ?)`).run(
			draftFormatId,
			draftId,
			formatId
		);

		// Attempt to delete -> PreDelete blocks it and the page surfaces the error.
		await confirmDelete(page);
		await expect(page.getByText(/assigned drafts/)).toBeVisible();
	} finally {
		// Clean up the draft_format row so it can't affect other tests.
		db.prepare('DELETE FROM draft_format WHERE id = ?').run(draftFormatId);
		db.close();
	}

	// With the draft gone, the format can now be deleted.
	await confirmDelete(page);
	await expect(page).toHaveURL('/formats');
	await expect(page.getByRole('link', { name: `In Use Format ${unique}` })).toHaveCount(0);
});
