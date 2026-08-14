// ScoringStructure API client. ScoringStructure records are backed by the
// existing REST CRUD routes generated from `model.ScoringStructure` (see
// main.go) plus the score-counting-type enumeration endpoint:
//
//	GET    /api/scoring_structure        -> { resource: ScoringStructure[] }
//	GET    /api/scoring_structure/:id    -> { resource: ScoringStructure }
//	POST   /api/scoring_structure        -> { resource: ScoringStructure }  (owner = token user)
//	PUT    /api/scoring_structure/:id    -> { resource: ScoringStructure }
//	DELETE /api/scoring_structure/:id    -> { resource: ScoringStructure }
//	GET    /api/score_counting_types     -> { resource: ScoreCountingType[] }
//	GET    /api/scoring_structure/:id/secondary_scoring_structures
//	      -> { resource: ScoringStructure[] }  (ordered by SecondaryIndex)
//	PUT    /api/scoring_structure/:id/secondary_scoring_structures
//	      -> { resource: ScoringStructure[] }  (body: { secondary_scoring_structures: string[] })
//
// The PUT replaces the structure's entire ordered list of secondary
// (tie-breaker) scoring structure references; an empty list makes the
// structure non-composite. Only the structure's owner (or a sysadmin) may set
// them.
//
// A ScoringStructure is how a match is scored (win condition + counting
// type), owned by a User. Record IDs are 16-char hex strings; an empty string
// represents "not set" (InvalidRecordId).
import { authFetch } from '$lib/auth';

export interface WinCondition {
	win_threshold: number;
	must_win_by: number;
	instant_win_threshold: number;
}

export interface ScoringStructure {
	id: string;
	owner: string;
	name: string;
	win_condition_counting_type: number;
	win_condition: WinCondition;
}

export interface ScoreCountingType {
	type: number;
	name: string;
}

// ScoreCountingType values, mirroring model.ScoreCountingType in Go.
export const ScoreCountingTypes = {
	Point: 0,
	Game: 1,
	Set: 2
} as const;

// secondaryCountingTypeFor returns the score-counting type that secondary
// (tie-breaker) structures must use for a primary with the given counting
// type (model.ScoreCountingType.Secondary), or null if the primary cannot be
// composite (Point-based win conditions).
export function secondaryCountingTypeFor(type: number): number | null {
	switch (type) {
		case ScoreCountingTypes.Set:
			return ScoreCountingTypes.Game;
		case ScoreCountingTypes.Game:
			return ScoreCountingTypes.Point;
		default:
			return null;
	}
}

// maximumScoreCountingUnits mirrors model.ScoringStructure.
// MaximumScoreCountingUnitsPlayed: the maximum number of primary units that
// can be played under this win condition, which a composite structure must
// score via secondary structures. Returns null if the win condition cannot be
// composite (win-by-2-or-more without an instant-win threshold).
export function maximumScoreCountingUnits(winCondition: WinCondition): number | null {
	if (winCondition.instant_win_threshold > 0) {
		return winCondition.instant_win_threshold * 2 - 1;
	}
	if (winCondition.must_win_by > 1) {
		return null;
	}
	return winCondition.win_threshold * 2 - 1;
}

export interface ScoringStructureInput {
	name: string;
	win_condition_counting_type: number;
	win_condition: WinCondition;
}

const BASE = '/api/scoring_structure';
const COUNTING_TYPES_BASE = '/api/score_counting_types';

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

export async function listScoringStructures(): Promise<ScoringStructure[]> {
	return unwrap(await authFetch(BASE));
}

export async function getScoringStructure(id: string): Promise<ScoringStructure> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export async function getScoreCountingTypes(): Promise<ScoreCountingType[]> {
	return unwrap(await authFetch(COUNTING_TYPES_BASE));
}

export async function createScoringStructure(input: ScoringStructureInput): Promise<ScoringStructure> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updateScoringStructure(id: string, input: ScoringStructureInput): Promise<ScoringStructure> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

// getScoringStructureSecondaries returns the structure's secondary
// (tie-breaker) scoring structures as full records, ordered by SecondaryIndex.
export async function getScoringStructureSecondaries(id: string): Promise<ScoringStructure[]> {
	return unwrap(await authFetch(`${BASE}/${id}/secondary_scoring_structures`));
}

// setScoringStructureSecondaries replaces the structure's entire ordered list
// of secondary scoring structure references (an empty list makes it
// non-composite), returning the updated list ordered by SecondaryIndex.
export async function setScoringStructureSecondaries(
	id: string,
	secondaryIds: string[]
): Promise<ScoringStructure[]> {
	return unwrap(
		await authFetch(`${BASE}/${id}/secondary_scoring_structures`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ secondary_scoring_structures: secondaryIds })
		})
	);
}

export async function deleteScoringStructure(id: string): Promise<void> {
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
