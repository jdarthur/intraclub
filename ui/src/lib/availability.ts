// Availability API client. Player availability is exposed via the REST surface
// registered in route/availability (see main.go):
//
//	POST /api/availability/set               body: { week_id, available } -> { resource: Availability }
//	GET  /api/availability?draft_id=:id      -> { resource: Availability[] }  (own availability)
//	GET  /api/availability?draft_id=:id&user_id=:id -> { resource: Availability[] } (sysadmin only)
//	GET  /api/availability/team?team_id=:id&draft_id=:id -> { resource: TeamAvailabilityEntry[] } (captain/co-captain)
//
// Availability options are numeric: 0 = unset, 1 = available, 2 = maybe,
// 3 = not available. Only a season participant may set availability, and only
// for themselves. Team availability is viewable by a team's captain/co-captain.
// Record IDs are 16-char hex strings.
import { authFetch } from '$lib/auth';

export const AVAILABILITY_UNSET = 0;
export const AVAILABILITY_AVAILABLE = 1;
export const AVAILABILITY_MAYBE = 2;
export const AVAILABILITY_NOT_AVAILABLE = 3;

export type AvailabilityOption = typeof AVAILABILITY_UNSET | typeof AVAILABILITY_AVAILABLE | typeof AVAILABILITY_MAYBE | typeof AVAILABILITY_NOT_AVAILABLE;

export interface Availability {
	id: string;
	user_id: string;
	week_id: string;
	available: AvailabilityOption;
}

export interface TeamAvailabilityEntry {
	user_id: string;
	availabilities: Availability[];
}

export const AVAILABILITY_OPTIONS: { value: AvailabilityOption; label: string }[] = [
	{ value: AVAILABILITY_UNSET, label: 'Unset' },
	{ value: AVAILABILITY_AVAILABLE, label: 'Available' },
	{ value: AVAILABILITY_MAYBE, label: 'Maybe' },
	{ value: AVAILABILITY_NOT_AVAILABLE, label: 'Not available' }
];

export function availabilityLabel(option: number): string {
	return AVAILABILITY_OPTIONS.find((o) => o.value === option)?.label ?? 'Unset';
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

export async function setAvailability(weekId: string, available: number): Promise<Availability> {
	return unwrap(
		await authFetch('/api/availability/set', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ week_id: weekId, available })
		})
	);
}

export async function getAvailabilityForUser(draftId: string, userId?: string): Promise<Availability[]> {
	const query = userId ? `?draft_id=${draftId}&user_id=${userId}` : `?draft_id=${draftId}`;
	return unwrap(await authFetch(`/api/availability${query}`));
}

export async function getAvailabilityForTeam(teamId: string, draftId: string): Promise<TeamAvailabilityEntry[]> {
	return unwrap(await authFetch(`/api/availability/team?team_id=${teamId}&draft_id=${draftId}`));
}
