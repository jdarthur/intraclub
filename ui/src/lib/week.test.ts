import { describe, expect, it } from 'vitest';
import { currentWeek, type Week } from './week';

function week(id: string, date: string, closed: boolean): Week {
	return { id, draft_id: 'd', date, note: '', closed };
}

describe('currentWeek', () => {
	it('returns null when every week is closed', () => {
		expect(
			currentWeek([
				week('a', '2026-01-05T08:00:00Z', true),
				week('b', '2026-01-12T08:00:00Z', true)
			])
		).toBeNull();
	});

	it('returns the earliest week when none are closed', () => {
		expect(
			currentWeek([
				week('a', '2026-01-05T08:00:00Z', false),
				week('b', '2026-01-12T08:00:00Z', false)
			])?.id
		).toBe('a');
	});

	it('returns the first unclosed week in a mixed list', () => {
		expect(
			currentWeek([
				week('a', '2026-01-05T08:00:00Z', true),
				week('b', '2026-01-12T08:00:00Z', false),
				week('c', '2026-01-19T08:00:00Z', false)
			])?.id
		).toBe('b');
	});

	it('returns null for an empty list', () => {
		expect(currentWeek([])).toBeNull();
	});

	it('sorts by date rather than relying on input order', () => {
		expect(
			currentWeek([
				week('c', '2026-01-19T08:00:00Z', false),
				week('a', '2026-01-05T08:00:00Z', true),
				week('b', '2026-01-12T08:00:00Z', false)
			])?.id
		).toBe('b');
	});
});
