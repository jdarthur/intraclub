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

// tokenFor obtains a fresh JWT for an existing user via the dev-mode magic-link
// flow, so the test can act as that captain when making selections.
async function tokenFor(page: Page, email: string): Promise<string> {
	const res = await page.request.post(`${API}/one_time_password`, { data: { email } });
	expect(res.ok()).toBeTruthy();
	const { token } = await res.json();
	const jwtRes = await page.request.post(`${API}/token`, { data: { token } });
	expect(jwtRes.ok()).toBeTruthy();
	const { jwt } = await jwtRes.json();
	return jwt;
}

test('live draft board gates picks to the on-clock captain and completes a draft', async ({
	page
}) => {
	await login(page);
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	const unique = Date.now();

	// Two captains plus four roster players = 6 players across 2 teams (3 rounds).
	const people = [
		{ first: `Alex${unique}`, last: 'Captain' },
		{ first: `Blake${unique}`, last: 'Captain' },
		{ first: `Casey${unique}`, last: 'Roster' },
		{ first: `Drew${unique}`, last: 'Roster' },
		{ first: `Erin${unique}`, last: 'Roster' },
		{ first: `Faye${unique}`, last: 'Roster' }
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
	const [alexId, blakeId, caseyId, drewId, erinId, fayeId] = userIds;

	const formatRes = await page.request.get(`${API}/format`);
	const formats = (await formatRes.json()).resource as { id: string; name: string }[];
	const format = formats.find((f) => f.name === SEED_FORMAT);
	expect(format).toBeTruthy();

	const draftName = `Board Draft ${unique}`;
	const draftRes = await page.request.post(`${API}/draft`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: draftName, format: format!.id }
	});
	expect(draftRes.ok()).toBeTruthy();
	const draftId = (await draftRes.json()).resource.id as string;

	const alexName = `${people[0].first} ${people[0].last}`;
	const blakeName = `${people[1].first} ${people[1].last}`;
	const rosterOne = `${people[2].first} ${people[2].last}`;

	// Set up the draft via the API: 2 captains, 4 roster players, rating cutoffs.
	const initRes = await page.request.post(`${API}/draft/${draftId}/initialize`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { captains: [alexId, blakeId] }
	});
	expect(initRes.ok()).toBeTruthy();

	const assignRes = await page.request.post(`${API}/draft/${draftId}/assign_draftable_players`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { players: [caseyId, drewId, erinId, fayeId] }
	});
	expect(assignRes.ok()).toBeTruthy();

	const ratingsRes = await page.request.get(`${API}/format/${format!.id}/possible_ratings`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(ratingsRes.ok()).toBeTruthy();
	const ratings = (await ratingsRes.json()).resource as { id: string; name: string }[];
	expect(ratings.length).toBeGreaterThanOrEqual(2);
	// Pick 1 -> Men's 1; picks 2-5 -> Men's 2; pick 6 -> Men's 3.
	await page.request.post(`${API}/draft/${draftId}/assign_rating_cutoff`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { rating: ratings[0].id, cutoff: 1 }
	});
	await page.request.post(`${API}/draft/${draftId}/assign_rating_cutoff`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { rating: ratings[1].id, cutoff: 5 }
	});

	// --- View the board as the commissioner (not a captain): locked out ---
	await page.goto(`/drafts/${draftId}/draft`);
	await expect(page.getByRole('heading', { name: draftName })).toBeVisible();
	await waitForHydration(page);
	await expect(page.getByText(`${alexName} is on the clock`)).toBeVisible();
	// The commissioner cannot pick (the board is gated to the on-clock captain).
	await expect(page.getByRole('button', { name: 'Select' }).first()).toBeDisabled();

	// --- Captain 1 (Alex) is on the clock and can pick through the UI ---
	const alexToken = await tokenFor(page, `${people[0].first.toLowerCase()}@example.com`);
	await page.evaluate((t) => localStorage.setItem('intraclub_jwt', t), alexToken);
	await page.goto(`/drafts/${draftId}/draft`);
	await waitForHydration(page);
	await expect(page.getByText("You're on the clock")).toBeVisible();

	const rosterRow = page.locator('tr', { hasText: rosterOne });
	await rosterRow.getByRole('button', { name: 'Select' }).click();
	// The pick lands on the board and the clock advances to Blake.
	await expect(page.getByText(`${blakeName} is on the clock`)).toBeVisible();
	await expect(page.locator('table.draft-board').getByText(rosterOne)).toBeVisible();
	// Casey is now taken, so their pick button is disabled.
	await expect(rosterRow.getByRole('button', { name: 'Select' })).toBeDisabled();

	// --- Complete the draft via the API, verifying illegal picks are rejected ---
	const blakeToken = await tokenFor(page, `${people[1].first.toLowerCase()}@example.com`);

	// Blake is on the clock for pick 2; Alex trying to pick again must fail.
	const illegalRes = await page.request.post(`${API}/draft/${draftId}/select`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': alexToken },
		data: { player_id: drewId }
	});
	expect(illegalRes.ok()).toBeFalsy();

	// Snake order with 2 teams over 3 rounds: pick 1 Alex, then Blake, Blake,
	// Alex, Alex, Blake. Casey was taken in pick 1 via the UI.
	const pickOrder = [
		{ captain: blakeToken, player: drewId },
		{ captain: blakeToken, player: erinId },
		{ captain: alexToken, player: fayeId },
		{ captain: alexToken, player: alexId }, // Alex picks themselves
		{ captain: blakeToken, player: blakeId } // Blake picks themselves
	];
	for (const { captain, player } of pickOrder) {
		const res = await page.request.post(`${API}/draft/${draftId}/select`, {
			headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': captain },
			data: { player_id: player }
		});
		expect(res.ok(), `pick ${player} with expected captain failed: ${await res.text()}`).toBeTruthy();
	}

	// All 6 players have been drafted -> the board reflects completion.
	await page.goto(`/drafts/${draftId}/draft`);
	await waitForHydration(page);
	await expect(page.getByText('Completed')).toBeVisible();
	await expect(page.getByText(`All 6 players drafted.`)).toBeVisible();
	// No selectable players remain.
	await expect(page.getByRole('button', { name: 'Select' }).first()).toBeDisabled();

	// Clean up the draft so the table doesn't accumulate rows across runs.
	const cleanup = await page.request.delete(`${API}/draft/${draftId}`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(cleanup.ok()).toBeTruthy();
});
