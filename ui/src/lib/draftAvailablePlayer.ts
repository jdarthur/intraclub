// DraftAvailablePlayer API client. Records are backed by the REST CRUD routes
// generated from `model.DraftAvailablePlayer` (see route/draft):
//
//	GET    /api/draft_available_player        -> { resource: DraftAvailablePlayer[] }
//	GET    /api/draft_available_player/:id    -> { resource: DraftAvailablePlayer }
//	POST   /api/draft_available_player        -> { resource: DraftAvailablePlayer }
//	PUT    /api/draft_available_player/:id    -> { resource: DraftAvailablePlayer }
//	DELETE /api/draft_available_player/:id    -> { resource: DraftAvailablePlayer }
//
// A DraftAvailablePlayer links a Draft to a User who is eligible to be
// drafted. Record IDs are 16-char hex strings.
import { authFetch } from '$lib/auth';

export interface DraftAvailablePlayer {
	id: string;
	draft_id: string;
	player_id: string;
	created_at: string;
	updated_at: string;
}

const BASE = '/api/draft_available_player';

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

export async function listDraftAvailablePlayers(): Promise<DraftAvailablePlayer[]> {
	return unwrap(await authFetch(BASE));
}

export async function getDraftAvailablePlayer(id: string): Promise<DraftAvailablePlayer> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export interface DraftAvailablePlayerInput {
	draft_id: string;
	player_id: string;
}

export async function createDraftAvailablePlayer(
	input: DraftAvailablePlayerInput
): Promise<DraftAvailablePlayer> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updateDraftAvailablePlayer(
	id: string,
	input: DraftAvailablePlayerInput
): Promise<DraftAvailablePlayer> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deleteDraftAvailablePlayer(id: string): Promise<void> {
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
