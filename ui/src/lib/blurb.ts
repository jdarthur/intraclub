// Blurb, BlurbPhoto, and BlurbReaction API clients.
//
// Blurbs are season-scoped posts (news/announcements) that users can comment
// on and react to. The read surface is the generic CRUD registered in main.go:
//
//	GET    /api/blurb           -> { resource: Blurb[] }
//	GET    /api/blurb/:id       -> { resource: Blurb }
//	POST   /api/blurb           -> { resource: Blurb }   (owner = token user)
//	PUT    /api/blurb/:id       -> { resource: Blurb }
//	DELETE /api/blurb/:id       -> { resource: Blurb }
//	GET    /api/blurb_photo     -> { resource: BlurbPhoto[] }
//	GET    /api/blurb_reaction  -> { resource: BlurbReaction[] }
//
// The interactive writes (react/unreact and photo attach/detach) go through
// the custom routes in route/blurb (see main.go), which enforce
// season-participation and ownership:
//
//	POST   /api/blurb/:id/react      { reaction: "Fire" }
//	POST   /api/blurb/:id/unreact    { reaction: "Fire" }
//	POST   /api/blurb/:id/photos     { photo_id }
//	DELETE /api/blurb/:id/photos/:photoId
//
// Reactions are stored as numeric `reaction_type` values (mirroring
// model.reactionType); the react/unreact routes accept the human-readable
// reaction name. Record IDs are 16-char hex strings.
import { authFetch } from '$lib/auth';

// Reaction type enum values mirror model.reactionType (model/reaction.go).
export const REACTION_THUMBS_UP = 0;
export const REACTION_THUMBS_DOWN = 1;
export const REACTION_LAUGHING = 2;
export const REACTION_FIRE = 3;
export const REACTION_HEART = 4;
export const REACTION_CRYING = 5;

export interface ReactionType {
	value: number;
	name: string;
	emoji: string;
}

// REACTIONS lists the allowed reactions, matching model.AllowedReactions.
// `name` is the human-readable value sent to the react/unreact routes;
// `value` is the numeric `reaction_type` stored on child rows.
export const REACTIONS: ReactionType[] = [
	{ value: REACTION_THUMBS_UP, name: 'Thumbs up', emoji: '👍' },
	{ value: REACTION_THUMBS_DOWN, name: 'Thumbs down', emoji: '👎' },
	{ value: REACTION_LAUGHING, name: 'Laughing', emoji: '😂' },
	{ value: REACTION_FIRE, name: 'Fire', emoji: '🔥' },
	{ value: REACTION_HEART, name: 'Heart', emoji: '❤️' },
	{ value: REACTION_CRYING, name: 'Crying', emoji: '😭' }
];

export const reactionByValue = (value: number): ReactionType =>
	REACTIONS.find((r) => r.value === value) ?? { value, name: 'Unknown', emoji: '❔' };

export interface Blurb {
	id: string;
	title: string;
	content: string;
	owner: string;
	season: string;
}

export interface BlurbPhoto {
	id: string;
	blurb_id: string;
	photo_id: string;
}

export interface BlurbReaction {
	id: string;
	blurb_id: string;
	user_id: string;
	reaction_type: number;
}

const BASE = '/api/blurb';

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

export async function listBlurbs(): Promise<Blurb[]> {
	return unwrap(await authFetch(BASE));
}

export async function getBlurb(id: string): Promise<Blurb> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export async function createBlurb(input: { title: string; content: string; season: string }): Promise<Blurb> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updateBlurb(id: string, input: { title: string; content: string; season: string }): Promise<Blurb> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deleteBlurb(id: string): Promise<void> {
	await authFetch(`${BASE}/${id}`, { method: 'DELETE' });
}

export async function listBlurbPhotos(): Promise<BlurbPhoto[]> {
	return unwrap(await authFetch('/api/blurb_photo'));
}

export async function listBlurbReactions(): Promise<BlurbReaction[]> {
	return unwrap(await authFetch('/api/blurb_reaction'));
}

// ensureOk throws the backend error message on a non-2xx response. Custom-route
// writes return a bare status/JSON, so callers need this to surface failures.
async function ensureOk(res: Response): Promise<void> {
	if (res.ok) return;
	let message = `Request failed (${res.status})`;
	try {
		const body = await res.json();
		if (body?.error) message = body.error;
	} catch {
		// non-JSON error body; fall back to the status message
	}
	throw new Error(message);
}

// reactToBlurb adds a reaction (by human-readable name) from the current user.
export async function reactToBlurb(blurbId: string, reaction: string): Promise<void> {
	await ensureOk(
		await authFetch(`${BASE}/${blurbId}/react`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ reaction })
		})
	);
}

// unreactToBlurb removes the current user's reaction of the given type.
export async function unreactToBlurb(blurbId: string, reaction: string): Promise<void> {
	await ensureOk(
		await authFetch(`${BASE}/${blurbId}/unreact`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ reaction })
		})
	);
}

// addBlurbPhoto attaches a photo (owned by the current user) to a blurb.
export async function addBlurbPhoto(blurbId: string, photoId: string): Promise<void> {
	await ensureOk(
		await authFetch(`${BASE}/${blurbId}/photos`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ photo_id: photoId })
		})
	);
}

// removeBlurbPhoto detaches a photo from a blurb.
export async function removeBlurbPhoto(blurbId: string, photoId: string): Promise<void> {
	await ensureOk(await authFetch(`${BASE}/${blurbId}/photos/${photoId}`, { method: 'DELETE' }));
}
