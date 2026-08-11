// DraftPick API client. Records are backed by the REST CRUD routes generated
// from `model.DraftPick` (see route/draft):
//
//	GET    /api/draft_pick        -> { resource: DraftPick[] }
//	GET    /api/draft_pick/:id    -> { resource: DraftPick }
//	POST   /api/draft_pick        -> { resource: DraftPick }
//	PUT    /api/draft_pick/:id    -> { resource: DraftPick }
//	DELETE /api/draft_pick/:id    -> { resource: DraftPick }
//
// A DraftPick records which user a team selected in a given round/pick, and the
// rating they consequently received. Record IDs are 16-char hex strings.
import { authFetch } from '$lib/auth';

export interface DraftPick {
	id: string;
	draft_id: string;
	team_id: string;
	user_id: string;
	round: number;
	pick: number;
	rating: string;
	created_at: string;
	updated_at: string;
}

const BASE = '/api/draft_pick';

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

export async function listDraftPicks(): Promise<DraftPick[]> {
	return unwrap(await authFetch(BASE));
}

export async function getDraftPick(id: string): Promise<DraftPick> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export interface DraftPickInput {
	draft_id: string;
	team_id: string;
	user_id: string;
	round: number;
	pick: number;
	rating: string;
}

export async function createDraftPick(input: DraftPickInput): Promise<DraftPick> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updateDraftPick(id: string, input: DraftPickInput): Promise<DraftPick> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deleteDraftPick(id: string): Promise<void> {
	const res = await authFetch(`${BASE}/${id}`, { method: 'DELETE' });
	if (!res.ok) {
		let message = `Delete failed (${res.status})`;
		try {
			const body = await res.json();
			if (body?.error) message = body.error;
		} catch {
			// non-JSON error body; fall back to the status message
		}
		throw new Error(message);
	}
}
