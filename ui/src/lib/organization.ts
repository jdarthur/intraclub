// Organization API client. Organization records are backed by the existing REST
// CRUD routes generated from `model.Organization` (see main.go), plus custom
// membership routes in route/organization:
//
//	GET    /api/organization              -> { resource: Organization[] }
//	GET    /api/organization/:id          -> { resource: Organization }
//	POST   /api/organization              -> { resource: Organization } (owner = token user)
//	PUT    /api/organization/:id          -> { resource: Organization }
//	DELETE /api/organization/:id          -> { resource: Organization }
//
//	GET    /api/organization/:id/members          -> { resource: User[] }
//	POST   /api/organization/:id/members          -> { resource: OrganizationMember } (body: { user_id })
//	DELETE /api/organization/:id/members/:userId  -> { resource: OrganizationMember }
//
// Record IDs are 16-char hex strings; an empty string represents "not set"
// (InvalidRecordId).
import { authFetch } from '$lib/auth';
import type { User } from '$lib/user';

export interface Organization {
	id: string;
	user_id: string;
	name: string;
	created_at: string;
	updated_at: string;
	deleted_at: string | null;
}

const BASE = '/api/organization';

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

export async function listOrganizations(): Promise<Organization[]> {
	return unwrap(await authFetch(BASE));
}

export async function getOrganization(id: string): Promise<Organization> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export interface OrganizationInput {
	name: string;
}

export async function createOrganization(input: OrganizationInput): Promise<Organization> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updateOrganization(id: string, input: OrganizationInput): Promise<Organization> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deleteOrganization(id: string): Promise<void> {
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

// listMembers returns the current flat membership roster (registered User
// records) of the organization.
export async function listMembers(orgId: string): Promise<User[]> {
	return unwrap(await authFetch(`${BASE}/${orgId}/members`));
}

// addMember adds a registered User to the organization's membership roster.
export async function addMember(orgId: string, userId: string): Promise<void> {
	const res = await authFetch(`${BASE}/${orgId}/members`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ user_id: userId })
	});
	if (!res.ok) {
		let message = `Add member failed (${res.status})`;
		try {
			const body = await res.json();
			if (body?.error) message = body.error;
		} catch {
			// non-JSON error body; fall back to the status message
		}
		throw new Error(message);
	}
}

// removeMember removes a registered User from the organization's membership
// roster.
export async function removeMember(orgId: string, userId: string): Promise<void> {
	const res = await authFetch(`${BASE}/${orgId}/members/${userId}`, { method: 'DELETE' });
	if (!res.ok) {
		let message = `Remove member failed (${res.status})`;
		try {
			const body = await res.json();
			if (body?.error) message = body.error;
		} catch {
			// non-JSON error body; fall back to the status message
		}
		throw new Error(message);
	}
}
