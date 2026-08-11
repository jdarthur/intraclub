// Pure JWT-expiry helpers. Kept free of SvelteKit / DOM dependencies so the
// countdown logic can be unit-tested in isolation (see session.test.ts).
export const SESSION_EXPIRED_MESSAGE = 'Your session expired. Please log in again.';

/**
 * Decodes the `exp` claim of a JWT (an epoch-seconds integer) and returns it as
 * a millisecond timestamp, or null if the token is malformed or has no `exp`.
 * The signature/header are not verified here — the client only needs the expiry
 * to drive the countdown; the server is the authority on validity.
 */
export function getTokenExpiry(token: string): number | null {
	try {
		const part = token.split('.')[1];
		if (!part) return null;
		const b64 = part.replace(/-/g, '+').replace(/_/g, '/');
		const padded = b64.padEnd(Math.ceil(b64.length / 4) * 4, '=');
		const payload = JSON.parse(atob(padded));
		if (typeof payload?.exp === 'number') return payload.exp * 1000;
	} catch {
		// malformed token — caller decides how to handle it
	}
	return null;
}

/**
 * Returns true when the token has an `exp` claim at or before `now`. A token
 * with no decodable `exp` is treated as not-expired (left for the 401 handler).
 */
export function isSessionExpired(token: string, now: number = Date.now()): boolean {
	const exp = getTokenExpiry(token);
	if (exp === null) return false;
	return now >= exp;
}
