// Format API client. Format records are backed by the existing REST CRUD
// routes generated from `model.Format` (see main.go):
//
//	GET    /api/format        -> { resource: Format[] }
//	GET    /api/format/:id    -> { resource: Format }
//	POST   /api/format        -> { resource: Format }   (owner = token user)
//	PUT    /api/format/:id    -> { resource: Format }
//	DELETE /api/format/:id    -> { resource: Format }
//
// A Format is just a name owned by a User; the possible ratings and lines live
// in separate join records (FormatRating / FormatLine) and are future work for
// this page. Record IDs are 16-char hex strings; an empty string represents
// "not set" (InvalidRecordId).
import { authFetch } from '$lib/auth';

export interface Format {
	id: string;
	user_id: string;
	name: string;
}

const BASE = '/api/format';

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

export async function listFormats(): Promise<Format[]> {
	return unwrap(await authFetch(BASE));
}

export async function getFormat(id: string): Promise<Format> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export interface FormatInput {
	name: string;
}

export async function createFormat(input: FormatInput): Promise<Format> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updateFormat(id: string, input: FormatInput): Promise<Format> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deleteFormat(id: string): Promise<void> {
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
