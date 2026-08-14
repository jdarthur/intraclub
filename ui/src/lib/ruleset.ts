// Ruleset API client. Ruleset records are backed by the REST CRUD routes
// generated from `model.Ruleset` (see main.go) plus a custom amend route:
//
//	GET    /api/ruleset            -> { resource: Ruleset[] }
//	GET    /api/ruleset/:id        -> { resource: Ruleset }
//	POST   /api/ruleset            -> { resource: Ruleset }   (owner = token user)
//	PUT    /api/ruleset/:id/amend  -> { resource: Ruleset }   (new superseding revision)
//	DELETE /api/ruleset/:id        -> { resource: Ruleset }
//
// A Ruleset is the club rulebook. Direct modification is forbidden by the
// model (use Ruleset.Amend instead), so editing a name produces a NEW revision
// that supersedes the current one. Record IDs are 16-char hex strings; an
// empty string represents "not set" (InvalidRecordId).
import { authFetch } from '$lib/auth';

export interface Ruleset {
	id: string;
	name: string;
	revision: number;
	superseded_by: string;
	date: string;
	owner: string;
}

const BASE = '/api/ruleset';

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

export async function listRulesets(): Promise<Ruleset[]> {
	return unwrap(await authFetch(BASE));
}

export async function getRuleset(id: string): Promise<Ruleset> {
	return unwrap(await authFetch(`${BASE}/${id}`));
}

export async function createRuleset(name: string): Promise<Ruleset> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name })
		})
	);
}

// amendRulesetName edits a Ruleset's name by producing a new revision that
// supersedes the current one; it returns the NEW superseding Ruleset.
export async function amendRulesetName(id: string, name: string): Promise<Ruleset> {
	return unwrap(
		await authFetch(`${BASE}/${id}/amend`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name })
		})
	);
}

export async function deleteRuleset(id: string): Promise<void> {
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

// A RuleSection is a single rule block (title + markdown) belonging to a
// ruleset. Sections are exposed via generic CRUD (see main.go).
export interface RuleSection {
	section_id: string;
	parent: string;
	title: string;
	markdown: string;
	owner: string;
}

// A RulesetSection is the join-table record that orders a RuleSection within a
// ruleset. SectionIndex preserves the ordering.
export interface RulesetSection {
	id: string;
	ruleset_id: string;
	section_id: string;
	section_index: number;
}

// RuleAmendmentType values mirror the Go enum in model/rule_amendment.go.
export const RULE_AMENDMENT_ADD = 0;
export const RULE_AMENDMENT_REMOVE = 1;
export const RULE_AMENDMENT_MODIFY = 2;
export const RULE_AMENDMENT_REORDER = 3;

export interface RuleAmendment {
	type: number;
	target_section?: string;
	new_section?: Partial<RuleSection>;
	after?: string;
}

const RULE_SECTION_BASE = '/api/rule_section';
const RULESET_SECTION_BASE = '/api/ruleset_section';

export async function listRuleSections(): Promise<RuleSection[]> {
	return unwrap(await authFetch(RULE_SECTION_BASE));
}

export async function listRulesetSections(): Promise<RulesetSection[]> {
	return unwrap(await authFetch(RULESET_SECTION_BASE));
}

// getRulesetSections returns the sections of a ruleset, in order (as enforced
// by SectionIndex in the RulesetSection join table).
export async function getRulesetSections(rulesetId: string): Promise<RuleSection[]> {
	const [sections, relations] = await Promise.all([listRuleSections(), listRulesetSections()]);
	const ordered = relations
		.filter((r) => r.ruleset_id === rulesetId)
		.sort((a, b) => a.section_index - b.section_index);
	return ordered
		.map((r) => sections.find((s) => s.section_id === r.section_id))
		.filter((s): s is RuleSection => !!s);
}

// amendSections applies a RuleAmendment (add / remove / modify / reorder
// section) to a ruleset and returns the resulting Ruleset. Add/remove/reorder
// may produce a new superseding revision; a pure content edit updates the
// section in place.
export async function amendSections(id: string, amendment: RuleAmendment): Promise<Ruleset> {
	return unwrap(
		await authFetch(`${BASE}/${id}/amend_sections`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(amendment)
		})
	);
}
