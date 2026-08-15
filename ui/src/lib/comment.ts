// Comment and CommentReaction API clients.
//
// Comments attach to a blurb and support replies; reactions attach to
// comments. The read surface is the generic CRUD registered in main.go:
//
//	GET    /api/comment          -> { resource: Comment[] }
//	GET    /api/comment/:id      -> { resource: Comment }
//	POST   /api/comment          -> { resource: Comment }   (owner = token user)
//	PUT    /api/comment/:id      -> { resource: Comment }
//	DELETE /api/comment/:id      -> { resource: Comment }
//	GET    /api/comment_reaction -> { resource: CommentReaction[] }
//
// The interactive writes (react/unreact) go through the custom routes in
// route/comment (see main.go), which enforce season-participation:
//
//	POST /api/comment/:id/react    { reaction: "Fire" }
//	POST /api/comment/:id/unreact  { reaction: "Fire" }
//
// A comment's `reply_to` field holds the id of the comment it replies to
// (empty for a top-level comment). Record IDs are 16-char hex strings.
import { authFetch } from '$lib/auth';

export interface Comment {
	id: string;
	blurb: string;
	reply_to: string;
	user_id: string;
	content: string;
	created_at: string;
	edited_at: string;
}

export interface CommentReaction {
	id: string;
	comment_id: string;
	user_id: string;
	reaction_type: number;
}

const BASE = '/api/comment';

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

export async function listComments(): Promise<Comment[]> {
	return unwrap(await authFetch(BASE));
}

// createComment posts a comment on a blurb. When `replyTo` is provided the
// comment is a reply to that comment (both must be on the same blurb).
export async function createComment(
	blurbId: string,
	content: string,
	replyTo = ''
): Promise<Comment> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ blurb: blurbId, content, reply_to: replyTo })
		})
	);
}

export async function updateComment(id: string, content: string): Promise<Comment> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ content })
		})
	);
}

export async function deleteComment(id: string): Promise<void> {
	await authFetch(`${BASE}/${id}`, { method: 'DELETE' });
}

export async function listCommentReactions(): Promise<CommentReaction[]> {
	return unwrap(await authFetch('/api/comment_reaction'));
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

// reactToComment adds a reaction (by human-readable name) from the current user.
export async function reactToComment(commentId: string, reaction: string): Promise<void> {
	await ensureOk(
		await authFetch(`${BASE}/${commentId}/react`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ reaction })
		})
	);
}

// unreactToComment removes the current user's reaction of the given type.
export async function unreactToComment(commentId: string, reaction: string): Promise<void> {
	await ensureOk(
		await authFetch(`${BASE}/${commentId}/unreact`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ reaction })
		})
	);
}
