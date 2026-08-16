import { expect, test, type Page } from '@playwright/test';
import { waitForHydration } from './helpers';

// Mobile viewport project (#201). Runs only under the `Mobile Chrome` project
// (see playwright.config.ts) — the mobile-appropriate subset: the nav sheet,
// one list page and the landing page at 375px. The wide-table / draft-board
// specs are deliberately NOT run here (see mobile-nav.spec.ts for #200, which
// stays on the chromium project with its own viewport).
//
// Theme is driven via `colorScheme` emulation, same as visual.spec.ts.
test.use({ viewport: { width: 375, height: 667 } });

const API = 'http://127.0.0.1:8080/api';

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

async function shot(page: Page, name: string) {
	await expect(page).toHaveScreenshot(name, { animations: 'disabled', caret: 'hide' });
}

for (const scheme of SCHEMES) {
	test(`mobile: landing page (${scheme})`, async ({ page }) => {
		await page.emulateMedia({ colorScheme: scheme });
		await page.goto('/');
		await waitForHydration(page);
		await expect(page.getByRole('heading', { name: 'IntraClub', exact: true })).toBeVisible();
		await shot(page, `mobile-landing-${scheme}.png`);
	});

	test(`mobile: nav sheet open (${scheme})`, async ({ page }) => {
		await page.emulateMedia({ colorScheme: scheme });
		await page.goto('/');
		await waitForHydration(page);
		await page.getByRole('button', { name: 'Open menu' }).click();
		const dialog = page.getByRole('dialog');
		await expect(dialog).toBeVisible();
		await shot(page, `mobile-nav-sheet-${scheme}.png`);
	});

	test(`mobile: facilities list with a seeded row (${scheme})`, async ({ page }) => {
		await page.emulateMedia({ colorScheme: scheme });
		await login(page);
		const auth = await token(page);
		// Distinct from the desktop fixture names ("Baseline Facility *") so the
		// two specs' filter queries never match each other's rows.
		const name = `Mobile Facility ${scheme}`;
		const create = await page.request.post(`${API}/facility`, {
			headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': auth },
			data: { name, address: `1 ${scheme} Way`, courts: 4 }
		});
		expect(create.ok()).toBeTruthy();
		const facilityId = (await create.json()).resource.id as string;
		try {
			await page.goto('/facilities');
			await waitForHydration(page);
			await page.getByLabel('Filter facilities').fill(name);
			const row = page.getByRole('link', { name, exact: true });
			await expect(row).toHaveCount(1);
			await shot(page, `mobile-facilities-list-${scheme}.png`);
		} finally {
			const cleanup = await page.request.delete(`${API}/facility/${facilityId}`, {
				headers: { 'X-INTRACLUB-TOKEN': auth }
			});
			expect(cleanup.ok()).toBeTruthy();
		}
	});
}
