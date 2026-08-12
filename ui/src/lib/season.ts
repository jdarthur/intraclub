// Season API client. Season records are created through the draft finalize
// flow (POST /api/draft/:id/create_season); this client exposes the read-only
// surface registered in main.go:
//
//	GET  /api/season            -> { resource: Season[] }
//	GET  /api/season/:id        -> { resource: Season }
//	GET  /api/team              -> { resource: Team[] }
//	GET  /api/season_team       -> { resource: SeasonTeam[] }   (join)
//	GET  /api/team_assignment   -> { resource: TeamAssignment[] } (members)
//	GET  /api/team_rating       -> { resource: TeamRating[] }   (draft rating)
//
// Record IDs are 16-char hex strings. A Season's start_time is the daily kickoff
// time as a 24-hour "HH:MM" string (e.g. "08:30").
import { authFetch } from '$lib/auth';

export interface Season {
	id: string;
	name: string;
	facility: string;
	start_time: string;
	draft_id: string;
	schedule_id: string;
	playoff_structure: string;
	owner: string;
}

export interface SeasonTeam {
	id: string;
	season_id: string;
	team_id: string;
	created_at: string;
	updated_at: string;
}

export interface TeamRating {
	id: string;
	team_id: string;
	user_id: string;
	rating_id: string;
}

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

export async function listSeasons(): Promise<Season[]> {
	return unwrap(await authFetch('/api/season'));
}

export async function getSeason(id: string): Promise<Season> {
	return unwrap(await authFetch(`/api/season/${id}`));
}

export async function listSeasonTeams(): Promise<SeasonTeam[]> {
	return unwrap(await authFetch('/api/season_team'));
}

export async function listTeamRatings(): Promise<TeamRating[]> {
	return unwrap(await authFetch('/api/team_rating'));
}
