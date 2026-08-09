// Rating API client. Rating records are backed by the existing REST CRUD
// routes generated from `model.Rating` (see main.go):
//
//	GET    /api/rating        -> { resource: Rating[] }
//	GET    /api/rating/:id    -> { resource: Rating }
//	POST   /api/rating        -> { resource: Rating }   (owner = token user)
//	PUT    /api/rating/:id    -> { resource: Rating }
//	DELETE /api/rating/:id    -> { resource: Rating }
//
// A Rating is a skill rating (e.g. "2.5") with a name and description, owned
// by a User. Record IDs are 16-char hex strings; an empty string represents
// "not set" (InvalidRecordId).
import { authFetch } from '$lib/auth';

export interface Rating {
	id: string;
	user_id: string;
	name: string;
	description: string;
}

const BASE = '/api/rating';

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

export async function listRatings(): Promise<Rating[]> {
	return unwrap(await authFetch(BASE));
}

export async function getRating(id: string): Promise<Rating> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export interface RatingInput {
	name: string;
	description: string;
}

export async function createRating(input: RatingInput): Promise<Rating> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updateRating(id: string, input: RatingInput): Promise<Rating> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deleteRating(id: string): Promise<void> {
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
