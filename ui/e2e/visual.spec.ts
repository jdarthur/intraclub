import { expect, test, type Page } from '@playwright/test';
import { waitForHydration } from './helpers';

// Visual regression baselines (#201). Runs only under the `visual-baselines`
// project (see playwright.config.ts), which the chromium project *depends on*,
// so these tests execute before any other spec — against the freshly-wiped DB
// that `make e2e` leaves behind. That is what makes the signed-in dashboard
// baseline the genuine empty state: no season has been created yet.
//
// Theme is driven via Playwright's `colorScheme` emulation
// (prefers-color-scheme), never by clicking the toggle, so a baseline does not
// depend on the toggle's own markup. mode-watcher defaults to `system` mode,
// which follows the OS preference.
//
// Determinism notes:
//   - None of these pages render time-dependent content (no Week dates, no JWT
//     countdown), and the seeded row / organization are created by the test
//     itself, so there are no random 16-char hex IDs in the captured regions.
//   - The facilities list is filtered to the seeded row before the shot, so
//     other facilities (none at this point, but future-proof) cannot leak in.
//   - Fixture names are fixed per scheme ("... Light" / "... Dark"): the text
//     is byte-identical across runs, and the suffix keeps the light/dark
//     variants from tripping the backend's unique-name constraint while they
//     run in parallel.
const API = 'http://127.0.0.1:8080/api';

// The Go backend runs with --dev-token (see playwright.config.ts), so login
// with the user seeded by SeedDevData (model/dev_data.go).
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

async function token(page: Page): Promise<string> {
	const t = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(t).toBeTruthy();
	return t;
}

const SCHEMES = ['light', 'dark'] as const;

// Every baseline is a viewport shot of a deterministic page: hydration done,
// data loaded, animations and the caret disabled.
async function shot(page: Page, name: string) {
	await expect(page).toHaveScreenshot(name, { animations: 'disabled', caret: 'hide' });
}

for (const scheme of SCHEMES) {
	test(`visual: signed-out hero (${scheme})`, async ({ page }) => {
		await page.emulateMedia({ colorScheme: scheme });
		await page.goto('/');
		await waitForHydration(page);
		// The hero heading is the "data has rendered" barrier.
		await expect(page.getByRole('heading', { name: 'IntraClub', exact: true })).toBeVisible();
		await shot(page, `signed-out-hero-${scheme}.png`);
	});

	test(`visual: signed-in dashboard empty state (${scheme})`, async ({ page }) => {
		await page.emulateMedia({ colorScheme: scheme });
		await login(page);
		await waitForHydration(page);
		// Identity resolved (greets by name) and wave 1+2 finished with no
		// season -> the empty state, not the skeleton or a season dashboard.
		await expect(
			page.getByRole('heading', { name: 'Welcome back, JD Arthur' })
		).toBeVisible();
		await expect(page.getByText('No season yet')).toBeVisible();
		await shot(page, `signed-in-empty-${scheme}.png`);
	});

	test(`visual: facilities list with a seeded row (${scheme})`, async ({ page }) => {
		await page.emulateMedia({ colorScheme: scheme });
		await login(page);
		const auth = await token(page);
		const name = `Baseline Facility ${scheme}`;
		const create = await page.request.post(`${API}/facility`, {
			headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': auth },
			data: { name, address: `1 ${scheme} Way`, courts: 4 }
		});
		expect(create.ok()).toBeTruthy();
		const facilityId = (await create.json()).resource.id as string;
		try {
			await page.goto('/facilities');
			await waitForHydration(page);
			// Isolate the seeded row so no other facility can appear in the shot.
			await page.getByLabel('Filter facilities').fill(name);
			const row = page.getByRole('link', { name, exact: true });
			await expect(row).toHaveCount(1);
			await shot(page, `facilities-list-${scheme}.png`);
		} finally {
			const cleanup = await page.request.delete(`${API}/facility/${facilityId}`, {
				headers: { 'X-INTRACLUB-TOKEN': auth }
			});
			expect(cleanup.ok()).toBeTruthy();
		}
	});

	test(`visual: new facility form (${scheme})`, async ({ page }) => {
		await page.emulateMedia({ colorScheme: scheme });
		await login(page);
		await page.goto('/facilities/new');
		await waitForHydration(page);
		await expect(page.getByRole('heading', { name: 'New facility' })).toBeVisible();
		await shot(page, `facility-new-form-${scheme}.png`);
	});

	test(`visual: organization detail tabs (${scheme})`, async ({ page }) => {
		await page.emulateMedia({ colorScheme: scheme });
		await login(page);
		const auth = await token(page);
		// The Details tab is active by default — this is where the TabsTrigger
		// active-state styling (#194) matters.
		const name = `Baseline Organization ${scheme}`;
		const create = await page.request.post(`${API}/organization`, {
			headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': auth },
			data: { name }
		});
		expect(create.ok()).toBeTruthy();
		const orgId = (await create.json()).resource.id as string;
		try {
			await page.goto(`/organizations/${orgId}`);
			await waitForHydration(page);
			await expect(page.getByRole('heading', { name, exact: true })).toBeVisible();
			await shot(page, `organization-tabs-${scheme}.png`);
		} finally {
			const cleanup = await page.request.delete(`${API}/organization/${orgId}`, {
				headers: { 'X-INTRACLUB-TOKEN': auth }
			});
			expect(cleanup.ok()).toBeTruthy();
		}
	});
}
