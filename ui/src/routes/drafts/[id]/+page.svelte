<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import {
		getDraft,
		initializeDraft,
		assignDraftablePlayers,
		assignRatingCutoff
	} from '$lib/draft';
	import type { Draft } from '$lib/draft';
	import { listFormats } from '$lib/format';
	import type { Format } from '$lib/format';
	import { getFormatPossibleRatings } from '$lib/format';
	import { listRatings } from '$lib/rating';
	import type { Rating } from '$lib/rating';
	import { listUsers, fullName } from '$lib/user';
	import type { User } from '$lib/user';
	import { listDraftCaptains } from '$lib/draftCaptain';
	import type { DraftCaptain } from '$lib/draftCaptain';
	import {
		listDraftAvailablePlayers,
		deleteDraftAvailablePlayer
	} from '$lib/draftAvailablePlayer';
	import type { DraftAvailablePlayer } from '$lib/draftAvailablePlayer';
	import {
		listDraftRatingCutoffs,
		updateDraftRatingCutoff
	} from '$lib/draftRatingCutoff';
	import type { DraftRatingCutoff } from '$lib/draftRatingCutoff';
	import { listDraftPicks } from '$lib/draftPick';
	import type { DraftPick } from '$lib/draftPick';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Card,
		CardContent,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		NativeSelect,
		NativeSelectOption
	} from '$lib/components/ui/native-select/index.js';

	const id = () => page.params.id as string;

	let draft = $state<Draft | null>(null);
	let format = $state<Format | null>(null);
	let possibleRatings = $state<Rating[]>([]);
	let ratingsById = $state<Record<string, Rating>>({});
	let users = $state<User[]>([]);

	let captains = $state<DraftCaptain[]>([]);
	let availablePlayers = $state<DraftAvailablePlayer[]>([]);
	let ratingCutoffs = $state<DraftRatingCutoff[]>([]);
	let picks = $state<DraftPick[]>([]);

	let loading = $state(true);
	let loadError = $state('');

	// Captains / teams configuration (pre-initialization).
	let selectedCaptains = $state<string[]>([]);
	let captainToAdd = $state('');
	let initializing = $state(false);
	let actionError = $state('');

	// Available players configuration.
	let playerToAdd = $state('');
	let addingPlayer = $state(false);

	// Rating cutoffs configuration. Kept as an array of entry objects so each
	// input can bind to a stable item property (binding to a dynamic object key
	// does not propagate in Svelte 5).
	let cutoffEntries = $state<{ ratingId: string; value: string }[]>([]);
	let savingCutoffs = $state(false);

	async function load() {
		loading = true;
		loadError = '';
		actionError = '';
		try {
			const draftData = await getDraft(id());
			const [
				formatList,
				userList,
				ratingList,
				captainList,
				availableList,
				cutoffList,
				pickList
			] = await Promise.all([
				listFormats(),
				listUsers(),
				listRatings(),
				listDraftCaptains(),
				listDraftAvailablePlayers(),
				listDraftRatingCutoffs(),
				listDraftPicks()
			]);

			draft = draftData;
			users = userList;
			ratingsById = Object.fromEntries(ratingList.map((r) => [r.id, r]));

			const fmt = formatList.find((f) => f.id === draftData.format) ?? null;
			format = fmt;
			possibleRatings = fmt ? await getFormatPossibleRatings(fmt.id) : [];

			captains = captainList.filter((c) => c.draft_id === draftData.id);
			availablePlayers = availableList.filter((a) => a.draft_id === draftData.id);
			ratingCutoffs = cutoffList.filter((c) => c.draft_id === draftData.id);
			picks = pickList.filter((p) => p.draft_id === draftData.id);

			// Seed the cutoff inputs from the existing rows.
			cutoffEntries = ratingsToConfigure().map((rating) => {
				const row = ratingCutoffs.find((c) => c.rating_id === rating.id);
				return { ratingId: rating.id, value: row ? String(row.cutoff_index) : '' };
			});
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load draft';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function userName(userId: string): string {
		const u = users.find((x) => x.id === userId);
		return u ? fullName(u) : userId;
	}

	// Ratings to configure = all of the format's possible ratings except the
	// lowest-skill one (which never receives a cutoff).
	function ratingsToConfigure(): Rating[] {
		return possibleRatings.length > 0 ? possibleRatings.slice(0, -1) : [];
	}

	// --- Captains / teams ---

	function addCaptain() {
		if (captainToAdd && !selectedCaptains.includes(captainToAdd)) {
			selectedCaptains = [...selectedCaptains, captainToAdd];
		}
		captainToAdd = '';
	}

	function removeCaptain(userId: string) {
		selectedCaptains = selectedCaptains.filter((c) => c !== userId);
	}

	async function handleInitialize() {
		actionError = '';
		if (selectedCaptains.length === 0) {
			actionError = 'Select at least one captain to initialize the draft.';
			return;
		}
		initializing = true;
		try {
			await initializeDraft(id(), selectedCaptains);
			selectedCaptains = [];
			await load();
		} catch (e) {
			actionError = e instanceof Error ? e.message : 'Failed to initialize draft';
		} finally {
			initializing = false;
		}
	}

	// --- Available players ---

	async function handleAddPlayer() {
		actionError = '';
		if (!playerToAdd) return;
		addingPlayer = true;
		try {
			await assignDraftablePlayers(id(), [playerToAdd]);
			playerToAdd = '';
			await load();
		} catch (e) {
			actionError = e instanceof Error ? e.message : 'Failed to add player';
		} finally {
			addingPlayer = false;
		}
	}

	async function handleRemovePlayer(row: DraftAvailablePlayer) {
		actionError = '';
		try {
			await deleteDraftAvailablePlayer(row.id);
			await load();
		} catch (e) {
			actionError = e instanceof Error ? e.message : 'Failed to remove player';
		}
	}

	// --- Rating cutoffs ---

	// Returns an error string when the current cutoff inputs are not a valid
	// (strictly increasing) set for the format's ratings, otherwise null.
	function cutoffError(): string | null {
		let previous = -1;
		for (const entry of cutoffEntries) {
			const name = ratingsById[entry.ratingId]?.name ?? entry.ratingId;
			const raw = entry.value.trim();
			if (raw === '') return `Enter a cutoff for ${name}.`;
			const value = Number(raw);
			if (!Number.isInteger(value) || value <= 0) {
				return `Cutoff for ${name} must be a positive integer.`;
			}
			if (value <= previous) {
				return `Cutoff for ${name} must be greater than ${previous}.`;
			}
			previous = value;
		}
		return null;
	}

	async function handleSaveCutoffs() {
		actionError = '';
		const error = cutoffError();
		if (error) {
			actionError = error;
			return;
		}
		savingCutoffs = true;
		try {
			for (const entry of cutoffEntries) {
				const value = Number(entry.value.trim());
				const existing = ratingCutoffs.find((c) => c.rating_id === entry.ratingId);
				if (existing) {
					await updateDraftRatingCutoff(existing.id, {
						draft_id: draft!.id,
						rating_id: entry.ratingId,
						cutoff_index: value
					});
				} else {
					await assignRatingCutoff(draft!.id, entry.ratingId, value);
				}
			}
			await load();
		} catch (e) {
			actionError = e instanceof Error ? e.message : 'Failed to save rating cutoffs';
		} finally {
			savingCutoffs = false;
		}
	}

	// --- Readiness indicators ---

	const teams = $derived(captains.length);
	const playerCount = $derived(availablePlayers.length);
	const pickCount = $derived(picks.length);
	const cutoffsReady = $derived(
		ratingsToConfigure().length > 0 && cutoffError() === null
	);
	const readyToStart = $derived(
		teams > 0 && playerCount > 0 && cutoffsReady
	);

	function playersPerTeamText(): string {
		if (teams === 0) return 'No teams yet.';
		const perTeam = Math.floor(playerCount / teams);
		return `${playerCount} available players across ${teams} teams (${perTeam} per team)`;
	}

	function userOptions(exclude: string[]): User[] {
		return users.filter((u) => !exclude.includes(u.id));
	}
</script>

<svelte:head>
	<title>{draft ? draft.name : 'Draft'}</title>
</svelte:head>

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">{draft?.name ?? 'Draft'}</h1>
	<div class="ml-auto flex items-center gap-3">
		<a href={`/drafts/${id()}/draft`} class="text-sm text-muted-foreground hover:text-foreground">
			Live draft board →
		</a>
		<a href={`/drafts/${id()}/grades`} class="text-sm text-muted-foreground hover:text-foreground">
			Pre-draft grades →
		</a>
		<a href="/drafts" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to drafts</a>
	</div>
</div>

{#if loading}
	<p class="mt-4 text-muted-foreground">Loading...</p>
{:else if loadError}
	<p class="mt-4 text-sm font-medium text-destructive">{loadError}</p>
{:else if draft}
	<!-- State summary / readiness -->
	<Card class="mt-6">
		<CardHeader>
			<CardTitle class="text-base">Setup status</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="flex flex-wrap items-center gap-4 text-sm">
				<div class="flex items-center gap-2">
					<Badge variant={teams > 0 ? 'default' : 'secondary'}>
						{teams} team{teams === 1 ? '' : 's'}
					</Badge>
					<span class="text-muted-foreground">captains</span>
				</div>
				<div class="flex items-center gap-2">
					<Badge variant={playerCount > 0 ? 'default' : 'secondary'}>
						{playerCount} player{playerCount === 1 ? '' : 's'}
					</Badge>
					<span class="text-muted-foreground">{playersPerTeamText()}</span>
				</div>
				<div class="flex items-center gap-2">
					<Badge variant="secondary">{pickCount} pick{pickCount === 1 ? '' : 's'} made</Badge>
				</div>
				<div class="ml-auto">
					<Badge variant={readyToStart ? 'default' : 'secondary'}>
						{readyToStart ? 'Ready to grade & pick' : 'Not ready to start'}
					</Badge>
				</div>
			</div>
		</CardContent>
	</Card>

	{#if actionError}
		<p class="mt-4 text-sm font-medium text-destructive">{actionError}</p>
	{/if}

	<div class="mt-6 grid gap-6 lg:grid-cols-2">
		<!-- Captains / teams -->
		<Card>
			<CardHeader>
				<CardTitle class="text-base">Captains &amp; teams</CardTitle>
			</CardHeader>
			<CardContent>
				{#if captains.length === 0}
					<p class="text-sm text-muted-foreground">
						Pick the captains who will run the draft. Initializing creates the teams
						and the draft order.
					</p>
					<div class="mt-4 flex flex-col gap-2">
						<Label for="captain-select">Captain</Label>
						<div class="flex gap-2">
							<NativeSelect id="captain-select" bind:value={captainToAdd} class="w-full">
								<NativeSelectOption value="" disabled>Select a captain…</NativeSelectOption>
								{#each userOptions(selectedCaptains) as u}
									<NativeSelectOption value={u.id}>{fullName(u)}</NativeSelectOption>
								{/each}
							</NativeSelect>
							<Button type="button" variant="outline" onclick={addCaptain}>Add captain</Button>
						</div>
					</div>
					{#if selectedCaptains.length > 0}
						<ul class="mt-4 flex flex-col gap-2">
							{#each selectedCaptains as captainId}
								<li class="flex items-center justify-between rounded-lg border px-3 py-2 text-sm">
									<span>{userName(captainId)}</span>
									<Button type="button" variant="ghost" onclick={() => removeCaptain(captainId)}>
										Remove
									</Button>
								</li>
							{/each}
						</ul>
					{/if}
					<Button
						class="mt-4"
						onclick={handleInitialize}
						disabled={initializing || selectedCaptains.length === 0}
					>
						{initializing ? 'Initializing...' : 'Initialize draft'}
					</Button>
				{:else}
					<div class="overflow-hidden rounded-lg border">
						<table class="w-full text-sm">
							<thead>
								<tr class="border-b bg-muted/50 text-left">
									<th class="px-3 py-2">Order</th>
									<th class="px-3 py-2">Team</th>
									<th class="px-3 py-2">Captain</th>
									<th class="px-3 py-2"></th>
								</tr>
							</thead>
							<tbody>
								{#each captains as captain, i}
									<tr class="border-b last:border-0">
										<td class="px-3 py-2">{i + 1}</td>
										<td class="px-3 py-2">Team {i + 1}</td>
										<td class="px-3 py-2">{userName(captain.captain_id)}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
					<p class="mt-3 text-xs text-muted-foreground">
						Captains are assigned in draft order and are fixed once the draft is
						initialized.
					</p>
				{/if}
			</CardContent>
		</Card>

		<!-- Available players -->
		<Card>
			<CardHeader>
				<CardTitle class="text-base">Available players</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="text-sm text-muted-foreground">
					Add the players who are eligible to be drafted. Keep the count even-ish across
					teams so each team ends up with a fair roster.
				</p>
				<div class="mt-4 flex flex-col gap-2">
					<Label for="player-select">Player</Label>
					<div class="flex gap-2">
						<NativeSelect id="player-select" bind:value={playerToAdd} class="w-full">
							<NativeSelectOption value="" disabled>Select a player…</NativeSelectOption>
							{#each userOptions(availablePlayers.map((a) => a.player_id)) as u}
								<NativeSelectOption value={u.id}>{fullName(u)}</NativeSelectOption>
							{/each}
						</NativeSelect>
						<Button type="button" variant="outline" onclick={handleAddPlayer} disabled={addingPlayer}>
							Add player
						</Button>
					</div>
				</div>
				{#if availablePlayers.length === 0}
					<p class="mt-4 text-sm text-muted-foreground">No available players yet.</p>
				{:else}
					<ul class="mt-4 flex flex-col gap-2">
						{#each availablePlayers as player}
							<li class="flex items-center justify-between rounded-lg border px-3 py-2 text-sm">
								<span>{userName(player.player_id)}</span>
								<Button
									type="button"
									variant="ghost"
									onclick={() => handleRemovePlayer(player)}
								>
									Remove
								</Button>
							</li>
						{/each}
					</ul>
				{/if}
			</CardContent>
		</Card>

		<!-- Rating cutoffs -->
		<Card>
			<CardHeader>
				<CardTitle class="text-base">Rating cutoffs</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="text-sm text-muted-foreground">
					Assign a cutoff (the last pick index) for each rating except the lowest. Cutoffs
					must increase as ratings get lower, e.g. the strongest rating gets the earliest
					picks.
				</p>
				{#if possibleRatings.length === 0}
					<p class="mt-4 text-sm text-muted-foreground">
						This draft's format has no ratings to configure.
					</p>
				{:else}
					<form
						class="mt-4 flex flex-col gap-3"
						onsubmit={(e) => {
							e.preventDefault();
							handleSaveCutoffs();
						}}
					>
						{#each cutoffEntries as entry}
							<div class="flex items-center justify-between gap-4">
								<Label for={`cutoff-${entry.ratingId}`} class="whitespace-nowrap">
									{ratingsById[entry.ratingId]?.name ?? entry.ratingId}
								</Label>
								<input
									id={`cutoff-${entry.ratingId}`}
									type="number"
									min="1"
									class="h-8 w-28 rounded-lg border border-input bg-transparent px-2.5 py-1 text-sm outline-none transition-colors placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
									value={entry.value}
									oninput={(e) => {
										entry.value = (e.currentTarget as HTMLInputElement).value;
									}}
									placeholder="cutoff"
								/>
							</div>
						{/each}
						<div class="mt-2">
							<Button type="submit" disabled={savingCutoffs}>
								{savingCutoffs ? 'Saving...' : 'Save cutoffs'}
							</Button>
						</div>
					</form>
				{/if}
			</CardContent>
		</Card>
	</div>
{/if}
