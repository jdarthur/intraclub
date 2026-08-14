// Match API client. Weekly match scoring is exposed via the REST surface
// registered in route/match (see main.go):
//
//	POST /api/match/generate     body: { week_id, scoring_structure_id }
//	GET  /api/match/week?week_id=:id     -> { resource: WeekMatchDetail }
//	POST /api/match/score        body: { individual_match_id, main_value, secondary_value, win_override }
//	POST /api/match/:id/complete -> mark an individual match complete (determines winner)
//	GET  /api/match/standings?season_id=:id -> { resource: StandingsEntry[] }
//
// Only a season commissioner may generate a week's matches; only the match
// editors (the season commissioners, added when matches are generated) and
// sysadmins may record/complete scores. The week score sheet and standings are
// viewable by everyone. Record IDs are 16-char hex strings.
import { authFetch } from '$lib/auth';

export interface IndividualMatch {
	id: string;
	structure: string;
	main_value: number;
	secondary_value: number;
	win_override: boolean;
	/** 0 = unstarted, 1 = in progress, 2 = won, 3 = lost */
	status: number;
	team_id: string;
	player1: string;
	player2: string;
	format_line_index: number;
	opponent: string;
	opponent_status: number;
}

export interface TeamMatch {
	id: string;
	week_id: string;
	home_team_id: string;
	away_team_id: string;
	complete: boolean;
	home_wins: number;
	away_wins: number;
	winner_team_id: string;
	matches: IndividualMatch[];
}

export interface WeekMatchDetail {
	week_id: string;
	season_id: string;
	closed: boolean;
	team_matches: TeamMatch[];
}

export interface StandingsEntry {
	team_id: string;
	wins: number;
	losses: number;
	ties: number;
}

const BASE = '/api/match';

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

export async function generateMatches(weekId: string, scoringStructureId: string): Promise<WeekMatchDetail> {
	return unwrap(
		await authFetch(`${BASE}/generate`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ week_id: weekId, scoring_structure_id: scoringStructureId })
		})
	);
}

export async function getWeekMatches(weekId: string): Promise<WeekMatchDetail> {
	return unwrap(await authFetch(`${BASE}/week?week_id=${weekId}`));
}

export async function recordScore(
	individualMatchId: string,
	mainValue: number,
	secondaryValue: number,
	winOverride = false
): Promise<IndividualMatch> {
	return unwrap(
		await authFetch(`${BASE}/score`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				individual_match_id: individualMatchId,
				main_value: mainValue,
				secondary_value: secondaryValue,
				win_override: winOverride
			})
		})
	);
}

export async function completeMatch(individualMatchId: string): Promise<IndividualMatch> {
	return unwrap(
		await authFetch(`${BASE}/${individualMatchId}/complete`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' }
		})
	);
}

export async function getStandings(seasonId: string): Promise<StandingsEntry[]> {
	return unwrap(await authFetch(`${BASE}/standings?season_id=${seasonId}`));
}
