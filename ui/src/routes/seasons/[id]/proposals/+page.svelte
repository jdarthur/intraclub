<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getSeason } from '$lib/season';
	import type { Season } from '$lib/season';
	import { getScheduleForSeason } from '$lib/schedule';
	import type { ScheduleDetail } from '$lib/schedule';
	import {
		listProposals,
		createProposal,
		getProposalDetail
	} from '$lib/commissionerProposal';
	import type { CommissionerProposal, ProposalDetail } from '$lib/commissionerProposal';
	import { getCurrentUserId } from '$lib/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Card,
		CardContent,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';

	const id = () => page.params.id as string;

	let season = $state<Season | null>(null);
	let schedule = $state<ScheduleDetail | null>(null);
	let details = $state<Record<string, ProposalDetail>>({});
	let loading = $state(true);
	let loadError = $state('');

	// create-proposal form state
	let showCreate = $state(false);
	let description = $state('');
	let mustBeUnanimous = $state(false);
	let creating = $state(false);
	let createError = $state('');

	const isCommissioner = $derived(
		schedule !== null && schedule.commissioners.includes(getCurrentUserId() ?? '')
	);

	async function load() {
		loading = true;
		loadError = '';
		try {
			const [seasonData, scheduleData, proposals] = await Promise.all([
				getSeason(id()),
				getScheduleForSeason(id()),
				listProposals()
			]);
			season = seasonData;
			schedule = scheduleData;
			const seasonProposals = proposals.filter((p) => p.season_id === seasonData.id);
			const detailMap: Record<string, ProposalDetail> = {};
			for (const p of seasonProposals) {
				try {
					detailMap[p.id] = await getProposalDetail(p.id);
				} catch {
					// an eligible user can always fetch detail; skip on transient error
				}
			}
			details = detailMap;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load proposals';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	async function submitCreate() {
		createError = '';
		creating = true;
		try {
			await createProposal({
				description,
				season_id: id(),
				must_be_unanimous: mustBeUnanimous
			});
			description = '';
			mustBeUnanimous = false;
			showCreate = false;
			await load();
		} catch (e) {
			createError = e instanceof Error ? e.message : 'Failed to create proposal';
		} finally {
			creating = false;
		}
	}

	function statusLabel(d: ProposalDetail): { text: string; variant: 'default' | 'destructive' | 'secondary' } {
		if (d.accepted) return { text: 'Accepted', variant: 'default' };
		if (d.rejected) return { text: 'Rejected', variant: 'destructive' };
		return { text: 'Pending', variant: 'secondary' };
	}
</script>

<svelte:head>
	<title>Intraclub | {season?.name ?? 'Season'} proposals</title>
</svelte:head>

<div class="flex items-center justify-between gap-4">
	<div>
		<h1 class="text-2xl font-semibold tracking-tight">Commissioner proposals</h1>
		{#if season}
			<p class="text-sm text-muted-foreground">{season.name}</p>
		{/if}
	</div>
	{#if isCommissioner}
		<Button type="button" onclick={() => (showCreate = !showCreate)}>
			{showCreate ? 'Cancel' : 'New proposal'}
		</Button>
	{/if}
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if loadError}
	<p class="text-sm font-medium text-destructive">{loadError}</p>
{:else}
	{#if showCreate}
		<Card class="mt-6">
			<CardHeader>
				<CardTitle class="text-base">New proposal</CardTitle>
			</CardHeader>
			<CardContent>
				{#if createError}
					<p class="mb-4 text-sm font-medium text-destructive">{createError}</p>
				{/if}
				<div class="flex flex-col gap-4">
					<div class="flex flex-col gap-1">
						<Label for="proposal-description">Description</Label>
						<Textarea
							id="proposal-description"
							placeholder="Describe the rule change or administrative action…"
							bind:value={description}
							required
						/>
					</div>
					<div class="flex items-center gap-2">
						<Checkbox checked={mustBeUnanimous} onCheckedChange={(c) => (mustBeUnanimous = !!c)} />
						<Label class="text-sm">Require unanimous consent</Label>
					</div>
					<div>
						<Button
							type="button"
							disabled={creating || description.trim().length === 0}
							onclick={submitCreate}
						>
							{creating ? 'Creating…' : 'Create proposal'}
						</Button>
					</div>
				</div>
			</CardContent>
		</Card>
	{/if}

	{#if Object.keys(details).length === 0}
		<p class="mt-6 text-muted-foreground">No proposals for this season yet.</p>
	{:else}
		<div class="mt-6 flex flex-col gap-4">
			{#each Object.values(details) as detail (detail.proposal.id)}
				{@const st = statusLabel(detail)}
				<Card>
					<CardContent class="pt-6">
						<div class="flex items-start justify-between gap-4">
							<div class="flex-1">
								<p class="font-medium">{detail.proposal.description}</p>
								<div class="mt-1 flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
									<Badge variant={st.variant}>{st.text}</Badge>
									<span>
										{detail.votes_for} for · {detail.votes_against} against
									</span>
									{#if detail.proposal.must_be_unanimous}
										<span>· unanimous</span>
									{:else}
										<span>· majority</span>
									{/if}
								</div>
							</div>
							<Button
								type="button"
								variant="outline"
								size="sm"
								href={`/seasons/${id()}/proposals/${detail.proposal.id}`}
							>
								View
							</Button>
						</div>
					</CardContent>
				</Card>
			{/each}
		</div>
	{/if}
{/if}
