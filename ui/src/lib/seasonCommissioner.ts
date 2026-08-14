// SeasonCommissioner API client. A SeasonCommissioner links a Season to a User
// acting as a co-commissioner. The read surface is the generic CRUD registered
// in main.go; writes go through the custom routes in
// route/seasoncommissioner (same paths), which enforce the model's sysadmin-only
// EditableBy constraint:
//
//	GET    /api/season_commissioner      -> { resource: SeasonCommissioner[] }
//	GET    /api/season_commissioner/:id  -> { resource: SeasonCommissioner }
//	POST   /api/season_commissioner      -> { resource: SeasonCommissioner }
//	DELETE /api/season_commissioner/:id  -> { resource: SeasonCommissioner }
//
// Records are readable by everyone; only a system administrator may create or
// delete them (see model.SeasonCommissioner.EditableBy). Record IDs are
// 16-char hex strings.
import { authFetch } from '$lib/auth';

export interface SeasonCommissioner {
	id: string;
	season_id: string;
	user_id: string;
	created_at: string;
	updated_at: string;
}

const BASE = '/api/season_commissioner';

// unwrap parses a CrudCommon response (`{ resource: ... }`) and throws the
// backend error message on non-2xx responses so callers can surface it.
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

export async function listSeasonCommissioners(): Promise<SeasonCommissioner[]> {
	return unwrap(await authFetch(BASE));
}

// addSeasonCommissioner adds a user to a season as a co-commissioner.
export async function addSeasonCommissioner(seasonId: string, userId: string): Promise<SeasonCommissioner> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ season_id: seasonId, user_id: userId })
		})
	);
}

export async function removeSeasonCommissioner(id: string): Promise<void> {
	await authFetch(`${BASE}/${id}`, { method: 'DELETE' });
}
