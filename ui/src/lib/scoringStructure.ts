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
