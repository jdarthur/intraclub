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

// SeedDevData creates a format named "Men's Intraclub" with the possible
// ratings "Men's 1", "Men's 2", "Men's 3" (ordered highest-skill first).
const SEED_FORMAT = "Men's Intraclub";

const API = 'http://127.0.0.1:8080/api';

test('pre-draft grading page assigns grades and ranks players', async ({ page }) => {
	await login(page);
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	const unique = Date.now();

	// Create a roster user to grade via the CSV import endpoint (the
	// self-register route needs a configured mailer, which isn't available in
	// the dev e2e environment).
	const people = [{ first: `Casey${unique}`, last: 'Roster' }];
	const csv = [
		'First Name, Last Name, Email',
		...people.map((p) => `${p.first}, ${p.last}, ${p.first.toLowerCase()}@example.com`)
	].join('\n');
	const importRes = await page.request.post(`${API}/import_users_from_csv`, {
		form: { file: csv }
	});
	expect(importRes.ok()).toBeTruthy();
	const importBody = await importRes.json();
	const userIds = [...importBody.Created, ...importBody.AlreadyExisting].map(
		(u: { id: string }) => u.id
	);
	expect(userIds).toHaveLength(people.length);

	const formatRes = await page.request.get(`${API}/format`);
	const formats = (await formatRes.json()).resource as { id: string; name: string }[];
	const format = formats.find((f) => f.name === SEED_FORMAT);
	expect(format).toBeTruthy();

	const draftName = `Grades Draft ${unique}`;
	const draftRes = await page.request.post(`${API}/draft`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: draftName, format: format!.id }
	});
	expect(draftRes.ok()).toBeTruthy();
	const draftId = (await draftRes.json()).resource.id as string;

	// Add the two roster players to the draft's available-to-draft list.
	const assignRes = await page.request.post(`${API}/draft/${draftId}/assign_draftable_players`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { players: userIds }
	});
	expect(assignRes.ok()).toBeTruthy();

	const rosterOne = `${people[0].first} ${people[0].last}`;

	// --- Grade a player ---
	await page.goto(`/drafts/${draftId}/grades`);
	await expect(page.getByRole('heading', { name: draftName })).toBeVisible();
	await waitForHydration(page);

	// Men's 1 is the highest rating; Strong (modifier 2) -> numeric 9.
	await page.locator('#modifier-0').selectOption({ label: 'Strong' });
	await page.locator('#rating-0').selectOption({ label: "Men's 1" });
	await page.getByRole('button', { name: 'Save grades' }).click();

	// The grade persists: the entry shows "saved" and the ranking shows 9.00.
	await expect(page.getByText('saved')).toBeVisible();
	await expect(page.getByRole('table').getByText('9.00')).toBeVisible();

	await page.reload();
	await waitForHydration(page);
	await expect(page.locator('#modifier-0')).toHaveValue('2');
	await expect(page.locator('#rating-0')).toHaveValue(
		(await page.request.get(`${API}/rating`).then((r) => r.json()))
			.resource.find((x: { name: string }) => x.name === "Men's 1").id
	);
	await expect(page.getByText('saved')).toBeVisible();
	await expect(page.getByRole('table').getByText('9.00')).toBeVisible();

	// Clean up the draft so the table doesn't accumulate rows across runs.
	const cleanup = await page.request.delete(`${API}/draft/${draftId}`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(cleanup.ok()).toBeTruthy();
});
