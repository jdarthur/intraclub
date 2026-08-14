import { goto } from '$app/navigation';
import { writable } from 'svelte/store';
import { isSessionExpired } from './session';

export { SESSION_EXPIRED_MESSAGE } from './session';

const TOKEN_KEY = 'intraclub_jwt';

/**
 * Set to true when a session ends via token expiry. The landing page reads it
 * to show the "please re-login" message. It is cleared when a new token is set
 * (i.e. the next successful login).
 */
export const sessionExpired = writable(false);

export function getToken(): string | null {
	if (typeof localStorage === 'undefined') return null;
	return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string): void {
	localStorage.setItem(TOKEN_KEY, token);
	sessionExpired.set(false);
}

export function clearToken(): void {
	localStorage.removeItem(TOKEN_KEY);
}

export function isLoggedIn(): boolean {
	return getToken() !== null;
}

// decodeBase64Url decodes a base64url-encoded string (as used in JWT payloads).
function decodeBase64Url(s: string): string {
	const base64 = s.replace(/-/g, '+').replace(/_/g, '/');
	const padded = base64 + '='.repeat((4 - (base64.length % 4)) % 4);
	return atob(padded);
}

// getCurrentUserId returns the authenticated user's id from the stored JWT's
// `sub` claim (the backend puts the user id there; see api/api_auth.go), or
// null if there is no valid token.
export function getCurrentUserId(): string | null {
	const token = getToken();
	if (!token) return null;
	try {
		const payload = token.split('.')[1];
		if (!payload) return null;
		const claims = JSON.parse(decodeBase64Url(payload));
		return typeof claims.sub === 'string' ? claims.sub : null;
	} catch {
		return null;
	}
}

// getRoles fetches the authenticated user's role names from
// GET /api/whoami/roles, e.g. ["System Administrator"]. Returns [] when no
// token is present.
export async function getRoles(): Promise<string[]> {
	const token = getToken();
	if (!token) return [];
	const res = await fetch('/api/whoami/roles', {
		headers: { 'X-INTRACLUB-TOKEN': token }
	});
	if (!res.ok) return [];
	const body = await res.json();
	return Array.isArray(body?.roles) ? body.roles : [];
}

let lastToken: string | null = null;
let expiredHandled = false;

/**
 * Tears down the current session: clears the stored token, flags the session as
 * expired, and navigates to the landing/login page. Idempotent so a burst of
 * concurrent 401s only redirects once.
 */
async function expireSession(): Promise<void> {
	if (expiredHandled) return;
	expiredHandled = true;
	clearToken();
	sessionExpired.set(true);
	try {
		await goto('/');
	} catch {
		// navigation may be rejected (e.g. mid-route-change); the token is
		// already cleared and the store updated, so the next page load will
		// reflect the logged-out state regardless.
	}
}

/**
 * A background ticker that watches the stored token's `exp` claim and expires
 * the session the moment it elapses. Because it re-reads localStorage each tick
 * it naturally restarts on a fresh token (new login) and self-resets its
 * idempotency guard when the token changes. Returns a cleanup function.
 */
export function startSessionMonitor(intervalMs = 1000): () => void {
	const id = setInterval(tick, intervalMs);
	tick();
	return () => clearInterval(id);
}

function tick(): void {
	const token = getToken();
	if (token !== lastToken) {
		lastToken = token;
		expiredHandled = false;
	}
	if (!token || expiredHandled) return;
	if (isSessionExpired(token)) {
		void expireSession();
	}
}

/**
 * fetch wrapper that attaches the auth token and, belt-and-suspenders for clock
 * or ticker drift, expires the session if the API rejects it with a 401. This
 * relies on the backend contract that a 401 is always an invalid/expired token
 * (see api/api_route.go auth gate) — login and magic-link endpoints use raw
 * `fetch` and are unaffected.
 */
export function authFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
	const token = getToken();
	const headers = new Headers(init?.headers);
	if (token) {
		headers.set('X-INTRACLUB-TOKEN', token);
	}
	return fetch(input, { ...init, headers }).then((res) => {
		if (res.status === 401) {
			void expireSession();
		}
		return res;
	});
}
