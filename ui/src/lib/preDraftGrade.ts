// PreDraftGrade API client. Records are backed by the REST CRUD routes
// generated from `model.PreDraftGrade` (see route/draft):
//
//	GET    /api/pre_draft_grade        -> { resource: PreDraftGrade[] }
//	GET    /api/pre_draft_grade/:id    -> { resource: PreDraftGrade }
//	POST   /api/pre_draft_grade        -> { resource: PreDraftGrade }
//	PUT    /api/pre_draft_grade/:id    -> { resource: PreDraftGrade }
//	DELETE /api/pre_draft_grade/:id    -> { resource: PreDraftGrade }
//
// A PreDraftGrade is a grade a grader gives a player before the draft,
// expressed as a rating plus a modifier (weak/average/strong). Note: the
// model exposes these fields without JSON tags, so their wire keys are the
// capitalized Go field names (`PlayerId`, `DraftId`, `GraderId`, `Modifier`,
// `Rating`). `GraderId` is set from the authenticated user on create. Record
// IDs are 16-char hex strings.
import { authFetch } from '$lib/auth';

export interface PreDraftGrade {
	id: string;
	PlayerId: string;
	DraftId: string;
	GraderId: string;
	Modifier: number;
	Rating: string;
}

// Modifier values: 0 = weak, 1 = average, 2 = strong.
export type PreDraftModifier = 0 | 1 | 2;

const BASE = '/api/pre_draft_grade';

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

export async function listPreDraftGrades(): Promise<PreDraftGrade[]> {
	return unwrap(await authFetch(BASE));
}

export async function getPreDraftGrade(id: string): Promise<PreDraftGrade> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export interface PreDraftGradeInput {
	PlayerId: string;
	DraftId: string;
	Modifier: PreDraftModifier;
	Rating: string;
	// GraderId is omitted: it is set from the authenticated user on create.
}

export async function createPreDraftGrade(input: PreDraftGradeInput): Promise<PreDraftGrade> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function updatePreDraftGrade(
	id: string,
	input: PreDraftGradeInput
): Promise<PreDraftGrade> {
	return unwrap(
		await authFetch(`${BASE}/${id}`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function deletePreDraftGrade(id: string): Promise<void> {
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
