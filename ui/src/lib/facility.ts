// Facility API client. Facility records are backed by the existing REST CRUD
// routes generated from `model.Facility` (see main.go):
//
//	GET    /api/facility        -> { resource: Facility[] }
//	GET    /api/facility/:id    -> { resource: Facility }
//	POST   /api/facility        -> { resource: Facility }   (owner = token user)
//	PUT    /api/facility/:id    -> { resource: Facility }
//	DELETE /api/facility/:id    -> { resource: Facility }
//
// Record IDs (and optional layout_photo) are 16-char hex strings; an empty
// string represents "not set" (InvalidRecordId).
import { authFetch } from '$lib/auth';

export interface Facility {
	id: string;
	user_id: string;
	name: string;
	address: string;
	courts: number;
	layout_photo: string;
}

const BASE = '/api/facility';

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

export async function listFacilities(): Promise<Facility[]> {
	return unwrap(await authFetch(BASE));
}

export async function getFacility(id: string): Promise<Facility> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export interface FacilityInput {
	name: string;
	address: string;
	courts: number;
	layout_photo?: string;
}

export async function createFacility(input: FacilityInput): Promise<Facility> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updateFacility(id: string, input: FacilityInput): Promise<Facility> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deleteFacility(id: string): Promise<void> {
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
