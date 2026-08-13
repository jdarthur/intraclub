// Schedule API client. Schedules and their weekly matchups are exposed via the
// REST surface registered in route/schedule (see main.go):
//
//	POST   /api/schedule                      body: { season_id }
//	GET    /api/schedule?season_id=:id        -> { resource: ScheduleDetail }
//	GET    /api/schedule/:id                  -> { resource: ScheduleDetail }
//	POST   /api/schedule/:id/weekly_matchup   body: { week_id, matchups }
//
// Only a season commissioner may create a schedule or assign weekly matchups;
// everyone else can view. A ScheduleDetail's `schedule` is null when the
// season has no schedule yet. Record IDs are 16-char hex strings.
import { authFetch } from '$lib/auth';

export interface TeamMatchup {
	home_team_id: string;
	away_team_id: string;
	bye: boolean;
}

export interface WeeklyMatchup {
	id: string;
	week_id: string;
	season_id: string;
	matchups: TeamMatchup[];
}

export interface Schedule {
	id: string;
	season_id: string;
}

export interface ScheduleDetail {
	schedule: Schedule | null;
	commissioners: string[];
	weekly_matchups: WeeklyMatchup[];
}

const BASE = '/api/schedule';

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

export async function createSchedule(seasonId: string): Promise<Schedule> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ season_id: seasonId })
		})
	);
}

export async function getScheduleForSeason(seasonId: string): Promise<ScheduleDetail> {
	return unwrap(await authFetch(`${BASE}?season_id=${seasonId}`));
}

export async function assignWeeklyMatchup(
	scheduleId: string,
	weekId: string,
	matchups: TeamMatchup[]
): Promise<WeeklyMatchup> {
	return unwrap(
		await authFetch(`${BASE}/${scheduleId}/weekly_matchup`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ week_id: weekId, matchups })
		})
	);
}
