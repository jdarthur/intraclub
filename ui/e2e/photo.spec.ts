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

// Delete now confirms through an in-app shadcn Popover (no native window.confirm):
// click the trigger, then the "Delete" button inside the popover.
async function confirmDelete(page: Page) {
	await page.getByRole('button', { name: 'Delete photo' }).click();
	await page.getByRole('button', { name: 'Delete', exact: true }).click();
}

// A 1x1 transparent PNG. The backend stores raw bytes and doesn't decode the
// image, so this is enough to exercise the upload path end to end.
const PNG_1PX =
	'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==';

test('photo CRUD: upload, view, update, delete', async ({ page }) => {
	await login(page);

	const unique = Date.now();
	const altText = `Test Photo ${unique}`;

	// Create (upload)
	await page.goto('/photos');
	await page.getByRole('link', { name: 'New photo' }).click();
	await expect(page).toHaveURL(/\/photos\/new$/);
	await waitForHydration(page);
	await page.getByLabel('Alt text').fill(altText);
	await page.setInputFiles('input[type=file]', {
		name: 'photo.png',
		mimeType: 'image/png',
		buffer: Buffer.from(PNG_1PX, 'base64')
	});
	// File type is auto-derived from the extension.
	await expect(page.getByLabel('File type')).toHaveValue('0'); // png
	await page.getByRole('button', { name: 'Create photo' }).click();

	// View: lands on the detail page with the image and alt text reflected
	await expect(page).toHaveURL(/\/photos\/[0-9a-f]+$/);
	await expect(page.getByRole('heading', { name: 'Photo' })).toBeVisible();
	await expect(page.locator('img.detail')).toHaveAttribute('src', /^data:image\/png;base64,/);
	await expect(page.getByText(altText)).toBeVisible();

	// Update: change alt text and swap the image
	await page.getByLabel('Alt text').fill(`${altText} Updated`);
	await page.setInputFiles('input[type=file]', {
		name: 'photo.jpg',
		mimeType: 'image/jpeg',
		buffer: Buffer.from(PNG_1PX, 'base64')
	});
	await expect(page.getByLabel('File type')).toHaveValue('1'); // jpg
	await page.getByRole('button', { name: 'Save changes' }).click();
	await expect(page.getByText(`${altText} Updated`)).toBeVisible();
	await expect(page.locator('img.detail')).toHaveAttribute('src', /^data:image\/jpg;base64,/);

	// The list reflects the update (thumbnail + alt text)
	await page.goto('/photos');
	const row = page
		.getByRole('link', { name: `${altText} Updated` })
		.filter({ hasText: `${altText} Updated` });
	await expect(row).toBeVisible();

	// Delete (confirm via the popover)
	await row.click();
	await expect(page.getByRole('heading', { name: 'Photo' })).toBeVisible();
	await confirmDelete(page);
	await expect(page).toHaveURL('/photos');
	await expect(page.getByRole('link', { name: `${altText} Updated` })).toHaveCount(0);
});
