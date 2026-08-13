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

test('season participant sets availability and captain sees team view', async ({ page }) => {
	await login(page, 'jdarthur@gatech.edu');
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
	const captainAEmail = `${people[0].first.toLowerCase()}@example.com`;

	const facilityRes = await page.request.post(`${API}/facility`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: `Avail Facility ${unique}`, address: `${unique} Main St`, courts: 4 }
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
		data: { name: `Avail Draft ${unique}`, format: format!.id }
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

	// Complete the draft: A, B, B, A snake order.
	const pickPlan = [
		{ player: playerAId, token: captainAToken },
		{ player: playerBId, token: captainBToken },
		{ player: captainBId, token: captainBToken },
		{ player: captainAId, token: captainAToken }
	];
	for (const { player, token: pickToken } of pickPlan) {
		const res = await page.request.post(`${API}/draft/${draftId}/select`, {
			headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': pickToken },
			data: { player_id: player }
		});
		expect(res.ok(), `pick failed: ${await res.text()}`).toBeTruthy();
	}

	const seasonName = `Avail ${unique}`;
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

	// --- log in as a team captain (season participant) and set availability ---
	await login(page, captainAEmail);
	await page.goto(`/seasons/${season.id}`);
	await waitForHydration(page);
	await expect(page.getByRole('heading', { name: seasonName })).toBeVisible();

	const availCard = page.getByText('Player availability', { exact: true });
	await expect(availCard).toBeVisible();

	// A captain sees the team availability table.
	await expect(page.getByText(/team availability/)).toBeVisible();

	// Set captain A's availability for Week 1 to "Available".
	await page.getByLabel(/Week 1/).selectOption({ label: 'Available' });

	// The backend records exactly one availability for captain A for the week.
	await expect
		.poll(async () => {
			const mine = await page.request.get(`${API}/availability?draft_id=${draftId}`, {
				headers: { 'X-INTRACLUB-TOKEN': captainAToken }
			});
			const body = await mine.json();
			return body.resource?.length;
		})
		.toBe(1);

	// The value persists in the team availability table for captain A.
	await expect(page.locator('table').getByText('Available')).toBeVisible();

	// Clean up the draft so the table doesn't accumulate rows across runs.
	const cleanup = await page.request.delete(`${API}/draft/${draftId}`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(cleanup.ok()).toBeTruthy();
});
