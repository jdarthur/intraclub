<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { getDraft, selectPlayer, assignDraftedPlayersToTeams, createSeason } from '$lib/draft';
	import type { Draft } from '$lib/draft';
	import { listFormats, getFormatPossibleRatings } from '$lib/format';
	import type { Format } from '$lib/format';
	import type { Rating } from '$lib/rating';
	import { listRatings } from '$lib/rating';
	import { listUsers, fullName } from '$lib/user';
	import type { User } from '$lib/user';
	import { getCurrentUserId } from '$lib/auth';
	import { listDraftCaptains } from '$lib/draftCaptain';
	import type { DraftCaptain } from '$lib/draftCaptain';
	import { listDraftAvailablePlayers } from '$lib/draftAvailablePlayer';
	import type { DraftAvailablePlayer } from '$lib/draftAvailablePlayer';
	import { listDraftPicks } from '$lib/draftPick';
	import type { DraftPick } from '$lib/draftPick';
	import { listFacilities } from '$lib/facility';
	import type { Facility } from '$lib/facility';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Card,
		CardContent,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		NativeSelect,
		NativeSelectOption
	} from '$lib/components/ui/native-select/index.js';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	const id = () => page.params.id as string;

	// The zero Go time marshals to this literal; treat it as "not set".
	const ZERO_TIME = '0001-01-01T00:00:00Z';

	// The current user is the one acting; their id comes from the JWT's `sub`.
	const currentUserId = getCurrentUserId();

	let draft = $state<Draft | null>(null);
	let format = $state<Format | null>(null);
	let possibleRatings = $state<Rating[]>([]);
	let ratingsById = $state<Record<string, Rating>>({});
	let users = $state<User[]>([]);
	let captains = $state<DraftCaptain[]>([]);
	let availablePlayers = $state<DraftAvailablePlayer[]>([]);
	let picks = $state<DraftPick[]>([]);
	let facilities = $state<Facility[]>([]);

	let loading = $state(true);
	let loadError = $state('');
	let actionError = $state('');
	let picking = $state(false);

	// Finalize-draft form state.
	let showFinalize = $state(false);
	let finalizeName = $state('');
	let finalizeFacility = $state('');
	let finalizeStartTime = $state('');
	let finalizing = $state(false);
	let finalizeError = $state('');

	async function load() {
		loading = true;
		loadError = '';
		actionError = '';
		try {
			const draftData = await getDraft(id());
			const [formatList, userList, ratingList, captainList, availableList, pickList, facilityList] =
				await Promise.all([
					listFormats(),
					listUsers(),
					listRatings(),
					listDraftCaptains(),
					listDraftAvailablePlayers(),
					listDraftPicks(),
					listFacilities()
				]);

			draft = draftData;
			users = userList;
			ratingsById = Object.fromEntries(ratingList.map((r) => [r.id, r]));
			facilities = facilityList;
			if (!finalizeFacility && facilityList.length > 0) {
				finalizeFacility = facilityList[0].id;
			}

			const fmt = formatList.find((f) => f.id === draftData.format) ?? null;
			format = fmt;
			possibleRatings = fmt ? await getFormatPossibleRatings(fmt.id) : [];

			// Captains are presented in draft order (mirrors the backend's
			// model.Draft.GetCaptains sort by DraftOrder).
			captains = captainList
				.filter((c) => c.draft_id === draftData.id)
				.sort((a, b) => a.draft_order - b.draft_order);
			availablePlayers = availableList.filter((a) => a.draft_id === draftData.id);
			picks = pickList.filter((p) => p.draft_id === draftData.id);
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load draft board';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function userName(userId: string): string {
		const u = users.find((x) => x.id === userId);
		return u ? fullName(u) : userId;
	}

	const numTeams = $derived(captains.length);
	const totalPlayers = $derived(availablePlayers.length);
	const picksCount = $derived(picks.length);
	// A draft is completed once every available player has been selected.
	const isCompleted = $derived(
		totalPlayers > 0 && picksCount >= totalPlayers
	);

	// The next selection index (number of picks already made) maps to a
	// (round, pick) pair; see model.Draft.GetRoundAndPickFromPicks.
	const rounds = $derived(numTeams > 0 ? Math.ceil(totalPlayers / numTeams) : 0);
	const nextRound = $derived(numTeams > 0 ? Math.floor(picksCount / numTeams) + 1 : 1);
	const nextPick = $derived(numTeams > 0 ? (picksCount % numTeams) + 1 : 1);

	// --- Draft order pattern ---
	// Mirrors model.DraftOrderPattern.GetCaptainOnTheClock for each pattern.
	function lastPickDouble(round: number, pick: number, n: number): number {
		if (round === 1) return pick - 1;
		if (pick === 1) return lastPickDouble(round - 1, n, n);
		return captainBefore(round, pick, n, n + 1);
	}
	function captainBefore(
		currentRound: number,
		currentPick: number,
		n: number,
		distance: number
	): number {
		let r = currentRound;
		let p = currentPick;
		for (let i = 0; i < distance; i++) {
			p -= 1;
			if (p === 0) {
				r -= 1;
				p = n;
			}
		}
		return lastPickDouble(r, p, n);
	}
	function onClockCaptainIndex(): number {
		if (numTeams === 0) return -1;
		const n = numTeams;
		const pattern = draft?.draft_order_pattern;
		if (pattern === 'Last pick double') {
			return lastPickDouble(nextRound, nextPick, n);
		}
		if (pattern === 'Straight-up') {
			return nextPick - 1;
		}
		// Snake (default): even rounds draft in reverse order.
		return nextRound % 2 === 0 ? n - nextPick : nextPick - 1;
	}

	const onClockIndex = $derived(onClockCaptainIndex());
	const onClockCaptain = $derived(
		onClockIndex >= 0 && onClockIndex < captains.length
			? captains[onClockIndex].captain_id
			: null
	);
	// A non-captain (e.g. the commissioner viewing) can watch but not pick.
	const canPick = $derived(
		!isCompleted &&
			onClockCaptain !== null &&
			currentUserId !== null &&
			currentUserId === onClockCaptain
	);

	const selectedUserIds = $derived(new Set(picks.map((p) => p.user_id)));
	const captainIds = $derived(new Set(captains.map((c) => c.captain_id)));

	// A player is selectable if they haven't been taken and they are either the
	// captain on the clock or not another captain at all (mirrors
	// model.Draft.GetAllAvailableToSelect).
	function isSelectable(playerId: string): boolean {
		if (selectedUserIds.has(playerId)) return false;
		if (captainIds.has(playerId) && playerId !== onClockCaptain) return false;
		return true;
	}

	function pickStatus(playerId: string): string {
		if (selectedUserIds.has(playerId)) return 'Selected';
		if (captainIds.has(playerId)) return 'Captain';
		return 'Available';
	}

	// --- Board grid (teams × rounds) ---
	const board = $derived(
		captains.map((captain, teamIndex) => {
			const cells: (DraftPick | null)[] = [];
			for (let r = 1; r <= rounds; r++) {
				const pick = picks.find((p) => p.team_id === captain.team_id && p.round === r);
				cells.push(pick ?? null);
			}
			return { teamIndex, captain, cells };
		})
	);

	function ratingName(ratingId: string): string {
		return ratingsById[ratingId]?.name ?? ratingId;
	}

	async function handlePick(playerId: string) {
		actionError = '';
		if (!canPick) return;
		picking = true;
		try {
			await selectPlayer(id(), playerId);
			await load();
		} catch (e) {
			actionError = e instanceof Error ? e.message : 'Failed to make the selection';
		} finally {
			picking = false;
		}
	}

	async function handleFinalize(e: Event) {
		e.preventDefault();
		finalizeError = '';
		if (!isCompleted) return;
		if (!finalizeName.trim()) {
			finalizeError = 'Season name must not be empty.';
			return;
		}
		finalizing = true;
		try {
			// Assign each drafted player to their team with their draft rating,
			// then create the Season from the completed draft.
			await assignDraftedPlayersToTeams(id());
			const season = await createSeason(id(), {
				name: finalizeName.trim(),
				facility: finalizeFacility,
				start_time: finalizeStartTime
			});
			await goto(`/seasons/${season.id}`);
		} catch (err) {
			finalizeError = err instanceof Error ? err.message : 'Failed to finalize the draft';
			finalizing = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | {draft ? `${draft.name} — Live draft` : 'Live draft'}</title>
</svelte:head>

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">{draft?.name ?? 'Live draft'}</h1>
	<div class="ml-auto flex items-center gap-3">
		<a href={`/drafts/${id()}`} class="text-sm text-muted-foreground hover:text-foreground">
			&larr; Draft setup
		</a>
		<a
			href={`/drafts/${id()}/results`}
			class="text-sm text-muted-foreground hover:text-foreground"
		>
			Results →
		</a>
	</div>
</div>

{#if loading}
	<p class="mt-4 text-muted-foreground">Loading...</p>
{:else if loadError}
	<p class="mt-4 text-sm font-medium text-destructive">{loadError}</p>
{:else if draft}
	<!-- Status -->
	<Card class="mt-6">
		<CardHeader>
			<CardTitle class="text-base">Draft status</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="flex flex-wrap items-center gap-4 text-sm">
				<Badge variant={isCompleted ? 'secondary' : 'default'}>
					{isCompleted ? 'Completed' : 'In progress'}
				</Badge>
				<Badge variant="outline">Pattern: {draft.draft_order_pattern || 'Snake'}</Badge>
				{#if !isCompleted}
					<span class="text-muted-foreground">
						Round {nextRound} &middot; Pick {nextPick}
					</span>
					<span class="text-muted-foreground">
						{onClockCaptain ? `${userName(onClockCaptain)} is on the clock` : 'Waiting for captains…'}
					</span>
				{:else}
					<span class="text-muted-foreground">
						All {totalPlayers} players drafted.
					</span>
				{/if}
			</div>
			{#if canPick}
				<p class="mt-3 text-sm text-success">
					You're on the clock — select a player below.
				</p>
			{:else if !isCompleted && onClockCaptain}
				<p class="mt-3 text-sm text-muted-foreground">
					Only {userName(onClockCaptain)} can pick right now.
				</p>
			{/if}
		</CardContent>
	</Card>

	{#if actionError}
		<p class="mt-4 text-sm font-medium text-destructive">{actionError}</p>
	{/if}

	<!-- Finalize flow: only available once every available player has been
	     picked. Assigns drafted players to their teams, then creates a Season. -->
	{#if isCompleted}
		<Card class="mt-6">
			<CardHeader>
				<CardTitle class="text-base">Finalize draft</CardTitle>
			</CardHeader>
			<CardContent>
				{#if !showFinalize}
					<p class="text-sm text-muted-foreground">
						All {totalPlayers} players have been drafted. Finalizing assigns the drafted
						rosters to their teams and creates a new Season.
					</p>
					<Button
						type="button"
						class="mt-4"
						onclick={() => {
							finalizeError = '';
							showFinalize = true;
						}}
					>
						Finalize draft
					</Button>
				{:else}
					<form onsubmit={handleFinalize} class="mt-2 space-y-4">
						<div class="grid gap-4 sm:grid-cols-2">
							<div>
								<Label for="season-name">Season name</Label>
								<Input
									id="season-name"
									class="mt-1.5"
									bind:value={finalizeName}
									placeholder="e.g. Men's Intraclub 2025"
									required
								/>
							</div>
							<div>
								<Label for="season-start-time">Start time (HH:MM)</Label>
								<Input
									id="season-start-time"
									class="mt-1.5"
									bind:value={finalizeStartTime}
									placeholder="e.g. 08:30"
									required
								/>
							</div>
						</div>
						<div>
							<Label for="season-facility">Facility</Label>
							<NativeSelect
								id="season-facility"
								class="mt-1.5"
								bind:value={finalizeFacility}
							>
								{#each facilities as facility}
									<NativeSelectOption value={facility.id}>{facility.name}</NativeSelectOption>
								{/each}
							</NativeSelect>
						</div>
						{#if finalizeError}
							<p class="text-sm font-medium text-destructive">{finalizeError}</p>
						{/if}
						<div class="flex items-center gap-3">
							<Button type="submit" disabled={finalizing}>
								{finalizing ? 'Finalizing…' : 'Create season'}
							</Button>
							<Button
								type="button"
								variant="outline"
								disabled={finalizing}
								onclick={() => {
									showFinalize = false;
									finalizeError = '';
								}}
							>
								Cancel
							</Button>
						</div>
					</form>
				{/if}
			</CardContent>
		</Card>
	{/if}

	<div class="mt-6 grid gap-6 lg:grid-cols-2">
		<!-- Draft board grid -->
		<Card>
			<CardHeader>
				<CardTitle class="text-base">Draft board</CardTitle>
			</CardHeader>
			<CardContent>
				{#if numTeams === 0}
					<p class="text-sm text-muted-foreground">
						This draft hasn't been initialized yet — go to the setup page to add captains.
					</p>
				{:else if rounds === 0}
					<p class="text-sm text-muted-foreground">
						No draftable players yet.
					</p>
				{:else}
					<div class="mt-2 overflow-x-auto rounded-lg border">
						<table class="draft-board w-full text-sm">
							<thead>
								<tr class="border-b bg-muted/50 text-left">
									<th class="px-3 py-2">Team</th>
									{#each Array(rounds) as _, r}
										<th class="px-3 py-2 text-center">Round {r + 1}</th>
									{/each}
								</tr>
							</thead>
							<tbody>
								{#each board as row}
									<tr class="border-b last:border-0">
										<td class="whitespace-nowrap px-3 py-2">
											<div class="font-medium">Team {row.teamIndex + 1}</div>
											<div class="text-xs text-muted-foreground">
												{userName(row.captain.captain_id)}
											</div>
										</td>
										{#each row.cells as pick}
											<td class="px-3 py-2 text-center align-top">
												{#if pick}
													<div class="text-xs">{userName(pick.user_id)}</div>
													<div class="text-xs text-muted-foreground">
														{ratingName(pick.rating)}
													</div>
												{:else}
													<span class="text-muted-foreground">—</span>
												{/if}
											</td>
										{/each}
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{/if}
			</CardContent>
		</Card>

		<!-- Pick area -->
		<Card>
			<CardHeader>
				<CardTitle class="text-base">Make a pick</CardTitle>
			</CardHeader>
			<CardContent>
				{#if availablePlayers.length === 0}
					<p class="text-sm text-muted-foreground">No draftable players yet.</p>
				{:else}
					<div class="mt-2 overflow-hidden rounded-lg border">
						<table class="w-full text-sm">
							<thead>
								<tr class="border-b bg-muted/50 text-left">
									<th class="px-3 py-2">Player</th>
									<th class="px-3 py-2">Status</th>
									<th class="px-3 py-2"></th>
								</tr>
							</thead>
							<tbody>
								{#each availablePlayers as player}
									{@const selectable = isSelectable(player.player_id)}
									<tr class="border-b last:border-0">
										<td class="px-3 py-2">{userName(player.player_id)}</td>
										<td class="px-3 py-2 text-xs text-muted-foreground">
											{pickStatus(player.player_id)}
										</td>
										<td class="px-3 py-2 text-right">
											<Button
												type="button"
												size="sm"
												variant="default"
												disabled={!canPick || !selectable || picking}
												onclick={() => handlePick(player.player_id)}
											>
												Select
											</Button>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
					{#if !canPick && !isCompleted}
						<p class="mt-3 text-xs text-muted-foreground">
							Picks are locked until it's your turn.
						</p>
					{/if}
				{/if}
			</CardContent>
		</Card>
	</div>
{/if}
