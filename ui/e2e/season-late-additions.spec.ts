import { expect, test, type Page } from '@playwright/test';
import { waitForHydration } from './helpers';

// The Go backend runs with --dev-token from the repo root (see
// playwright.config.ts), so POST /api/one_time_password returns the magic-link
// token in the response body. The seeded jdarthur@gatech.edu user is the system
// administrator in dev mode (see model.SeedDevData), which is required to
// exercise the sysadmin-only late-addition write surface.
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

// SeedDevData creates a format named "Men's Intraclub".
const SEED_FORMAT = "Men's Intraclub";
const API = 'http://127.0.0.1:8080/api';

// createSeason runs the minimal draft flow (one captain + one roster player)
// and finalizes it into a season, returning the season id.
async function createSeason(page: Page, token: string, unique: number): Promise<string> {
	const people = [
		{ first: `Cap${unique}`, last: 'LA' },
		{ first: `Play${unique}`, last: 'LA' }
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
		data: { name: `Late Add Facility ${unique}`, address: `${unique} Main St`, courts: 4 }
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
		data: { name: `Late Add Draft ${unique}`, format: format!.id }
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
	// Cutoffs are required for every rating except the last (lowest-skill) one.
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

	const seasonRes = await page.request.post(`${API}/draft/${draftId}/create_season`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: `Late Add Season ${unique}`, facility: facilityId, start_time: '08:30' }
	});
	expect(seasonRes.ok()).toBeTruthy();
	return (await seasonRes.json()).resource.id as string;
}

async function tokenFor(page: Page, email: string): Promise<string> {
	const res = await page.request.post(`${API}/one_time_password`, { data: { email } });
	expect(res.ok()).toBeTruthy();
	const { token } = await res.json();
	const jwtRes = await page.request.post(`${API}/token`, { data: { token } });
	expect(jwtRes.ok()).toBeTruthy();
	const { jwt } = await jwtRes.json();
	return jwt;
}

// lateAdditionUserIds returns the user ids of this season's late additions via
// the generic CRUD read surface.
async function lateAdditionUserIds(page: Page, seasonId: string): Promise<string[]> {
	const res = await page.request.get(`${API}/season_late_addition`);
	expect(res.ok()).toBeTruthy();
	const rows = (await res.json()).resource as { season_id: string; user_id: string }[];
	return rows.filter((r) => r.season_id === seasonId).map((r) => r.user_id);
}

test('sysadmin adds and removes late players on a season via the UI', async ({ page }) => {
	await login(page, 'jdarthur@gatech.edu');
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	const unique = Date.now();
	const seasonId = await createSeason(page, token, unique);

	// A user who was not drafted into the season, to be added late.
	const lateName = `Late${unique}`;
	const importRes = await page.request.post(`${API}/import_users_from_csv`, {
		form: {
			file: `First Name, Last Name, Email\n${lateName}, LA, ${lateName.toLowerCase()}@example.com`
		}
	});
	expect(importRes.ok()).toBeTruthy();
	const importBody = await importRes.json();
	const lateUserId = [...importBody.Created, ...importBody.AlreadyExisting].map(
		(u: { id: string }) => u.id
	)[0];

	await page.goto(`/seasons/${seasonId}`);
	await waitForHydration(page);

	// Scope transfer-list interactions to the Late additions card: the
	// season page also shows a Co-commissioners card whose source pool lists
	// the same eligible users.
	const lateAdditionsCard = page
		.locator('[data-slot="card"]')
		.filter({ has: page.getByRole('heading', { name: 'Late additions' }) });

	// The sysadmin sees the Late additions card.
	await expect(page.getByRole('heading', { name: 'Late additions' })).toBeVisible();

	// The new late player is not yet a late addition.
	expect(await lateAdditionUserIds(page, seasonId)).not.toContain(lateUserId);

	// Add the late player through the transfer list.
	await lateAdditionsCard.getByRole('button', { name: `${lateName} LA`, exact: true }).click();
	await lateAdditionsCard.getByRole('button', { name: 'Move selected to pool' }).click();
	await lateAdditionsCard.getByRole('button', { name: 'Save late additions' }).click();
	// The save button stays disabled until the join rows are persisted and the
	// list is re-seeded with the new late addition on the target side.
	await expect(lateAdditionsCard.getByRole('button', { name: 'Save late additions' })).toBeEnabled();
	await expect(lateAdditionsCard.getByRole('button', { name: `${lateName} LA`, exact: true })).toBeVisible();

	// The API reflects the new late addition.
	expect(await lateAdditionUserIds(page, seasonId)).toContain(lateUserId);

	// Remove the late player through the transfer list.
	await lateAdditionsCard.getByRole('button', { name: `${lateName} LA`, exact: true }).click();
	await lateAdditionsCard.getByRole('button', { name: 'Move selected out of pool' }).click();
	await lateAdditionsCard.getByRole('button', { name: 'Save late additions' }).click();
	await expect(lateAdditionsCard.getByRole('button', { name: 'Save late additions' })).toBeEnabled();
	expect(await lateAdditionUserIds(page, seasonId)).not.toContain(lateUserId);

	// The player is selectable again after removal (back in the source pool).
	await expect(lateAdditionsCard.getByRole('button', { name: `${lateName} LA`, exact: true })).toBeVisible();

	// Clean up the late addition row so the table doesn't accumulate across runs.
	const rows = (await (await page.request.get(`${API}/season_late_addition`)).json())
		.resource as { id: string; season_id: string }[];
	for (const row of rows.filter((r) => r.season_id === seasonId)) {
		const del = await page.request.delete(`${API}/season_late_addition/${row.id}`, {
			headers: { 'X-INTRACLUB-TOKEN': token }
		});
		expect(del.ok()).toBeTruthy();
	}
});

test('non-sysadmin does not see the late additions card', async ({ page }) => {
	await login(page, 'jdarthur@gatech.edu');
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	const unique = Date.now();
	const seasonId = await createSeason(page, token, unique);

	// Log in as a seeded non-sysadmin user.
	await page.evaluate(() => localStorage.removeItem('intraclub_jwt'));
	await login(page, 'avery.chen@example.com');

	await page.goto(`/seasons/${seasonId}`);
	await waitForHydration(page);

	await expect(page.getByRole('heading', { name: 'Late additions' })).not.toBeVisible();

	// And the non-sysadmin is rejected by the API write surface too.
	const otherToken = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	const whoami = await (
		await page.request.get(`${API}/whoami`, { headers: { 'X-INTRACLUB-TOKEN': otherToken } })
	).json();
	const res = await page.request.post(`${API}/season_late_addition`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': otherToken },
		data: { season_id: seasonId, user_id: whoami.id }
	});
	expect(res.status()).toBe(403);
});
