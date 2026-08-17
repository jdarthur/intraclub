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

// Delete now confirms through an in-app shadcn Popover (no native window.confirm):
// click the trigger, then the "Delete" button inside the popover.
async function confirmDelete(page: Page) {
	await page.getByRole('button', { name: 'Delete facility' }).click();
	await page.getByRole('button', { name: 'Delete', exact: true }).click();
}

// The Playwright test process runs with cwd = ui/, while the Go backend
// (spawned from the repo root) keeps its SQLite db at <repo>/intraclub.db.
const dbPath = fileURLToPath(new URL('../../intraclub.db', import.meta.url));

// A 1x1 transparent PNG. The backend stores raw bytes and doesn't decode the
// image, so this is enough to exercise the upload path end to end.
const PNG_1PX =
	'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==';

test('facility CRUD: create, view, update, delete', async ({ page }) => {
	await login(page);

	const unique = Date.now();

	// Create
	await page.goto('/facilities');
	await page.getByRole('link', { name: 'New facility' }).click();
	await expect(page).toHaveURL(/\/facilities\/new$/);
	await waitForHydration(page);
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

	// Delete (confirm via the popover)
	await row.click();
	await expect(page.getByRole('heading', { name: `Test Facility ${unique} Updated` })).toBeVisible();
	await confirmDelete(page);
	await expect(page).toHaveURL('/facilities');
	await expect(page.getByRole('link', { name: `Test Facility ${unique} Updated` })).toHaveCount(0);
});

test('facility delete is blocked when the facility is assigned to a season', async ({ page }) => {
	await login(page);

	// Create a facility to attempt deletion of.
	const unique = Date.now();
	await page.goto('/facilities/new');
	await waitForHydration(page);
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
		await confirmDelete(page);
		await expect(page.getByText(/is in use/)).toBeVisible();
	} finally {
		// Clean up the season row so it can't affect other tests.
		db.prepare('DELETE FROM season WHERE id = ?').run(seasonId);
		db.close();
	}

	// With the season gone, the facility can now be deleted.
	await confirmDelete(page);
	await expect(page).toHaveURL('/facilities');
	await expect(page.getByRole('link', { name: `In Use Facility ${unique}` })).toHaveCount(0);
});

test('facility layout photo: upload from the form and pick from existing photos', async ({
	page
}) => {
	await login(page);

	const unique = Date.now();
	const name = `Photo Facility ${unique}`;

	// Create a facility, attaching a layout photo uploaded inline in the form
	// (no need to visit /photos first).
	await page.goto('/facilities/new');
	await waitForHydration(page);
	await page.getByLabel('Name').fill(name);
	await page.getByLabel('Address').fill(`${unique} St`);
	await page.getByLabel('Number of courts').fill('3');
	await page.getByRole('button', { name: 'Select layout photo' }).click();
	await page.getByLabel('Alt text').fill(`Inline Photo ${unique}`);
	await page.setInputFiles('input[type=file]', {
		name: 'layout.png',
		mimeType: 'image/png',
		buffer: Buffer.from(PNG_1PX, 'base64')
	});
	await page.getByRole('button', { name: 'Upload', exact: true }).click();

	// The picker closes and the trigger reflects the new selection.
	const trigger = page.getByRole('button', { name: 'Select layout photo' });
	await expect(trigger).toContainText(`Inline Photo ${unique}`);
	await page.getByRole('button', { name: 'Create facility' }).click();

	// The selection persists on the detail/edit page.
	await expect(page).toHaveURL(/\/facilities\/[0-9a-f]+$/);
	await expect(page.getByRole('button', { name: 'Select layout photo' })).toContainText(
		`Inline Photo ${unique}`
	);

	// Upload a second photo via the photos page, then swap the layout photo to
	// it by picking it out of the thumbnail grid.
	await page.goto('/photos/new');
	await waitForHydration(page);
	await page.getByLabel('Alt text').fill(`Grid Photo ${unique}`);
	await page.setInputFiles('input[type=file]', {
		name: 'grid.png',
		mimeType: 'image/png',
		buffer: Buffer.from(PNG_1PX, 'base64')
	});
	await page.getByRole('button', { name: 'Create photo' }).click();
	await expect(page).toHaveURL(/\/photos\/[0-9a-f]+$/);

	await page.goto('/facilities');
	await page.getByRole('link', { name }).click();
	await expect(page.getByRole('heading', { name })).toBeVisible();
	await page.getByRole('button', { name: 'Select layout photo' }).click();
	await page.getByRole('button', { name: `Grid Photo ${unique}` }).click();
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByRole('button', { name: 'Select layout photo' })).toContainText(
		`Grid Photo ${unique}`
	);
});
