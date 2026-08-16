// Week API client. Weeks are exposed via the REST surface registered in
// route/week (see main.go):
//
//	GET  /api/week?season_id=:id  -> { resource: Week[] }
//
// Weeks are viewable by everyone; only a season commissioner can create them.
// Record IDs are 16-char hex strings; `date` is an RFC3339 timestamp.
import { authFetch } from '$lib/auth';

export interface Week {
	id: string;
	draft_id: string;
	date: string;
	note: string;
	/**
	 * Closed is set by the season commissioner once every team match in the
	 * week is complete. A closed week is final; standings are computed from
	 * closed (and completed) weeks' team matches.
	 */
	closed: boolean;
}

/**
 * The current week: the first week (by date) that has not been closed.
 * Null if all weeks are closed.
 *
 * `GET /api/week?season_id=` returns weeks sorted by date, but the list is
 * sorted defensively here so the helper keeps working if that guarantee ever
 * changes.
 */
export function currentWeek(weeks: Week[]): Week | null {
	const byDate = [...weeks].sort(
		(a, b) => new Date(a.date).getTime() - new Date(b.date).getTime()
	);
	return byDate.find((w) => !w.closed) ?? null;
}

async function unwrap(res: Response): Promise<any> {
	if (!res.ok) {
		let message = `Request failed (${res.status})`;
		try {
			const body = await res.json();
			if (body?.error) message = body.error;
		} catch {
			// non-JSON error body; fall back to the status message
		}
		throw new Error(message);
	}
	const body = await res.json();
	return body?.resource;
}

export async function listWeeksForSeason(seasonId: string): Promise<Week[]> {
	return unwrap(await authFetch(`/api/week?season_id=${seasonId}`));
}
