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

/**
 * Result of POST /api/import_users_from_csv. Field names are capitalized
 * because the backend's `route.CsvImportResult` struct has no JSON tags.
 */
export interface CsvImportResult {
	CreatedCount: number;
	Created: User[];
	AlreadyExisting: User[];
}

/**
 * Import users from CSV text. The backend expects the CSV as a urlencoded form
 * field named `file` (headers: `First Name`, `Last Name`, `Email`). Users whose
 * email already exists are skipped and returned in `AlreadyExisting`.
 */
export async function importUsersFromCsv(csv: string): Promise<CsvImportResult> {
	const res = await authFetch('/api/import_users_from_csv', {
		method: 'POST',
		headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
		body: new URLSearchParams({ file: csv })
	});
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
	return (await res.json()) as CsvImportResult;
}

export function fullName(user: Pick<User, 'first_name' | 'last_name'>): string {
	return `${user.first_name} ${user.last_name}`.trim();
}
