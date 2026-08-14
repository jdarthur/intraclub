import { expect, test, type Page } from '@playwright/test';
import { waitForHydration } from './helpers';

// The Go backend runs with --dev-token from the repo root (see
// playwright.config.ts), so POST /api/one_time_password returns the magic-link
// token in the response body. Login with the user seeded by SeedDevData.
async function login(page: Page, email: string) {
	await page.goto('/login');
	await page.waitForLoadState('networkidle');
	await page.getByLabel('Email').fill(email);
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

test('commissioner generates week matches, records scores, completes a match, and standings reflect it', async ({
	page
}) => {
	await login(page, 'jdarthur@gatech.edu');
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	const unique = Date.now();

	// Two captains + four roster players. With rating cutoffs Men's 1 = 5 and
	// Men's 2 = 6, draft picks 1-5 are all Men's 1 and pick 6 is Men's 2, so
	// both teams end up with two Men's 1 players and can field a (Men's 1 /
	// Men's 1) pairing at format line index 0.
	const people = [
		{ first: `CapA${unique}`, last: 'One' },
		{ first: `CapB${unique}`, last: 'Two' },
		{ first: `PlayA${unique}`, last: 'One' },
		{ first: `PlayB${unique}`, last: 'Two' },
		{ first: `PlayC${unique}`, last: 'Three' },
		{ first: `PlayD${unique}`, last: 'Four' }
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
	const playerCId = idFor(`${people[4].first.toLowerCase()}@example.com`);
	const playerDId = idFor(`${people[5].first.toLowerCase()}@example.com`);
	const captainAEmail = `${people[0].first.toLowerCase()}@example.com`;

	const facilityRes = await page.request.post(`${API}/facility`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: `Match Facility ${unique}`, address: `${unique} Main St`, courts: 4 }
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
		data: { name: `Match Draft ${unique}`, format: format!.id }
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
		data: { players: [playerAId, playerBId, playerCId, playerDId] }
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
		data: { rating: ratings[0].id, cutoff: 5 }
	});
	await page.request.post(`${API}/draft/${draftId}/assign_rating_cutoff`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { rating: ratings[1].id, cutoff: 6 }
	});

	const captainAToken = await tokenFor(page, captainAEmail);
	const captainBToken = await tokenFor(page, `${people[1].first.toLowerCase()}@example.com`);

	// Snake order A, B, B, A, A, B. Picks 1-5 are Men's 1 (cutoff 5) and pick 6
	// is Men's 2 (cutoff 6), so each team fields two Men's 1 non-captain players.
	const pickPlan = [
		{ player: playerAId, token: captainAToken },
		{ player: playerBId, token: captainBToken },
		{ player: playerCId, token: captainBToken },
		{ player: playerDId, token: captainAToken },
		{ player: captainAId, token: captainAToken },
		{ player: captainBId, token: captainBToken }
	];
	for (const { player, token: pickToken } of pickPlan) {
		const res = await page.request.post(`${API}/draft/${draftId}/select`, {
			headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': pickToken },
			data: { player_id: player }
		});
		expect(res.ok(), `pick failed: ${await res.text()}`).toBeTruthy();
	}

	const assignRes = await page.request.post(`${API}/draft/${draftId}/assign_drafted_players_to_teams`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token }
	});
	expect(assignRes.ok(), `assign failed: ${await assignRes.text()}`).toBeTruthy();

	const seasonName = `Matches ${unique}`;
	const seasonRes = await page.request.post(`${API}/draft/${draftId}/create_season`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: seasonName, facility: facilityId, start_time: '08:30' }
	});
	expect(seasonRes.ok()).toBeTruthy();
	const season = (await seasonRes.json()).resource as { id: string };

	const weekRes = await page.request.post(`${API}/week`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { draft_id: draftId, date: '2026-01-05T08:00:00Z', note: 'Week 1' }
	});
	expect(weekRes.ok()).toBeTruthy();
	const week = (await weekRes.json()).resource as { id: string };

	// Resolve the two teams from the draft captains (captain A = Team 1).
	const captains = (
		await (await page.request.get(`${API}/draft_captain`)).json()
	).resource as { team_id: string; captain_id: string }[];
	const teamA = captains.find((c) => c.captain_id === captainAId)!.team_id;
	const teamB = captains.find((c) => c.captain_id === captainBId)!.team_id;
	expect(teamA).toBeTruthy();
	expect(teamB).toBeTruthy();

	// Create a schedule with a weekly matchup (Team A home vs Team B away).
	const scheduleRes = await page.request.post(`${API}/schedule`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { season_id: season.id }
	});
	expect(scheduleRes.ok()).toBeTruthy();
	const schedule = (await scheduleRes.json()).resource as { id: string };
	const assignRes2 = await page.request.post(`${API}/schedule/${schedule.id}/weekly_matchup`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: {
			week_id: week.id,
			matchups: [{ home_team_id: teamA, away_team_id: teamB, bye: false }]
		}
	});
	expect(assignRes2.ok(), `weekly matchup failed: ${await assignRes2.text()}`).toBeTruthy();

	// Build and confirm both teams' lineups (Men's 1 / Men's 1 at line 0), then
	// mark them official as the commissioner.
	const setLineup = async (captainToken: string, teamId: string, p1: string, p2: string) => {
		const res = await page.request.post(`${API}/lineup/set`, {
			headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': captainToken },
			data: {
				team_id: teamId,
				week_id: week.id,
				pairings: [{ player1: p1, player2: p2, format_line_index: 0 }]
			}
		});
		expect(res.ok(), `set lineup failed: ${await res.text()}`).toBeTruthy();
		const detail = await res.json();
		return detail.resource.lineup.id as string;
	};
	const lineupA = await setLineup(captainAToken, teamA, playerAId, playerDId);
	const lineupB = await setLineup(captainBToken, teamB, playerBId, playerCId);
	const confirmA = await page.request.post(`${API}/lineup/${lineupA}/confirm`, {
		headers: { 'X-INTRACLUB-TOKEN': captainAToken }
	});
	expect(confirmA.ok(), `confirm A failed: ${await confirmA.text()}`).toBeTruthy();
	const confirmB = await page.request.post(`${API}/lineup/${lineupB}/confirm`, {
		headers: { 'X-INTRACLUB-TOKEN': captainBToken }
	});
	expect(confirmB.ok(), `confirm B failed: ${await confirmB.text()}`).toBeTruthy();
	for (const lineupId of [lineupA, lineupB]) {
		const officialRes = await page.request.post(`${API}/lineup/${lineupId}/official`, {
			headers: { 'X-INTRACLUB-TOKEN': token }
		});
		expect(officialRes.ok(), `official failed: ${await officialRes.text()}`).toBeTruthy();
	}

	// --- generate the week's matches through the UI ---
	await page.goto(`/seasons/${season.id}`);
	await waitForHydration(page);
	await expect(page.getByRole('heading', { name: seasonName })).toBeVisible();

	const scoringCard = page.getByText('Match scoring', { exact: true });
	await expect(scoringCard).toBeVisible();

	await page.getByRole('button', { name: 'Generate matches' }).click();
	// Generation is async; wait for the generate button to disappear (the card
	// swaps from the generate form to the rendered team matches).
	await expect(page.getByRole('button', { name: /Generat/ })).toHaveCount(0);
	await expect(
		page
			.locator('[data-slot="card"]')
			.filter({ hasText: 'Match scoring' })
			.getByText('Team 1 vs Team 2')
	).toBeVisible();

	// Read the generated week score sheet to learn the individual match ids.
	const weekRes2 = await page.request.get(`${API}/match/week?week_id=${week.id}`);
	expect(weekRes2.ok()).toBeTruthy();
	const detail = (await weekRes2.json()).resource;
	expect(detail.team_matches.length).toBe(1);
	const tm = detail.team_matches[0];
	expect(tm.matches.length).toBe(2);
	const homeMatch = tm.matches.find((m: { team_id: string }) => m.team_id === teamA);
	const awayMatch = tm.matches.find((m: { team_id: string }) => m.team_id === teamB);
	expect(homeMatch).toBeTruthy();
	expect(awayMatch).toBeTruthy();

	// A match li scoped to an individual match's score input.
	const matchLi = (matchId: string) =>
		page.locator(`xpath=//input[@id="${matchId}-main"]/ancestor::li[1]`);

	// Record home 6 - away 3, then complete the home side.
	await page.locator(`[id="${homeMatch.id}-main"]`).fill('6');
	await matchLi(homeMatch.id).getByRole('button', { name: 'Save' }).click();
	await page.locator(`[id="${awayMatch.id}-main"]`).fill('3');
	await matchLi(awayMatch.id).getByRole('button', { name: 'Save' }).click();

	// Completing before both sides are scored would fail; both are now scored.
	await matchLi(homeMatch.id).getByRole('button', { name: 'Complete' }).click();

	// The team match now shows the result: 1-0, Team 1 win.
	await expect(page.getByText('1-0 · Team 1 win', { exact: true })).toBeVisible();

	// Standings reflect the completed team match.
	const standingsCard = page.getByText('Standings', { exact: true });
	await expect(standingsCard).toBeVisible();
	await expect(page.getByRole('cell', { name: 'Team 1' })).toBeVisible();
	const standingsRow = page.locator('tr', { hasText: 'Team 1' });
	await expect(standingsRow).toContainText('1');

	// Clean up the draft so the table doesn't accumulate rows across runs.
	const cleanup = await page.request.delete(`${API}/draft/${draftId}`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(cleanup.ok()).toBeTruthy();
});
