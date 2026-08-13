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

// SeedDevData creates a format named "Men's Intraclub".
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

test('teams list, roster, and promote to co-captain', async ({ page }) => {
	await login(page);
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	const unique = Date.now();

	// One captain plus one roster player is enough to complete a draft.
	const people = [
		{ first: `TCap${unique}`, last: 'One' },
		{ first: `TPlay${unique}`, last: 'One' }
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
	const [captainId, playerId] = [...importBody.Created, ...importBody.AlreadyExisting].map(
		(u: { id: string }) => u.id
	);

	const facilityRes = await page.request.post(`${API}/facility`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: `Team Facility ${unique}`, address: `${unique} Main St`, courts: 4 }
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
		data: { name: `Team Draft ${unique}`, format: format!.id }
	});
	expect(draftRes.ok()).toBeTruthy();
	const draftId = (await draftRes.json()).resource.id as string;

	await page.request.post(`${API}/draft/${draftId}/initialize`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { captains: [captainId] }
	});
	await page.request.post(`${API}/draft/${draftId}/assign_draftable_players`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { players: [playerId] }
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

	const captainToken = await tokenFor(page, `${people[0].first.toLowerCase()}@example.com`);

	// The captain drafts the roster player, then themselves -> 2 picks = complete.
	for (const player of [playerId, captainId]) {
		const res = await page.request.post(`${API}/draft/${draftId}/select`, {
			headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': captainToken },
			data: { player_id: player }
		});
		expect(res.ok(), `pick failed: ${await res.text()}`).toBeTruthy();
	}

	// Finalize: assign the drafted players to their team (creates the
	// team_assignment / team_rating roster rows).
	const assignRes = await page.request.post(`${API}/draft/${draftId}/assign_drafted_players_to_teams`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(assignRes.ok()).toBeTruthy();

	const seasonRes = await page.request.post(`${API}/draft/${draftId}/create_season`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: `Team Season ${unique}`, facility: facilityId, start_time: '08:30' }
	});
	expect(seasonRes.ok()).toBeTruthy();
	const season = (await seasonRes.json()).resource as { id: string };

	// Resolve the team ID the season contains.
	const seasonTeams = (
		await (await page.request.get(`${API}/season_team`, { headers: { 'X-INTRACLUB-TOKEN': token } })).json()
	).resource as { season_id: string; team_id: string }[];
	const seasonTeam = seasonTeams.find((st) => st.season_id === season.id);
	expect(seasonTeam).toBeTruthy();
	const teamId = seasonTeam!.team_id;

	// The commissioner can see the team (via season commissioner access).
	const asCommissioner = await page.request.get(`${API}/team/${teamId}`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(asCommissioner.ok()).toBeTruthy();

	// Log in as the captain (a team member) to exercise the promote UI.
	await page.evaluate((jwt) => localStorage.setItem('intraclub_jwt', jwt), captainToken);
	await page.goto('/teams');
	await waitForHydration(page);

	// The teams list shows the team with the captain + player.
	await expect(page.getByRole('heading', { name: 'Teams' })).toBeVisible();
	const teamLink = page.getByRole('link', { name: /Team 1/ });
	await expect(teamLink).toBeVisible();
	await teamLink.click();
	await page.waitForURL(`/teams/${teamId}`);
	await waitForHydration(page);

	// Members tab lists both members with role badges.
	await expect(page.getByRole('heading', { name: 'Team 1' })).toBeVisible();
	const membersTab = page.getByLabel('Members');
	await expect(membersTab.getByText(people[0].first, { exact: false })).toBeVisible();
	await expect(membersTab.getByText(people[1].first, { exact: false })).toBeVisible();

	// The captain can manage co-captains: promote the player.
	await page.getByRole('tab', { name: 'Co-captains' }).click();
	const promoteButton = page.getByRole('button', { name: 'Promote to co-captain' });
	await expect(promoteButton).toBeVisible();
	await promoteButton.click();

	await expect(page.getByText(`${people[1].first} One is now a co-captain.`)).toBeVisible();

	// After promotion the player is no longer promotable.
	await expect(page.getByRole('button', { name: 'Promote to co-captain' })).not.toBeVisible();

	// The Members tab reflects the new co-captain role.
	await page.getByRole('tab', { name: 'Members' }).click();
	await expect(page.getByLabel('Members').getByText('Co-captain')).toBeVisible();

	// Clean up the draft so the table doesn't accumulate rows across runs.
	const cleanup = await page.request.delete(`${API}/draft/${draftId}`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(cleanup.ok()).toBeTruthy();
});
