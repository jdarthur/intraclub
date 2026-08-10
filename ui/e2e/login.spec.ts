import { expect, test } from '@playwright/test';

// The Go backend runs with --dev-token in the Playwright harness (see
// playwright.config.ts), so POST /api/one_time_password returns the magic-link
// token in the response body instead of emailing it.
test('dev-mode login renders a magic link and completes the token→JWT exchange', async ({
	page
}) => {
	await page.goto('/login');

	// Wait for SvelteKit hydration to finish so the bound email input and the
	// form's submit handler are live (filling before hydration resets the value).
	await page.waitForLoadState('networkidle');

	// Submit the email of the user seeded by SeedDevData (model/dev_data.go).
	await page.getByLabel('Email').fill('jdarthur@gatech.edu');
	await page.getByRole('button', { name: 'Send login link' }).click();

	// Dev mode: the API returned a token, so the page shows the DEV MODE ONLY
	// annotation and a clickable magic link instead of "check your email".
	await expect(page.getByText(/DEV MODE ONLY/)).toBeVisible();

	const link = page.getByRole('main').getByRole('link', { name: 'Log in' });
	await expect(link).toHaveAttribute('href', /^\/auth\/callback\?token=/);

	// Clicking the link runs the existing token→JWT exchange and lands on "/".
	await link.click();
	await expect(page).toHaveURL('/');
	await expect(page.getByRole('heading', { name: 'Welcome to IntraClub' })).toBeVisible();

	// The exchanged JWT should be persisted.
	const jwt = await page.evaluate(() => localStorage.getItem('intraclub_jwt'));
	expect(jwt).toBeTruthy();
});
