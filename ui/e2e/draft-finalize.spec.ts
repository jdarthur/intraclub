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

async function tokenFor(page: Page, email: string): Promise<string> {
	const res = await page.request.post(`${API}/one_time_password`, { data: { email } });
	expect(res.ok()).toBeTruthy();
	const { token } = await res.json();
	const jwtRes = await page.request.post(`${API}/token`, { data: { token } });
	expect(jwtRes.ok()).toBeTruthy();
	const { jwt } = await jwtRes.json();
	return jwt;
}

test('finalizing a completed draft assigns rosters and creates a Season', async ({ page }) => {
	await login(page);
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	const unique = Date.now();

	// Two captains plus four roster players = 6 players across 2 teams (3 rounds).
	const people = [
		{ first: `Final${unique}`, last: 'Captain' },
		{ first: `Fini${unique}`, last: 'Captain' },
		{ first: `Rosa${unique}`, last: 'One' },
		{ first: `Rose${unique}`, last: 'Two' },
		{ first: `Rost${unique}`, last: 'Three' },
		{ first: `Rosy${unique}`, last: 'Four' }
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

	// Create a Facility to host the season.
	const facilityName = `River Club ${unique}`;
	const facilityRes = await page.request.post(`${API}/facility`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: facilityName, address: `${unique} Main St`, courts: 4 }
	});
	expect(facilityRes.ok()).toBeTruthy();
	const facilityId = (await facilityRes.json()).resource.id as string;

	const formatRes = await page.request.get(`${API}/format`);
	const formats = (await formatRes.json()).resource as { id: string; name: string }[];
	const format = formats.find((f) => f.name === SEED_FORMAT);
	expect(format).toBeTruthy();

	const draftName = `Finalize Draft ${unique}`;
	const draftRes = await page.request.post(`${API}/draft`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: draftName, format: format!.id }
	});
	expect(draftRes.ok()).toBeTruthy();
	const draftId = (await draftRes.json()).resource.id as string;

	// Set up and complete the draft via the API (see draft-board.spec.ts).
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
	await page.request.post(`${API}/draft/${draftId}/assign_rating_cutoff`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { rating: ratings[0].id, cutoff: 1 }
	});
	await page.request.post(`${API}/draft/${draftId}/assign_rating_cutoff`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { rating: ratings[1].id, cutoff: 5 }
	});

	const alexToken = await tokenFor(page, `${people[0].first.toLowerCase()}@example.com`);
	const blakeToken = await tokenFor(page, `${people[1].first.toLowerCase()}@example.com`);

	// Snake order with 2 teams over 3 rounds: pick 1 Alex, then Blake, Blake,
	// Alex, Alex, Blake.
	const pickOrder = [
		{ captain: alexToken, player: caseyId },
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
		expect(res.ok(), `pick ${player} failed: ${await res.text()}`).toBeTruthy();
	}

	// --- Finalize through the UI ---
	const seasonName = `Season ${unique}`;
	await page.goto(`/drafts/${draftId}/draft`);
	await waitForHydration(page);
	await expect(page.getByText('Completed')).toBeVisible();
	const finalizeButton = page.getByRole('button', { name: 'Finalize draft' });
	await expect(finalizeButton).toBeVisible();
	await finalizeButton.click();

	await page.getByLabel('Season name').fill(seasonName);
	await page.getByLabel('Start time (HH:MM)').fill('08:30');
	await page.getByLabel('Facility').selectOption(facilityId);
	await page.getByRole('button', { name: 'Create season' }).click();

	// Navigation lands on the new Season with the drafted rosters.
	await page.waitForURL(/\/seasons\/.+/, { timeout: 15000 });
	await waitForHydration(page);
	await expect(page.getByRole('heading', { name: seasonName })).toBeVisible();
	await expect(page.getByText('Team 1')).toBeVisible();
	await expect(page.getByText('Team 2')).toBeVisible();

	// Each team shows its drafted members (captains drafted themselves plus
	// their assigned roster players).
	const alexName = `${people[0].first} ${people[0].last}`;
	const blakeName = `${people[1].first} ${people[1].last}`;
	await expect(page.getByText(alexName, { exact: false })).toBeVisible();
	await expect(page.getByText(blakeName, { exact: false })).toBeVisible();
	await expect(page.getByText(`${people[2].first} ${people[2].last}`)).toBeVisible();

	// The backend produced exactly one Season tied to this draft with rosters.
	const seasonRes = await page.request.get(`${API}/season`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(seasonRes.ok()).toBeTruthy();
	const seasons = (await seasonRes.json()).resource as { id: string; draft_id: string }[];
	const season = seasons.find((s) => s.draft_id === draftId);
	expect(season).toBeTruthy();

	// The season links exactly the two drafted teams.
	const seasonTeamsRes = await page.request.get(`${API}/season_team`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(seasonTeamsRes.ok()).toBeTruthy();
	const seasonTeams = (await seasonTeamsRes.json()).resource as {
		season_id: string;
		team_id: string;
	}[];
	const teamIds = seasonTeams
		.filter((st) => st.season_id === season!.id)
		.map((st) => st.team_id);
	expect(teamIds).toHaveLength(2);

	// Every drafted player lands on a season team: drafted non-captain players
	// get a team_rating row, and captains (who draft themselves but are
	// pre-assigned at draft init) appear as draft_captain rows.
	const teamRatingsRes = await page.request.get(`${API}/team_rating`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(teamRatingsRes.ok()).toBeTruthy();
	const rosterRatings = (await teamRatingsRes.json()).resource as {
		team_id: string;
		user_id: string;
	}[];
	const captainsRes = await page.request.get(`${API}/draft_captain`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(captainsRes.ok()).toBeTruthy();
	const allCaptains = (await captainsRes.json()).resource as {
		team_id: string;
		captain_id: string;
	}[];
	const rosterUserIds = new Set<string>([
		...rosterRatings.filter((r) => teamIds.includes(r.team_id)).map((r) => r.user_id),
		...allCaptains.filter((c) => teamIds.includes(c.team_id)).map((c) => c.captain_id)
	]);
	expect(rosterUserIds.size).toBe(userIds.length);
	for (const id of userIds) {
		expect(rosterUserIds.has(id)).toBeTruthy();
	}

	// --- Results page: read-only view of final teams, rosters, ratings, and
	// the created Season ---
	await page.goto(`/drafts/${draftId}/results`);
	await waitForHydration(page);
	// Wait for the page to finish loading (the Season link appears once the
	// draft + results + seasons have all resolved).
	await expect(page.getByRole('link', { name: seasonName })).toBeVisible();
	await expect(page.getByRole('heading', { name: draftName })).toBeVisible();
	// Completion state is shown and links back to the setup page.
	await expect(page.getByText('Completed')).toBeVisible();
	await expect(page.getByRole('link', { name: 'Draft setup' })).toBeVisible();
	// Both teams and their drafted members are listed with assigned ratings.
	await expect(page.getByText('Team 1')).toBeVisible();
	await expect(page.getByText('Team 2')).toBeVisible();
	// Captains appear both in the "Captain:" line and their roster entry, so
	// scope to the roster list item to avoid a strict-mode match.
	await expect(page.getByText(`${alexName} (captain)`)).toBeVisible();
	await expect(page.getByText(`${blakeName} (captain)`)).toBeVisible();
	await expect(page.getByText(`${people[2].first} ${people[2].last}`)).toBeVisible();

	// Clean up the draft so the table doesn't accumulate rows across runs.
	const cleanup = await page.request.delete(`${API}/draft/${draftId}`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(cleanup.ok()).toBeTruthy();
});
