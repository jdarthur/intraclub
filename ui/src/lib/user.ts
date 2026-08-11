// User API client. User records are backed by the existing REST CRUD routes
// generated from `model.User` (see main.go):
//
//	GET    /api/user      -> { resource: User[] }
//	GET    /api/user/:id  -> { resource: User }
//
// Only reads are exposed (GET one / GET many); writes go through the dedicated
// registration/login routes. Record IDs are 16-char hex strings.
import { authFetch } from '$lib/auth';

export interface User {
	id: string;
	first_name: string;
	last_name: string;
	phone_number: string;
	email: string;
	verified: boolean;
}

const BASE = '/api/user';

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

export async function listUsers(): Promise<User[]> {
	return unwrap(await authFetch(BASE));
}

export async function getUser(id: string): Promise<User> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export function fullName(user: Pick<User, 'first_name' | 'last_name'>): string {
	return `${user.first_name} ${user.last_name}`.trim();
}
