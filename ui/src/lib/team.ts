// Team API client. Teams are largely immutable after a draft finalizes, so the
// backend exposes them through a constrained surface in route/team rather than
// generic CRUD:
//
//	GET  /api/team              -> { resource: TeamRoster[] }  (teams I can see)
//	GET  /api/team/:id          -> { resource: TeamRoster }
//	POST /api/team/:id/promote_co_captain -> { resource: TeamAssignment }
//
// A TeamRoster bundles a Team with its member assignments (each with a role of
// `captain` / `co_captain` / `member`). Record IDs are 16-char hex strings.
import { authFetch } from '$lib/auth';

export type TeamRole = 'captain' | 'co_captain' | 'member';

export interface TeamColor {
	name: string;
	hex: string;
}

export interface Team {
	id: string;
	name: string;
	color: TeamColor;
	created_at: string;
	updated_at: string;
	deleted_at: string | null;
}

export interface TeamAssignment {
	id: string;
	team_id: string;
	user_id: string;
	role: TeamRole;
	created_at: string;
	updated_at: string;
	deleted_at: string | null;
}

export interface TeamRoster {
	team: Team;
	assignments: TeamAssignment[];
}

// unwrap parses a response (`{ resource: ... }`) and throws the backend error
// message on non-2xx responses so callers can surface it.
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

export async function listTeams(): Promise<TeamRoster[]> {
	return unwrap(await authFetch('/api/team'));
}

export async function getTeam(id: string): Promise<TeamRoster> {
	return unwrap(await authFetch(`/api/team/${id}`));
}

export async function promoteCoCaptain(teamId: string, userId: string): Promise<TeamAssignment> {
	return unwrap(
		await authFetch(`/api/team/${teamId}/promote_co_captain`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ user_id: userId })
		})
	);
}
