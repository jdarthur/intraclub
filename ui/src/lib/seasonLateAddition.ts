// SeasonLateAddition API client. A SeasonLateAddition links a Season to a User
// who joined after the draft was completed. The read surface is the generic
// CRUD registered in main.go; writes go through the custom routes in
// route/lateaddition (same paths), which enforce the model's sysadmin-only
// EditableBy constraint:
//
//	GET    /api/season_late_addition      -> { resource: SeasonLateAddition[] }
//	GET    /api/season_late_addition/:id  -> { resource: SeasonLateAddition }
//	POST   /api/season_late_addition      -> { resource: SeasonLateAddition }
//	DELETE /api/season_late_addition/:id  -> { resource: SeasonLateAddition }
//
// Records are readable by everyone; only a system administrator may create or
// delete them (see model.SeasonLateAddition.EditableBy). Record IDs are
// 16-char hex strings.
import { authFetch } from '$lib/auth';

export interface SeasonLateAddition {
	id: string;
	season_id: string;
	user_id: string;
	created_at: string;
	updated_at: string;
}

const BASE = '/api/season_late_addition';

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

export async function listSeasonLateAdditions(): Promise<SeasonLateAddition[]> {
	return unwrap(await authFetch(BASE));
}

// addSeasonLateAddition adds a user to a season as a late addition.
export async function addSeasonLateAddition(seasonId: string, userId: string): Promise<SeasonLateAddition> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ season_id: seasonId, user_id: userId })
		})
	);
}

export async function removeSeasonLateAddition(id: string): Promise<void> {
	await authFetch(`${BASE}/${id}`, { method: 'DELETE' });
}
