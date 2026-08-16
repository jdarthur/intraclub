<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getDraft, getDraftResults } from '$lib/draft';
	import type { Draft, DraftResults } from '$lib/draft';
	import { listUsers, fullName } from '$lib/user';
	import type { User } from '$lib/user';
	import { listRatings } from '$lib/rating';
	import type { Rating } from '$lib/rating';
	import { listSeasons } from '$lib/season';
	import type { Season } from '$lib/season';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import {
		Card,
		CardContent,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card/index.js';

	const id = () => page.params.id as string;

	let draft = $state<Draft | null>(null);
	let results = $state<DraftResults | null>(null);
	let users = $state<User[]>([]);
	let ratingsById = $state<Record<string, Rating>>({});
	let season = $state<Season | null>(null);

	let loading = $state(true);
	let loadError = $state('');

	async function load() {
		loading = true;
		loadError = '';
		try {
			const draftData = await getDraft(id());
			const [resultsData, userList, ratingList, seasonList] = await Promise.all([
				getDraftResults(id()),
				listUsers(),
				listRatings(),
				listSeasons()
			]);
			draft = draftData;
			results = resultsData;
			users = userList;
			ratingsById = Object.fromEntries(ratingList.map((r) => [r.id, r]));
			season = seasonList.find((s) => s.draft_id === draftData.id) ?? null;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load draft results';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function userName(userId: string): string {
		const u = users.find((x) => x.id === userId);
		return u ? fullName(u) : userId;
	}

	function ratingName(ratingId: string): string {
		return ratingsById[ratingId]?.name ?? ratingId;
	}

	// The draft's final teams are returned by the backend in draft order.
	const teams = $derived(results?.teams ?? []);

	// A draft's completion state is read from the draft record's completed_at
	// timestamp (zero value = not yet completed).
	const ZERO_TIME = '0001-01-01T00:00:00Z';
	const isCompleted = $derived(
		draft !== null && draft.completed_at !== '' && draft.completed_at !== ZERO_TIME
	);

	function sortedMembers(members: { user_id: string; rating: string }[]) {
		return [...members].sort((a, b) => userName(a.user_id).localeCompare(userName(b.user_id)));
	}
</script>

<svelte:head>
	<title>Intraclub | {draft ? `${draft.name} — Results` : 'Draft results'}</title>
</svelte:head>

<div class="flex flex-wrap items-center gap-x-4 gap-y-2">
	<h1 class="text-2xl font-semibold tracking-tight">{draft?.name ?? 'Draft results'}</h1>
	<div class="ml-auto flex flex-wrap items-center gap-x-3 gap-y-2">
		<a href={`/drafts/${id()}`} class="text-sm text-muted-foreground hover:text-foreground">
			&larr; Draft setup
		</a>
		<a
			href={`/drafts/${id()}/draft`}
			class="text-sm text-muted-foreground hover:text-foreground"
		>
			Live draft board
		</a>
	</div>
</div>

{#if loading}
	<p class="mt-4 text-muted-foreground">Loading...</p>
{:else if loadError}
	<p class="mt-4 text-sm font-medium text-destructive">{loadError}</p>
{:else if draft}
	<!-- Completion state & created Season -->
	<Card class="mt-6">
		<CardHeader>
			<CardTitle>Draft status</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="flex flex-wrap items-center gap-4 text-sm">
				<Badge variant={isCompleted ? 'secondary' : 'default'}>
					{isCompleted ? 'Completed' : 'In progress'}
				</Badge>
				{#if season}
					<span class="text-muted-foreground">
						Season created:
						<a
							href={`/seasons/${season.id}`}
							class="ml-1 font-medium text-foreground underline underline-offset-2 hover:text-primary"
						>
							{season.name}
						</a>
					</span>
				{:else}
					<span class="text-muted-foreground">No season has been created from this draft yet.</span>
				{/if}
			</div>
		</CardContent>
	</Card>

	{#if teams.length === 0}
		<p class="mt-6 text-sm text-muted-foreground">
			This draft hasn't been initialized yet — go to the setup page to add captains.
		</p>
	{:else}
		<div class="mt-6 grid gap-6 lg:grid-cols-2">
			{#each teams as team}
				<Card>
					<CardHeader>
						<CardTitle>Team {team.draft_order + 1}</CardTitle>
						<p class="text-xs text-muted-foreground">Captain: {userName(team.captain_id)}</p>
					</CardHeader>
					<CardContent>
						{#if team.selections.length === 0}
							<p class="text-sm text-muted-foreground">No players assigned.</p>
						{:else}
							<ul class="mt-2 space-y-2 text-sm">
								{#each sortedMembers(team.selections.map((s) => ({ user_id: s.user.id, rating: s.rating }))) as member}
									<li class="flex items-center justify-between gap-3">
										<span class="font-medium">
											{userName(member.user_id)}
											{#if member.user_id === team.captain_id}
												<span class="text-xs text-muted-foreground">(captain)</span>
											{/if}
										</span>
										<span class="text-xs text-muted-foreground">
											{ratingName(member.rating)}
										</span>
									</li>
								{/each}
							</ul>
						{/if}
					</CardContent>
				</Card>
			{/each}
		</div>
	{/if}
{/if}
