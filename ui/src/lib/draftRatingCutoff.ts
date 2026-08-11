// DraftRatingCutoff API client. Records are backed by the REST CRUD routes
// generated from `model.DraftRatingCutoff` (see route/draft):
//
//	GET    /api/draft_rating_cutoff        -> { resource: DraftRatingCutoff[] }
//	GET    /api/draft_rating_cutoff/:id    -> { resource: DraftRatingCutoff }
//	POST   /api/draft_rating_cutoff        -> { resource: DraftRatingCutoff }
//	PUT    /api/draft_rating_cutoff/:id    -> { resource: DraftRatingCutoff }
//	DELETE /api/draft_rating_cutoff/:id    -> { resource: DraftRatingCutoff }
//
// A DraftRatingCutoff assigns a cutoff index (the last selection index matching
// that rating) to a rating for a particular Draft. Record IDs are 16-char hex
// strings.
import { authFetch } from '$lib/auth';

export interface DraftRatingCutoff {
	id: string;
	draft_id: string;
	rating_id: string;
	cutoff_index: number;
}

const BASE = '/api/draft_rating_cutoff';

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

export async function listDraftRatingCutoffs(): Promise<DraftRatingCutoff[]> {
	return unwrap(await authFetch(BASE));
}

export async function getDraftRatingCutoff(id: string): Promise<DraftRatingCutoff> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export interface DraftRatingCutoffInput {
	draft_id: string;
	rating_id: string;
	cutoff_index: number;
}

export async function createDraftRatingCutoff(
	input: DraftRatingCutoffInput
): Promise<DraftRatingCutoff> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updateDraftRatingCutoff(
	id: string,
	input: DraftRatingCutoffInput
): Promise<DraftRatingCutoff> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deleteDraftRatingCutoff(id: string): Promise<void> {
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
