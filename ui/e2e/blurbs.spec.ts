import { expect, test, type Page } from '@playwright/test';
import { waitForHydration } from './helpers';

// The Go backend runs with --dev-token from the repo root (see
// playwright.config.ts), so POST /api/one_time_password returns the magic-link
// token in the response body. The seeded jdarthur@gatech.edu user is the system
// administrator in dev mode (see model.SeedDevData), which is used to create a
// season; the season's captain (a participant) drives the blurb UI flow, since
// commenting/reacting requires season participation.
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

// A 1x1 transparent PNG, used to exercise the photo-upload path end to end.
const PNG_1PX =
	'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==';

// createSeason runs the minimal draft flow (one captain + one roster player)
// and finalizes it into a season, returning the season id. The captain (whose
// email is cap<unique>@example.com) is a season participant.
async function createSeason(page: Page, token: string, unique: number): Promise<string> {
	const people = [
		{ first: `Cap${unique}`, last: 'B' },
		{ first: `Play${unique}`, last: 'B' }
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
		data: { name: `Blurb Facility ${unique}`, address: `${unique} Main St`, courts: 4 }
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
		data: { name: `Blurb Draft ${unique}`, format: format!.id }
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

	const captainToken = await tokenFor(page, `cap${unique}@example.com`);

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
		data: { name: `Blurb Season ${unique}`, facility: facilityId, start_time: '08:30' }
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

test('blurbs feed: create blurb with photo, comment, reply, react, delete', async ({ page }) => {
	// The blurbs page uses window.confirm for destructive actions.
	page.on('dialog', (dialog) => dialog.accept());

	await login(page, 'jdarthur@gatech.edu');
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	const unique = Date.now();
	const seasonId = await createSeason(page, token, unique);

	// Log in as the season's captain (a participant) to drive the UI flow.
	await page.evaluate(() => localStorage.removeItem('intraclub_jwt'));
	await login(page, `cap${unique}@example.com`);

	await page.goto(`/seasons/${seasonId}/blurbs`);
	await waitForHydration(page);

	// --- Create a blurb with an attached photo --------------------------------
	await page.getByLabel('Title').fill('Kickoff announcement');
	await page.getByLabel('Content').fill('Welcome to the season!');
	await page.getByLabel('Attach a photo (optional)').setInputFiles({
		name: 'photo.png',
		mimeType: 'image/png',
		buffer: Buffer.from(PNG_1PX, 'base64')
	});
	// Wait for the FileReader to finish staging the image before submitting.
	await expect(page.getByText('Photo ready to attach')).toBeVisible();
	await page.getByRole('button', { name: 'Post blurb' }).click();
	await expect(page.getByText('Kickoff announcement', { exact: true })).toBeVisible();
	await expect(page.getByText('Welcome to the season!')).toBeVisible();
	await expect(page.locator('img[alt="Kickoff announcement"]')).toBeVisible();

	// --- React to the blurb (Fire) -------------------------------------------
	// The blurb reaction emoji row renders before any comment rows, so .first()
	// targets the blurb's Fire button.
	await page.getByTitle('Fire').first().click();
	await expect(page.getByText('🔥 1')).toBeVisible();

	// --- Add a comment --------------------------------------------------------
	const commentInput = page.getByPlaceholder('Add a comment...');
	await commentInput.fill('Great news!');
	await commentInput.press('Enter');
	await expect(page.getByText('Great news!')).toBeVisible();

	// --- React to the comment (Thumbs up) -------------------------------------
	// The comment's reaction buttons render after the blurb's; for a single
	// comment, .last() is the comment's Thumbs up button.
	await page.getByTitle('Thumbs up').last().click();
	await expect(page.getByText('👍 1')).toBeVisible();

	// --- Reply to the comment --------------------------------------------------
	await page.getByRole('button', { name: 'Reply' }).first().click();
	const replyInput = page.getByPlaceholder('Write a reply...');
	await replyInput.fill('Agreed');
	await replyInput.press('Enter');
	await expect(page.getByText('Agreed')).toBeVisible();

	// --- Delete the comment ----------------------------------------------------
	// The comment's Delete button renders after the blurb's header Delete, so
	// .last() is the comment delete for a single comment.
	await page.getByRole('button', { name: 'Delete', exact: true }).last().click();
	await expect(page.getByText('Great news!')).toHaveCount(0);

	// --- Delete the blurb -------------------------------------------------------
	await page.getByRole('button', { name: 'Delete', exact: true }).first().click();
	await expect(page.getByText('Kickoff announcement', { exact: true })).toHaveCount(0);
	await expect(page.getByText('No blurbs yet for this season.')).toBeVisible();

	// Clean up the season's draft + facility via the API isn't required; the
	// dev DB is ephemeral per e2e run.
});

test('non-participant cannot react to a blurb', async ({ page }) => {
	await login(page, 'jdarthur@gatech.edu');
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	const unique = Date.now();
	const seasonId = await createSeason(page, token, unique);

	// The captain (participant) posts a blurb via the API.
	const captainToken = await tokenFor(page, `cap${unique}@example.com`);
	const blurbRes = await page.request.post(`${API}/blurb`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': captainToken },
		data: { title: 'Private', content: 'Participants only', season: seasonId }
	});
	expect(blurbRes.ok()).toBeTruthy();
	const blurbId = (await blurbRes.json()).resource.id as string;

	// Import a user who is NOT a participant of the season.
	const outsider = `Out${unique}`;
	const importRes = await page.request.post(`${API}/import_users_from_csv`, {
		form: { file: `First Name, Last Name, Email\n${outsider}, B, ${outsider.toLowerCase()}@example.com` }
	});
	expect(importRes.ok()).toBeTruthy();
	const outsiderToken = await tokenFor(page, `${outsider.toLowerCase()}@example.com`);

	// Reacting to the blurb as a non-participant is rejected with 403 (the
	// custom route's participant gate).
	const reactRes = await page.request.post(`${API}/blurb/${blurbId}/react`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': outsiderToken },
		data: { reaction: 'Fire' }
	});
	expect(reactRes.status()).toBe(403);

	// Creating a comment as a non-participant is rejected (400) by the model's
	// DynamicallyValid season-participation check on the generic CRUD create.
	const commentRes = await page.request.post(`${API}/comment`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': outsiderToken },
		data: { blurb: blurbId, content: 'intrusion', reply_to: '' }
	});
	expect(commentRes.status()).toBe(400);

	// Reacting to a participant's comment as a non-participant is 403.
	const captainComment = await page.request.post(`${API}/comment`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': captainToken },
		data: { blurb: blurbId, content: 'hello', reply_to: '' }
	});
	expect(captainComment.ok()).toBeTruthy();
	const commentId = (await captainComment.json()).resource.id as string;
	const commentReactRes = await page.request.post(`${API}/comment/${commentId}/react`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': outsiderToken },
		data: { reaction: 'Thumbs up' }
	});
	expect(commentReactRes.status()).toBe(403);
});
