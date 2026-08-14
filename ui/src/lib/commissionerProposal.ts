// Commissioner proposal API client. Backed by the CRUD routes registered in
// main.go plus the custom detail / vote endpoints in route/proposal:
//
//	GET    /api/commissioner_proposal        -> { resource: CommissionerProposal[] }
//	POST   /api/commissioner_proposal        -> { resource: CommissionerProposal }
//	GET    /api/commissioner_proposal/:id    -> { resource: CommissionerProposal }
//	PUT    /api/commissioner_proposal/:id    -> { resource: CommissionerProposal }
//	DELETE /api/commissioner_proposal/:id    -> { resource: CommissionerProposal }
//	GET    /api/commissioner_proposal/:id/detail -> { resource: ProposalDetail }
//	POST   /api/commissioner_proposal/:id/vote   -> { resource: ProposalDetail }
//
// Proposals are season-scoped; a commissioner creates them and the season's
// participants (commissioners + team captains) ratify them. Record IDs are
// 16-char hex strings.
import { authFetch } from '$lib/auth';

export interface CommissionerProposal {
	id: string;
	description: string;
	season_id: string;
	must_be_unanimous: boolean;
}

export interface ProposalDetail {
	proposal: CommissionerProposal;
	votes_for: number;
	votes_against: number;
	voters: string[];
	accepted: boolean;
	rejected: boolean;
	my_vote?: boolean;
}

export interface ProposalInput {
	description: string;
	season_id: string;
	must_be_unanimous: boolean;
}

const BASE = '/api/commissioner_proposal';

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

export async function listProposals(): Promise<CommissionerProposal[]> {
	return unwrap(await authFetch(BASE));
}

export async function createProposal(input: ProposalInput): Promise<CommissionerProposal> {
	return unwrap(
		await authFetch(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		})
	);
}

export async function getProposalDetail(id: string): Promise<ProposalDetail> {
	return unwrap(await authFetch(`${BASE}/${id}/detail`));
}

export async function castVote(id: string, vote: boolean): Promise<ProposalDetail> {
	return unwrap(
		await authFetch(`${BASE}/${id}/vote`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ vote })
		})
	);
}
