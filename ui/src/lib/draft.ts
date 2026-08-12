// Draft API client. Draft records are backed by the REST routes registered in
// route/draft (see main.go):
//
//	GET    /api/draft                -> { resource: Draft[] }
//	GET    /api/draft/:id            -> { resource: Draft }
//	POST   /api/draft                -> { resource: Draft }   (owner = token user)
//	PUT    /api/draft/:id            -> { resource: Draft }
//	DELETE /api/draft/:id            -> { resource: Draft }
//
// Custom action endpoints drive the draft lifecycle:
//
//	PUT    /api/draft/:id/draft_order_pattern   body: { draft_order_pattern }
//	POST   /api/draft/:id/initialize            body: { captains }
//	POST   /api/draft/:id/assign_draftable_players body: { players }
//	POST   /api/draft/:id/assign_rating_cutoff  body: { rating, cutoff }
//	POST   /api/draft/:id/select                body: { player_id }
//	POST   /api/draft/:id/assign_drafted_players_to_teams
//	POST   /api/draft/:id/create_season         body: { name, facility, start_time }
//
// `draft_order_pattern` is exposed on the wire as the pattern's Name() string
// (e.g. "Snake") rather than an opaque interface value. Record IDs are 16-char
// hex strings; an empty string represents "not set" (InvalidRecordId).
import { authFetch } from '$lib/auth';
import type { User } from '$lib/user';

export interface Draft {
	id: string;
	name: string;
	owner: string;
	format: string;
	completed_at: string;
	draft_order_pattern: string;
}

// DraftOrderPattern describes one selectable draft order pattern, as returned
// by GET /api/draft_order_patterns.
export interface DraftOrderPattern {
	name: string;
	description: string;
	example: number[][];
}

const BASE = '/api/draft';
const ORDER_PATTERNS = '/api/draft_order_patterns';

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

export async function listDrafts(): Promise<Draft[]> {
	return unwrap(await authFetch(BASE));
}

export async function getDraft(id: string): Promise<Draft> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

// DraftInput omits `draft_order_pattern`: the pattern defaults to Snake on
// create and is set via setDraftOrderPattern.
export interface DraftInput {
	name: string;
	format: string;
}

export async function createDraft(input: DraftInput): Promise<Draft> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updateDraft(id: string, input: DraftInput): Promise<Draft> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deleteDraft(id: string): Promise<void> {
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

// listDraftOrderPatterns returns the selectable draft order patterns.
export async function listDraftOrderPatterns(): Promise<DraftOrderPattern[]> {
	return unwrap(await authFetch(ORDER_PATTERNS));
}

// setDraftOrderPattern selects a draft's DraftOrderPattern by its Name()
// string (e.g. "Snake").
export async function setDraftOrderPattern(id: string, draftOrderPattern: string): Promise<Draft> {
	return unwrap(
		await authFetch(`${BASE}/${id}/draft_order_pattern`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ draft_order_pattern: draftOrderPattern })
		})
	);
}

// initializeDraft creates the draft's teams + DraftCaptain +
// DraftAvailablePlayer rows from the provided captain ids (in draft order).
export async function initializeDraft(id: string, captains: string[]): Promise<Draft> {
	return unwrap(
		await authFetch(`${BASE}/${id}/initialize`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ captains })
		})
	);
}

// assignDraftablePlayers adds players to the draft's available-to-draft list.
export async function assignDraftablePlayers(id: string, players: string[]): Promise<Draft> {
	return unwrap(
		await authFetch(`${BASE}/${id}/assign_draftable_players`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ players })
		})
	);
}

// assignRatingCutoff assigns a cutoff index to a rating for the draft.
export async function assignRatingCutoff(id: string, rating: string, cutoff: number): Promise<void> {
	await unwrap(
		await authFetch(`${BASE}/${id}/assign_rating_cutoff`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ rating, cutoff })
		})
	);
}

// selectPlayer makes a draft pick on behalf of the authenticated captain.
export async function selectPlayer(id: string, playerId: string): Promise<Draft> {
	return unwrap(
		await authFetch(`${BASE}/${id}/select`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ player_id: playerId })
		})
	);
}

// assignDraftedPlayersToTeams finalizes a completed draft, assigning each
// drafted player to their team with their draft rating.
export async function assignDraftedPlayersToTeams(id: string): Promise<Draft> {
	return unwrap(
		await authFetch(`${BASE}/${id}/assign_drafted_players_to_teams`, {
			method: 'POST'
		})
	);
}

// CreateSeasonInput captures the fields needed to create a Season from a
// completed draft. start_time is a 24-hour "HH:MM" string (e.g. "08:30").
export interface CreateSeasonInput {
	name: string;
	facility: string;
	start_time: string;
}

// createSeason creates the Season associated with a completed draft.
export async function createSeason(id: string, input: CreateSeasonInput): Promise<any> {
	return unwrap(
		await authFetch(`${BASE}/${id}/create_season`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

// DraftSelection is one player's result in a completed draft: the round/pick
// they were taken at, the drafted user, and their assigned rating.
export interface DraftSelection {
	round: number;
	pick: number;
	user: User;
	rating: string;
}

// DraftTeamResults groups one team's DraftSelection rows with its captain and
// draft order. It is one entry in the getDraftResults response.
export interface DraftTeamResults {
	team_id: string;
	captain_id: string;
	draft_order: number;
	selections: DraftSelection[];
}

// DraftResults is the response of getDraftResults: the draft's teams (in draft
// order) with each team's rosters and per-player assigned ratings.
export interface DraftResults {
	teams: DraftTeamResults[];
}

// getDraftResults fetches a read-only summary of a draft's final teams and
// rosters, including each player's assigned rating
// (GET /api/draft/:id/results).
export async function getDraftResults(id: string): Promise<DraftResults> {
	return unwrap(await authFetch(`${BASE}/${id}/results`));
}
