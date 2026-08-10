// Ruleset API client. Ruleset records are backed by the REST CRUD routes
// generated from `model.Ruleset` (see main.go) plus a custom amend route:
//
//	GET    /api/ruleset            -> { resource: Ruleset[] }
//	GET    /api/ruleset/:id        -> { resource: Ruleset }
//	POST   /api/ruleset            -> { resource: Ruleset }   (owner = token user)
//	PUT    /api/ruleset/:id/amend  -> { resource: Ruleset }   (new superseding revision)
//	DELETE /api/ruleset/:id        -> { resource: Ruleset }
//
// A Ruleset is the club rulebook. Direct modification is forbidden by the
// model (use Ruleset.Amend instead), so editing a name produces a NEW revision
// that supersedes the current one. Record IDs are 16-char hex strings; an
// empty string represents "not set" (InvalidRecordId).
import { authFetch } from '$lib/auth';

export interface Ruleset {
	id: string;
	name: string;
	revision: number;
	superseded_by: string;
	date: string;
	owner: string;
}

const BASE = '/api/ruleset';

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

export async function listRulesets(): Promise<Ruleset[]> {
	return unwrap(await authFetch(BASE));
}

export async function getRuleset(id: string): Promise<Ruleset> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export async function createRuleset(name: string): Promise<Ruleset> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name })
		})
	);
}

// amendRulesetName edits a Ruleset's name by producing a new revision that
// supersedes the current one; it returns the NEW superseding Ruleset.
export async function amendRulesetName(id: string, name: string): Promise<Ruleset> {
	return unwrap(
		await authFetch(`${BASE}/${id}/amend`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name })
		})
	);
}

export async function deleteRuleset(id: string): Promise<void> {
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
