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

test('draft setup page configures captains, players, and rating cutoffs', async ({ page }) => {
	await login(page);
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	const unique = Date.now();

	// Create extra users to serve as captains and draftable players via the
	// CSV import endpoint (the self-register route needs a configured mailer,
	// which isn't available in the dev e2e environment).
	const people = [
		{ first: `Alex${unique}`, last: 'Captain' },
		{ first: `Blake${unique}`, last: 'Captain' },
		{ first: `Casey${unique}`, last: 'Roster' },
		{ first: `Drew${unique}`, last: 'Roster' }
	];
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

	// Create a draft for the seed format via the API so the test can focus on
	// the setup page interactions.
	const formatRes = await page.request.get(`${API}/format`);
	const formats = (await formatRes.json()).resource as { id: string; name: string }[];
	const format = formats.find((f) => f.name === SEED_FORMAT);
	expect(format).toBeTruthy();

	const draftName = `Setup Draft ${unique}`;
	const draftRes = await page.request.post(`${API}/draft`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: draftName, format: format!.id }
	});
	expect(draftRes.ok()).toBeTruthy();
	const draftId = (await draftRes.json()).resource.id as string;

	const fullName = (first: string, last: string) => `${first} ${last}`;
	const captainOne = fullName(people[0].first, people[0].last);
	const captainTwo = fullName(people[1].first, people[1].last);
	const rosterOne = fullName(people[2].first, people[2].last);
	const rosterTwo = fullName(people[3].first, people[3].last);

	// --- Captains / teams ---
	await page.goto(`/drafts/${draftId}`);
	await expect(page.getByRole('heading', { name: draftName })).toBeVisible();
	await waitForHydration(page);

	// The active tab is visually indicated (bold) so it's clear at a glance
	// which section is shown. Captains is the default tab.
	await expect(page.getByRole('tab', { name: 'Captains & teams' })).toHaveCSS(
		'font-weight',
		'600'
	);

	await page.getByLabel('Captain', { exact: true }).selectOption({ label: captainOne });
	await page.getByRole('button', { name: 'Add captain' }).click();
	await page.getByLabel('Captain', { exact: true }).selectOption({ label: captainTwo });
	await page.getByRole('button', { name: 'Add captain' }).click();
	await page.getByRole('button', { name: 'Initialize draft' }).click();

	// Initialization creates the teams/draft order and adds captains to the
	// available players list.
	await expect(page.getByText('Ready to grade & pick')).not.toBeVisible();
	await expect(page.getByRole('table').getByText(captainOne)).toBeVisible();

	// --- Available players ---
	await page.getByRole('tab', { name: 'Available players' }).click();
	await page.getByLabel('Player', { exact: true }).selectOption({ label: rosterOne });
	await page.getByRole('button', { name: 'Add player' }).click();
	// Wait for the first add's reload to commit before adding the second, so the
	// select/button DOM is stable.
	await expect(page.getByRole('list').getByText(rosterOne)).toBeVisible();
	await page.getByLabel('Player', { exact: true }).selectOption({ label: rosterTwo });
	await page.getByRole('button', { name: 'Add player' }).click();
	await expect(page.getByRole('list').getByText(rosterTwo)).toBeVisible();

	// Two captains (auto-added) + two roster players = 4 across 2 teams.
	await expect(page.getByRole('list').getByText(rosterOne)).toBeVisible();
	await expect(page.getByText('4 available players across 2 teams (2 per team)')).toBeVisible();

	// --- Rating cutoffs ---
	await page.getByRole('tab', { name: 'Rating cutoffs' }).click();
	await page.getByLabel("Men's 1").fill('2');
	await page.getByLabel("Men's 2").fill('4');
	await page.getByRole('button', { name: 'Save cutoffs' }).click();

	// Fully configured -> ready to grade/pick, and the cutoffs persist across a
	// reload.
	await expect(page.getByText('Ready to grade & pick')).toBeVisible();
	await page.reload();
	await waitForHydration(page);
	await page.getByRole('tab', { name: 'Rating cutoffs' }).click();
	await expect(page.getByLabel("Men's 1")).toHaveValue('2');
	await expect(page.getByLabel("Men's 2")).toHaveValue('4');
	await expect(page.getByText('Ready to grade & pick')).toBeVisible();

	// Clean up the draft so the table doesn't accumulate rows across runs.
	const cleanup = await page.request.delete(`${API}/draft/${draftId}`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(cleanup.ok()).toBeTruthy();
});
