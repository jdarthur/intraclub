import { describe, expect, it } from 'vitest';
import { getTokenExpiry, isSessionExpired } from './session';

// Build a fake JWT with the given `exp` (epoch seconds). The client only reads
// the payload's exp claim, so the header/signature can be placeholders.
function tokenWithExp(exp: number): string {
	const payload = btoa(JSON.stringify({ exp })).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
	return `header.${payload}.sig`;
}

describe('getTokenExpiry', () => {
	it('decodes the exp claim as a millisecond timestamp', () => {
		const exp = 1_700_000_000;
		expect(getTokenExpiry(tokenWithExp(exp))).toBe(exp * 1000);
	});

	it('returns null for a malformed token', () => {
		expect(getTokenExpiry('not-a-jwt')).toBeNull();
		expect(getTokenExpiry('a.b')).toBeNull();
		expect(getTokenExpiry('')).toBeNull();
	});

	it('returns null when the payload has no numeric exp', () => {
		const payload = btoa(JSON.stringify({ sub: 'x' })).replace(/=+$/, '');
		expect(getTokenExpiry(`h.${payload}.s`)).toBeNull();
	});

	it('returns null when the payload is not valid JSON', () => {
		const payload = btoa('not json').replace(/=+$/, '');
		expect(getTokenExpiry(`h.${payload}.s`)).toBeNull();
	});
});

describe('isSessionExpired', () => {
	const nowMs = 1_700_000_000 * 1000;
	const future = nowMs / 1000 + 10;

	it('is false while the token is still valid', () => {
		expect(isSessionExpired(tokenWithExp(future), nowMs)).toBe(false);
	});

	it('is true the instant the token expires', () => {
		expect(isSessionExpired(tokenWithExp(Math.floor(nowMs / 1000)), nowMs)).toBe(true);
	});

	it('is true once the token is in the past', () => {
		expect(isSessionExpired(tokenWithExp(Math.floor(nowMs / 1000) - 5), nowMs)).toBe(true);
	});

	it('is false when the token has no decodable exp (left for the 401 handler)', () => {
		expect(isSessionExpired('malformed', nowMs)).toBe(false);
	});
});
