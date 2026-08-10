// PlayoffStructure API client. PlayoffStructure records are backed by the
// existing REST CRUD routes generated from `model.PlayoffStructure` (see
// main.go):
//
//	GET    /api/playoff_structure     -> { resource: PlayoffStructure[] }
//	GET    /api/playoff_structure/:id -> { resource: PlayoffStructure }
//	POST   /api/playoff_structure     -> { resource: PlayoffStructure }  (owner = token user)
//	PUT    /api/playoff_structure/:id -> { resource: PlayoffStructure }
//	DELETE /api/playoff_structure/:id -> { resource: PlayoffStructure }
//
// A PlayoffStructure is the playoff bracket shape (number of teams that make
// the playoffs, and how many get a bye week), owned by a User. Record IDs are
// 16-char hex strings; an empty string represents "not set" (InvalidRecordId).
import { authFetch } from '$lib/auth';

export interface PlayoffStructure {
	id: string;
	user_id: string;
	byes: number;
	number_of_teams: number;
}

export interface PlayoffStructureInput {
	byes: number;
	number_of_teams: number;
}

const BASE = '/api/playoff_structure';

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

export async function listPlayoffStructures(): Promise<PlayoffStructure[]> {
	return unwrap(await authFetch(BASE));
}

export async function getPlayoffStructure(id: string): Promise<PlayoffStructure> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export async function createPlayoffStructure(input: PlayoffStructureInput): Promise<PlayoffStructure> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updatePlayoffStructure(id: string, input: PlayoffStructureInput): Promise<PlayoffStructure> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deletePlayoffStructure(id: string): Promise<void> {
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
