import { expect, test, type Page } from '@playwright/test';
import { waitForHydration } from './helpers';

// Dedicated mobile spec for #200 (mobile navigation + responsive tables) at a
// 375px viewport. The full mobile-viewport project / visual-regression
// baselines live in #201; this stays small so the two don't duplicate work.
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

// The document must fit inside the viewport: no horizontal body scroll.
async function expectNoHorizontalScroll(page: Page) {
	const overflow = await page.evaluate(() => {
		const doc = document.documentElement;
		return doc.scrollWidth - doc.clientWidth;
	});
	expect(overflow, `document overflows horizontally by ${overflow}px`).toBeLessThanOrEqual(0);
}

test('mobile: sheet menu opens, closes, and navigates', async ({ page }) => {
	await page.goto('/');
	await waitForHydration(page);

	// Exactly one <nav> and exactly one Email-labelled input in the DOM, and no
	// horizontal body scroll at 375px.
	await expect(page.getByRole('navigation')).toHaveCount(1);
	await expect(page.getByLabel('Email')).toHaveCount(1);
	await expectNoHorizontalScroll(page);

	// Desktop nav is hidden at this viewport; the mobile trigger is exposed.
	await expect(page.getByRole('button', { name: 'Settings' })).toBeHidden();
	await expect(page.getByRole('button', { name: 'Open menu' })).toBeVisible();

	// Open: the sheet (a dialog) appears with the nav links. Still exactly one
	// <nav> — the sheet links are not wrapped in a second one.
	await page.getByRole('button', { name: 'Open menu' }).click();
	const dialog = page.getByRole('dialog');
	await expect(dialog).toBeVisible();
	await expect(page.getByRole('link', { name: 'Teams' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Seasons' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'New Draft' })).toBeVisible();
	await expect(page.getByRole('link', { name: 'Organizations' })).toBeVisible();
	await expect(page.getByRole('navigation')).toHaveCount(1);

	// Keyboard: Escape closes the sheet (focus is trapped inside while open).
	await page.keyboard.press('Escape');
	await expect(page.getByRole('link', { name: 'Seasons' })).toHaveCount(0);

	// Re-open and navigate: clicking a link closes the sheet and routes.
	await page.getByRole('button', { name: 'Open menu' }).click();
	await page.getByRole('link', { name: 'Seasons' }).click();
	await page.waitForURL('/seasons');
	await waitForHydration(page);
	await expect(page.getByRole('link', { name: 'Seasons' })).toHaveCount(0);
	await expectNoHorizontalScroll(page);
});

test('mobile: list pages have no horizontal scroll', async ({ page }) => {
	await login(page);
	for (const path of ['/drafts', '/seasons', '/teams', '/users', '/photos', '/organizations']) {
		await page.goto(path);
		await waitForHydration(page);
		await expectNoHorizontalScroll(page);
	}
});

test('mobile: vertical tab page stacks and stays usable', async ({ page }) => {
	await login(page);
	const token = await page.evaluate(() => localStorage.getItem('intraclub_jwt') ?? '');
	expect(token).toBeTruthy();

	// Create an organization via the API so the detail page (vertical tabs)
	// has something to render.
	const unique = Date.now();
	const create = await page.request.post(`${API}/organization`, {
		headers: { 'Content-Type': 'application/json', 'X-INTRACLUB-TOKEN': token },
		data: { name: `Mobile Org ${unique}` }
	});
	expect(create.ok()).toBeTruthy();
	const org = (await create.json()).resource as { id: string };

	await page.goto(`/organizations/${org.id}`);
	await waitForHydration(page);
	await expectNoHorizontalScroll(page);

	// The tab rail renders (stacked at this width) and switching tabs works.
	const membersTab = page.getByRole('tab', { name: 'Members' });
	await expect(membersTab).toBeVisible();
	await membersTab.click();
	await expect(page.getByText('0 members')).toBeVisible();
	await expectNoHorizontalScroll(page);

	// Clean up the organization so the list doesn't accumulate rows.
	const cleanup = await page.request.delete(`${API}/organization/${org.id}`, {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	expect(cleanup.ok()).toBeTruthy();
});
