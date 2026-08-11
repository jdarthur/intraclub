import { expect, test } from '@playwright/test';

// The client only decodes the JWT's `exp` claim to drive the countdown, so
// these tests can drive the expiry UX with self-minted tokens whose header and
// signature are placeholders — no need to wait for a real 2h expiry window.
function makeToken(expEpochSec: number): string {
	const payload = Buffer.from(JSON.stringify({ exp: expEpochSec })).toString('base64url');
	return `header.${payload}.signature`;
}

test('expired JWT proactively redirects to login with a re-login message', async ({ page }) => {
	// Token expiring ~7s from now so the background countdown has time to fire.
	const exp = Math.floor(Date.now() / 1000) + 7;

	await page.goto('/');
	await page.evaluate((t) => localStorage.setItem('intraclub_jwt', t), makeToken(exp));
	await page.reload();

	// The token is present, so the landing page treats the user as logged in.
	await expect(page.getByRole('heading', { name: 'Welcome to IntraClub' })).toBeVisible();

	// Once `exp` elapses the countdown clears the session and shows the message.
	await expect(page.getByText('Your session expired. Please log in again.')).toBeVisible({
		timeout: 20_000
	});
	await expect(page).toHaveURL('/');

	// The stored token is cleared on expiry.
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt'));
	expect(token).toBeNull();
});

test('a 401 response reactively expires the session and redirects to login', async ({ page }) => {
	// Far-future `exp` so the countdown does NOT fire — this isolates the
	// reactive 401 path. The bogus signature makes the backend reject the token.
	const exp = Math.floor(Date.now() / 1000) + 3600;

	await page.goto('/');
	await page.evaluate((t) => localStorage.setItem('intraclub_jwt', t), makeToken(exp));

	// /facilities calls listFacilities -> authFetch on mount, which 401s and
	// should trigger the same redirect + message.
	await page.goto('/facilities');
	await expect(page.getByText('Your session expired. Please log in again.')).toBeVisible({
		timeout: 15_000
	});
	await expect(page).toHaveURL('/');

	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt'));
	expect(token).toBeNull();
});
