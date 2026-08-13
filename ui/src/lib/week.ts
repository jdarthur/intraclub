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
