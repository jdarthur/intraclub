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

const SEED_FORMAT = "Men's Intraclub";
const API = 'http://127.0.0.1:8080/api';

async function tokenFor(page: Page, email: string): Promise<string> {
	const res = await page.request.post(`${API}/one_time_password`, { data: { email } });
	expect(res.ok()).toBeTruthy();
	const { token } = await res.json();
	const jwtRes = await page.request.post(`${API}/token`, { data: { token } });
	expect(jwtRes.ok()).toBeTruthy();
	const { jwt } = await jwtRes.json();
	return jwt;
}

test('commissioner builds a season schedule and participants can view it', async ({ page }) => {
	await login(page);
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	const unique = Date.now();

	// Two captains + two roster players is enough to complete a two-team draft.
	const people = [
		{ first: `CapA${unique}`, last: 'One' },
		{ first: `CapB${unique}`, last: 'Two' },
		{ first: `PlayA${unique}`, last: 'One' },
		{ first: `PlayB${unique}`, last: 'Two' }
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
	const imported = [...importBody.Created, ...importBody.AlreadyExisting] as {
		id: string;
		email: string;
	}[];
	const idFor = (email: string) => imported.find((u) => u.email === email)!.id;
	const captainAId = idFor(`${people[0].first.toLowerCase()}@example.com`);
	const captainBId = idFor(`${people[1].first.toLowerCase()}@example.com`);
	const playerAId = idFor(`${people[2].first.toLowerCase()}@example.com`);
	const playerBId = idFor(`${people[3].first.toLowerCase()}@example.com`);

	const facilityRes = await page.request.post(`${API}/facility`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: `Schedule Facility ${unique}`, address: `${unique} Main St`, courts: 4 }
	});
	expect(facilityRes.ok()).toBeTruthy();
	const facilityId = (await facilityRes.json()).resource.id as string;

	const formats = (await (await page.request.get(`${API}/format`)).json()).resource as {
		id: string;
		name: string;
	}[];
	const format = formats.find((f) => f.name === SEED_FORMAT);
	expect(format).toBeTruthy();

	const draftRes = await page.request.post(`${API}/draft`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: `Schedule Draft ${unique}`, format: format!.id }
	});
	expect(draftRes.ok()).toBeTruthy();
	const draftId = (await draftRes.json()).resource.id as string;

	const initRes = await page.request.post(`${API}/draft/${draftId}/initialize`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { captains: [captainAId, captainBId] }
	});
	expect(initRes.ok(), `initialize failed: ${await initRes.text()}`).toBeTruthy();
	await page.request.post(`${API}/draft/${draftId}/assign_draftable_players`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { players: [playerAId, playerBId] }
	});

	const ratings = (
		await (
			await page.request.get(`${API}/format/${format!.id}/possible_ratings`, {
				headers: { 'X-INTRACLUB-TOKEN': token }
			})
		).json()
	).resource as { id: string }[];
	expect(ratings.length).toBeGreaterThanOrEqual(2);
	await page.request.post(`${API}/draft/${draftId}/assign_rating_cutoff`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { rating: ratings[0].id, cutoff: 1 }
	});
	await page.request.post(`${API}/draft/${draftId}/assign_rating_cutoff`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { rating: ratings[1].id, cutoff: 5 }
	});

	const captainAToken = await tokenFor(page, `${people[0].first.toLowerCase()}@example.com`);
	const captainBToken = await tokenFor(page, `${people[1].first.toLowerCase()}@example.com`);


	// Each captain drafts themselves + one roster player -> 4 picks = complete.
	// With the default snake pattern and two captains the turn order is A, B, B, A.
	const pickPlan = [
		{ player: playerAId, token: captainAToken },
		{ player: playerBId, token: captainBToken },
		{ player: captainBId, token: captainBToken }, // captain B drafts themselves
		{ player: captainAId, token: captainAToken } // captain A drafts themselves
	];
	for (const { player, token } of pickPlan) {
		const res = await page.request.post(`${API}/draft/${draftId}/select`, {
			headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
			data: { player_id: player }
		});
		expect(res.ok(), `pick failed: ${await res.text()}`).toBeTruthy();
	}

	const seasonName = `Schedule ${unique}`;
	const seasonRes = await page.request.post(`${API}/draft/${draftId}/create_season`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: seasonName, facility: facilityId, start_time: '08:30' }
	});
	expect(seasonRes.ok()).toBeTruthy();
	const season = (await seasonRes.json()).resource as { id: string };

	// Commissioner creates a week for the season's draft.
	const weekRes = await page.request.post(`${API}/week`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { draft_id: draftId, date: '2026-01-05T08:00:00Z', note: 'Week 1' }
	});
	expect(weekRes.ok()).toBeTruthy();

	// --- build the schedule through the UI ---
	await page.goto(`/seasons/${season.id}`);
	await waitForHydration(page);
	await expect(page.getByRole('heading', { name: seasonName })).toBeVisible();

	const scheduleCard = page.getByText('Season schedule', { exact: true });
	await expect(scheduleCard).toBeVisible();

	await expect(page.getByRole('button', { name: 'Create schedule' })).toBeVisible();
	await page.getByRole('button', { name: 'Create schedule' }).click();
	await expect(page.getByRole('button', { name: 'Create schedule' })).toBeHidden();

	await expect(page.getByRole('button', { name: 'Edit' })).toBeVisible();
	await page.getByRole('button', { name: 'Edit' }).first().click();

	// Fill the matchup: Team 1 (home) vs Team 2 (away).
	await page.getByLabel('Home').selectOption({ label: 'Team 1' });
	await page.getByLabel('Away').selectOption({ label: 'Team 2' });
	await page.getByRole('button', { name: 'Save', exact: true }).click();

	// The assigned matchup is visible on the season page.
	await expect(page.getByText('Team 1 vs Team 2')).toBeVisible();

	// --- participants can view the schedule but not modify it ---
	const detailRes = await page.request.get(`${API}/schedule?season_id=${season.id}`, {
		headers: { 'X-INTRACLUB-TOKEN': captainAToken }
	});
	expect(detailRes.ok()).toBeTruthy();
	const detail = await detailRes.json();
	expect(detail.resource.schedule).toBeTruthy();
	expect(detail.resource.weekly_matchups.length).toBe(1);
	expect(detail.resource.weekly_matchups[0].matchups[0].home_team_id).toBeTruthy();

	// A team captain cannot assign a weekly matchup (commissioner-only).
	const forbidden = await page.request.post(
		`${API}/schedule/${detail.resource.schedule.id}/weekly_matchup`,
		{
			headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': captainAToken },
			data: { week_id: detail.resource.weekly_matchups[0].week_id, matchups: [] }
		}
	);
	expect(forbidden.status()).toBe(403);

	// Clean up the draft so the table doesn't accumulate rows across runs.
	const cleanup = await page.request.delete(`${API}/draft/${draftId}`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(cleanup.ok()).toBeTruthy();
});
