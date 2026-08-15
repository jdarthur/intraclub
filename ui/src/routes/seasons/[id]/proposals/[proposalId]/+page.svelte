<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getSeason } from '$lib/season';
	import type { Season } from '$lib/season';
	import { getProposalDetail, castVote } from '$lib/commissionerProposal';
	import type { ProposalDetail } from '$lib/commissionerProposal';
	import { getCurrentUserId } from '$lib/auth';
	import {
		Card,
		CardContent,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';

	const seasonId = () => page.params.id as string;
	const proposalId = () => page.params.proposalId as string;

	let season = $state<Season | null>(null);
	let detail = $state<ProposalDetail | null>(null);
	let loading = $state(true);
	let loadError = $state('');
	let voting = $state<'for' | 'against' | null>(null);
	let voteError = $state('');

	const currentUserId = $derived(getCurrentUserId() ?? '');
	const isVoter = $derived(detail !== null && detail.voters.includes(currentUserId));

	async function load() {
		loading = true;
		loadError = '';
		try {
			const [seasonData, d] = await Promise.all([
				getSeason(seasonId()),
				getProposalDetail(proposalId())
			]);
			season = seasonData;
			detail = d;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load proposal';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	async function submitVote(vote: boolean) {
		voteError = '';
		voting = vote ? 'for' : 'against';
		try {
			detail = await castVote(proposalId(), vote);
		} catch (e) {
			voteError = e instanceof Error ? e.message : 'Failed to cast vote';
		} finally {
			voting = null;
		}
	}

	function statusLabel(): { text: string; variant: 'default' | 'destructive' | 'secondary' } {
		if (!detail) return { text: 'Pending', variant: 'secondary' };
		if (detail.accepted) return { text: 'Accepted', variant: 'default' };
		if (detail.rejected) return { text: 'Rejected', variant: 'destructive' };
		return { text: 'Pending', variant: 'secondary' };
	}
</script>

<svelte:head>
	<title>Intraclub | Proposal</title>
</svelte:head>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if loadError}
	<p class="text-sm font-medium text-destructive">{loadError}</p>
{:else if detail}
	<div class="flex items-center justify-between gap-4">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Commissioner proposal</h1>
			{#if season}
				<p class="text-sm text-muted-foreground">{season.name}</p>
			{/if}
		</div>
		<Button type="button" variant="outline" href={`/seasons/${seasonId()}/proposals`}>
			Back to proposals
		</Button>
	</div>

	<Card class="mt-6">
		<CardHeader>
			<CardTitle class="text-base">Proposal</CardTitle>
		</CardHeader>
		<CardContent>
			{@const st = statusLabel()}
			<p class="text-sm">{detail.proposal.description}</p>
			<div class="mt-3 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
				<Badge variant={st.variant}>{st.text}</Badge>
				{#if detail.proposal.must_be_unanimous}
					<Badge variant="outline">Unanimous consent</Badge>
				{:else}
					<Badge variant="outline">Majority consent</Badge>
				{/if}
			</div>
		</CardContent>
	</Card>

	<Card class="mt-6">
		<CardHeader>
			<CardTitle class="text-base">Vote tally</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="flex items-center gap-6 text-sm">
				<span class="font-medium text-success">{detail.votes_for} for</span>
				<span class="font-medium text-destructive">{detail.votes_against} against</span>
				<span class="text-muted-foreground">{detail.voters.length} eligible voters</span>
			</div>

			{#if voteError}
				<p class="mt-4 text-sm font-medium text-destructive">{voteError}</p>
			{/if}

			{#if isVoter}
				<div class="mt-4 flex flex-wrap items-center gap-2">
					<span class="text-sm text-muted-foreground">Your vote:</span>
					{#if detail.my_vote === undefined}
						<span class="text-sm">Not cast yet</span>
					{:else}
						<span class="text-sm font-medium">
							{detail.my_vote ? 'For' : 'Against'}
						</span>
					{/if}
					<div class="ml-2 flex items-center gap-2">
						<Button
							type="button"
							size="sm"
							variant={detail.my_vote === true ? 'default' : 'outline'}
							disabled={voting !== null}
							onclick={() => submitVote(true)}
						>
							{voting === 'for' ? 'Voting…' : 'Vote for'}
						</Button>
						<Button
							type="button"
							size="sm"
							variant={detail.my_vote === false ? 'default' : 'outline'}
							disabled={voting !== null}
							onclick={() => submitVote(false)}
						>
							{voting === 'against' ? 'Voting…' : 'Vote against'}
						</Button>
					</div>
				</div>
			{:else}
				<p class="mt-4 text-sm text-muted-foreground">
					Only commissioners and team captains of this season can vote.
				</p>
			{/if}
		</CardContent>
	</Card>
{/if}
