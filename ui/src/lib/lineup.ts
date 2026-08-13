// Lineup API client. Weekly lineups are exposed via the REST surface
// registered in route/lineup (see main.go):
//
//	GET  /api/lineup/detail?team_id=&week_id=  -> { resource: LineupDetail }
//	POST /api/lineup/set                       body: { team_id, week_id, pairings }
//	POST /api/lineup/:id/confirm               -> mark lineup confirmed (captain)
//	POST /api/lineup/:id/official              -> mark lineup official (commissioner)
//
// Generic CRUD is also available at /api/lineup and /api/lineup_pairing. A
// lineup is built by a team's captain/co-captain, confirmed, and then marked
// official by the season commissioner. Record IDs are 16-char hex strings.
import { authFetch } from '$lib/auth';

export interface Lineup {
	id: string;
	team_id: string;
	week_id: string;
	confirmed: boolean;
	official: boolean;
}

export interface LineupPairing {
	id: string;
	lineup_id: string;
	team_id: string;
	player1: string;
	player2: string;
	format_line_index: number;
}

export interface FormatLine {
	id: string;
	format_id: string;
	player_1_rating: string;
	player_2_rating: string;
	format_index: number;
}

export interface LineupDetail {
	team_id: string;
	week_id: string;
	lineup: Lineup | null;
	lines: FormatLine[];
	pairings: LineupPairing[];
}

export interface PairingInput {
	player1: string;
	player2: string;
	format_line_index: number;
}

const BASE = '/api/lineup';

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

export async function getLineupDetail(teamId: string, weekId: string): Promise<LineupDetail> {
	return unwrap(await authFetch(`${BASE}/detail?team_id=${teamId}&week_id=${weekId}`));
}

export async function setLineup(
	teamId: string,
	weekId: string,
	pairings: PairingInput[]
): Promise<LineupDetail> {
	return unwrap(
		await authFetch(`${BASE}/set`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ team_id: teamId, week_id: weekId, pairings })
		})
	);
}

export async function confirmLineup(lineupId: string): Promise<Lineup> {
	return unwrap(
		await authFetch(`${BASE}/${lineupId}/confirm`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' }
		})
	);
}

export async function markOfficial(lineupId: string): Promise<Lineup> {
	return unwrap(
		await authFetch(`${BASE}/${lineupId}/official`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' }
		})
	);
}
