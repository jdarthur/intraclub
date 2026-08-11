// DraftCaptain API client. Records are backed by the REST CRUD routes
// generated from `model.DraftCaptain` (see route/draft):
//
//	GET    /api/draft_captain        -> { resource: DraftCaptain[] }
//	GET    /api/draft_captain/:id    -> { resource: DraftCaptain }
//	POST   /api/draft_captain        -> { resource: DraftCaptain }
//	PUT    /api/draft_captain/:id    -> { resource: DraftCaptain }
//	DELETE /api/draft_captain/:id    -> { resource: DraftCaptain }
//
// A DraftCaptain links a Draft to a Team and its captain, with a draft order.
// Record IDs are 16-char hex strings.
import { authFetch } from '$lib/auth';

export interface DraftCaptain {
	id: string;
	draft_id: string;
	team_id: string;
	captain_id: string;
	draft_order: number;
	created_at: string;
	updated_at: string;
}

const BASE = '/api/draft_captain';

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

export async function listDraftCaptains(): Promise<DraftCaptain[]> {
	return unwrap(await authFetch(BASE));
}

export async function getDraftCaptain(id: string): Promise<DraftCaptain> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export interface DraftCaptainInput {
	draft_id: string;
	team_id: string;
	captain_id: string;
	draft_order: number;
}

export async function createDraftCaptain(input: DraftCaptainInput): Promise<DraftCaptain> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updateDraftCaptain(
	id: string,
	input: DraftCaptainInput
): Promise<DraftCaptain> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deleteDraftCaptain(id: string): Promise<void> {
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
