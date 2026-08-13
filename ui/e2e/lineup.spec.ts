import { expect, test, type Page } from '@playwright/test';
import { waitForHydration } from './helpers';

const API = 'http://127.0.0.1:8080/api';

// Log in through the UI (dev mode returns the magic-link token in the
// one_time_password response, rendered as a clickable link on the login page).
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

test('captain builds a weekly lineup, confirms it, and the commissioner marks it official', async ({
	page
}) => {
	await login(page, 'jdarthur@gatech.edu');
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	const unique = Date.now();

	// Two captains + four roster players (two per team) is enough to field a
	// rated pairing on each team: team A ends up with a Men's 1 player and a
	// Men's 2 player, which exactly matches format line index 1.
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
		data: { name: `Lineup Facility ${unique}`, address: `${unique} Main St`, courts: 4 }
	});
	expect(facilityRes.ok()).toBeTruthy();
	const facilityId = (await facilityRes.json()).resource.id as string;

	const formats = (await (await page.request.get(`${API}/format`)).json()).resource as {
		id: string;
		name: string;
	}[];
	const format = formats.find((f) => f.name === "Men's Intraclub");
	expect(format).toBeTruthy();

	const draftRes = await page.request.post(`${API}/draft`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: `Lineup Draft ${unique}`, format: format!.id }
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
		data: { rating: ratings[0].id, cutoff: 1 }
	});
	await page.request.post(`${API}/draft/${draftId}/assign_rating_cutoff`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { rating: ratings[1].id, cutoff: 5 }
	});

	const tokenFor = async (email: string) => {
		const res = await page.request.post(`${API}/one_time_password`, { data: { email } });
		expect(res.ok()).toBeTruthy();
		const { token: otp } = await res.json();
		const jwtRes = await page.request.post(`${API}/token`, { data: { token: otp } });
		expect(jwtRes.ok()).toBeTruthy();
		return (await jwtRes.json()).jwt as string;
	};
	const captainAToken = await tokenFor(captainAEmail);
	const captainBToken = await tokenFor(`${people[1].first.toLowerCase()}@example.com`);

	// Complete the draft: A, B, B, A, A, B snake order. Team A gets picks 1, 4
	// and 5 (playerA with Men's 1, playerD with Men's 2, and the captain), so the
	// two rated non-captains on team A match format line index 1.
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

	// Finalize the draft so drafted players get team_rating rows with their
	// assigned ratings (the captain is pre-assigned and appears via draft_captain).
	const assignRes = await page.request.post(`${API}/draft/${draftId}/assign_drafted_players_to_teams`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token }
	});
	expect(assignRes.ok(), `assign failed: ${await assignRes.text()}`).toBeTruthy();

	const seasonName = `Lineup ${unique}`;
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

	// --- log in as the team captain and build a lineup ---
	await login(page, captainAEmail);
	await page.goto(`/seasons/${season.id}`);
	await waitForHydration(page);
	await expect(page.getByRole('heading', { name: seasonName })).toBeVisible();

	const lineupCard = page.getByText('Weekly lineups', { exact: true });
	await expect(lineupCard).toBeVisible();

	// Build the captain's team's lineup for Week 1.
	await page.getByRole('button', { name: 'Build' }).click();

	// Team A's rated non-captains are playerA (Men's 1) and playerD (Men's 2),
	// which match format line index 1 (Men's 1 / Men's 2).
	const playerALabel = `${people[2].first} ${people[2].last}`;
	const playerDLabel = `${people[5].first} ${people[5].last}`;
	await page.locator(`[id="${week.id}-lineup-p1-1"]`).selectOption({ label: playerALabel });
	await page.locator(`[id="${week.id}-lineup-p2-1"]`).selectOption({ label: playerDLabel });

	await page.getByRole('button', { name: 'Save' }).click();
	await page.getByRole('button', { name: 'Confirm' }).click();
	await expect(page.getByText('Confirmed', { exact: true })).toBeVisible();

	// --- log in as the commissioner and mark the lineup official ---
	await login(page, 'jdarthur@gatech.edu');
	await page.goto(`/seasons/${season.id}`);
	await waitForHydration(page);

	// The confirmed lineup for Team A shows a "Mark official" button.
	await page.getByRole('button', { name: 'Mark official' }).first().click();
	// Once marked official, the button disappears (no more confirmed lineups).
	await expect(page.getByRole('button', { name: 'Mark official' })).toHaveCount(0);

	// Clean up the draft so the table doesn't accumulate rows across runs.
	const cleanup = await page.request.delete(`${API}/draft/${draftId}`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(cleanup.ok()).toBeTruthy();
});
