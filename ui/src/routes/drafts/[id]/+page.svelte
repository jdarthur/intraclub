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
	import { listOrganizations, listMembers } from '$lib/organization';
	import type { Organization } from '$lib/organization';
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
	import {
		Tabs,
		TabsContent,
		TabsList,
		TabsTrigger
	} from '$lib/components/ui/tabs/index.js';
	import * as TransferList from '$lib/components/ui/transfer-list/index.js';

	const id = () => page.params.id as string;

	let draft = $state<Draft | null>(null);
	let format = $state<Format | null>(null);
	let possibleRatings = $state<Rating[]>([]);
	let ratingsById = $state<Record<string, Rating>>({});
	let users = $state<User[]>([]);

	// Organizations and their membership roster, used to filter the "Available
	// players" source pool (transient UI filter only — nothing persisted).
	let organizations = $state<Organization[]>([]);
	let userIdsByOrg = $state<Map<string, Set<string>>>(new Map());
	let selectedOrgIds = $state<string[]>([]);

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

	// Available players configuration (transfer-list PoC). The core holds the
	// in-memory source/target lists and selection; we persist diffs on save.
	let transferCore = $state<TransferList.Core<User> | null>(null);
	let savingPlayers = $state(false);

	// Rating cutoffs configuration. Kept as an array of entry objects so each
	// input can bind to a stable item property (binding to a dynamic object key
	// does not propagate in Svelte 5).
	let cutoffEntries = $state<{ ratingId: string; value: string }[]>([]);
	let savingCutoffs = $state(false);

	// Which setup section is shown in the sidebar at a time.
	let activeTab = $state('captains');

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
				pickList,
				organizationList
			] = await Promise.all([
				listFormats(),
				listUsers(),
				listRatings(),
				listDraftCaptains(),
				listDraftAvailablePlayers(),
				listDraftRatingCutoffs(),
				listDraftPicks(),
				listOrganizations()
			]);

			draft = draftData;
			users = userList;
			ratingsById = Object.fromEntries(ratingList.map((r) => [r.id, r]));

			// Build the org -> member userIds map (N+1 per org is acceptable for
			// the small number of organizations).
			organizations = organizationList;
			userIdsByOrg = new Map(
				await Promise.all(
					organizationList.map(async (org) => {
						const members = await listMembers(org.id);
						return [org.id, new Set(members.map((m) => m.id))] as [
							string,
							Set<string>
						];
					})
				)
			);
			selectedOrgIds = [];

			const fmt = formatList.find((f) => f.id === draftData.format) ?? null;
			format = fmt;
			possibleRatings = fmt ? await getFormatPossibleRatings(fmt.id) : [];

			captains = captainList.filter((c) => c.draft_id === draftData.id);
			availablePlayers = availableList.filter((a) => a.draft_id === draftData.id);
			ratingCutoffs = cutoffList.filter((c) => c.draft_id === draftData.id);
			picks = pickList.filter((p) => p.draft_id === draftData.id);

			// Seed the transfer list: source = every user not already in the
			// pool, target = the currently available players.
			const inPool = new Set(availablePlayers.map((a) => a.player_id));
			transferCore = new TransferList.Core<User>({
				initialSource: users.filter((u) => !inPool.has(u.id)),
				initialTarget: availablePlayers
					.map((a) => users.find((u) => u.id === a.player_id))
					.filter((u): u is User => u !== undefined),
				filterPredicate: (u, search) =>
					fullName(u).toLowerCase().includes(search.toLowerCase())
			});

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

	// Union of member userIds across the currently selected organizations.
	// Empty selection -> empty set (means "no filter", handled by callers).
	function memberIdsOfSelectedOrgs(): Set<string> {
		const ids = new Set<string>();
		for (const orgId of selectedOrgIds) {
			for (const uid of userIdsByOrg.get(orgId) ?? []) {
				ids.add(uid);
			}
		}
		return ids;
	}

	function isOrgSelected(orgId: string): boolean {
		return selectedOrgIds.includes(orgId);
	}

	function toggleOrg(orgId: string) {
		selectedOrgIds = isOrgSelected(orgId)
			? selectedOrgIds.filter((id) => id !== orgId)
			: [...selectedOrgIds, orgId];
		recomputeSource();
	}

	// Recomputes the transfer core's source list from the org filter. This is
	// what enforces the hard pool boundary: with orgs selected, ONLY members of
	// those orgs can ever appear in the source (and therefore be moved).
	// `target` (already-available players) is left untouched so in-progress
	// additions aren't clobbered by a filter change.
	function recomputeSource() {
		if (!transferCore) return;
		const inPool = new Set(availablePlayers.map((a) => a.player_id));
		const targetIds = new Set(transferCore.target.map((u) => u.id));
		const base = users.filter((u) => !inPool.has(u.id) && !targetIds.has(u.id));
		if (selectedOrgIds.length === 0) {
			transferCore.source = base;
		} else {
			const memberIds = memberIdsOfSelectedOrgs();
			transferCore.source = base.filter((u) => memberIds.has(u.id));
		}
		// Drop stale selections referencing users no longer in the source.
		transferCore.selectedSourceItems.clear();
		transferCore.sourceSearchQuery = '';
	}

	const filteredMemberCount = $derived(memberIdsOfSelectedOrgs().size);

	// Persists the transfer-list state: adds players that moved into the
	// target, removes players that moved back to the source.
	async function handleSavePlayers() {
		actionError = '';
		if (!transferCore) return;
		const poolIds = new Set(availablePlayers.map((a) => a.player_id));
		const targetIds = new Set(transferCore.target.map((u) => u.id));
		const toAdd = transferCore.target
			.filter((u) => !poolIds.has(u.id))
			.map((u) => u.id);
		const toRemove = availablePlayers.filter((a) => !targetIds.has(a.player_id));
		if (toAdd.length === 0 && toRemove.length === 0) return;
		savingPlayers = true;
		try {
			if (toAdd.length > 0) {
				await assignDraftablePlayers(id(), toAdd);
			}
			for (const row of toRemove) {
				await deleteDraftAvailablePlayer(row.id);
			}
			await load();
		} catch (e) {
			actionError = e instanceof Error ? e.message : 'Failed to save players';
		} finally {
			savingPlayers = false;
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
	<title>Intraclub | {draft ? draft.name : 'Draft'}</title>
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
		<a href={`/drafts/${id()}/results`} class="text-sm text-muted-foreground hover:text-foreground">
			Results →
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

	<Tabs bind:value={activeTab} orientation="vertical" class="mt-6 flex gap-6">
		<TabsList class="h-fit w-56 shrink-0 items-stretch">
			<TabsTrigger
				value="captains"
				class="group justify-start gap-2.5 px-3 py-2 text-base data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:font-semibold data-[state=active]:shadow-sm dark:data-[state=active]:bg-input/30"
			>
				<span
					class="size-1.5 shrink-0 rounded-full bg-primary opacity-0 transition-opacity group-data-[state=active]:opacity-100"
					aria-hidden="true"
				></span>
				Captains &amp; teams
			</TabsTrigger>
			<TabsTrigger
				value="players"
				class="group justify-start gap-2.5 px-3 py-2 text-base data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:font-semibold data-[state=active]:shadow-sm dark:data-[state=active]:bg-input/30"
			>
				<span
					class="size-1.5 shrink-0 rounded-full bg-primary opacity-0 transition-opacity group-data-[state=active]:opacity-100"
					aria-hidden="true"
				></span>
				Available players
			</TabsTrigger>
			<TabsTrigger
				value="cutoffs"
				class="group justify-start gap-2.5 px-3 py-2 text-base data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:font-semibold data-[state=active]:shadow-sm dark:data-[state=active]:bg-input/30"
			>
				<span
					class="size-1.5 shrink-0 rounded-full bg-primary opacity-0 transition-opacity group-data-[state=active]:opacity-100"
					aria-hidden="true"
				></span>
				Rating cutoffs
			</TabsTrigger>
		</TabsList>

		<TabsContent value="captains" class="flex-1">
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
		</TabsContent>

		<TabsContent value="players" class="flex-1">
		<!-- Available players -->
		<Card>
			<CardHeader>
				<CardTitle class="text-base">Available players</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="text-sm text-muted-foreground">
					Move players into the pool to make them eligible to be drafted. Keep the count
					even-ish across teams so each team ends up with a fair roster. Changes apply when
					you save.
				</p>
				{#if transferCore}
					<div class="mt-4">
						<div class="mb-4 flex flex-wrap items-center gap-2">
							<span class="text-sm font-medium">Organizations</span>
							{#if organizations.length === 0}
								<span class="text-sm text-muted-foreground">
									No organizations yet.
								</span>
							{:else}
								<button
									type="button"
									class="rounded-full border px-3 py-1 text-sm transition-colors {selectedOrgIds.length === 0
										? 'border-primary bg-primary text-primary-foreground'
										: 'border-input text-muted-foreground hover:text-foreground'}"
									onclick={() => {
										selectedOrgIds = [];
										recomputeSource();
									}}
								>
									All users
								</button>
								{#each organizations as org (org.id)}
									<button
										type="button"
										class="rounded-full border px-3 py-1 text-sm transition-colors {isOrgSelected(org.id)
											? 'border-primary bg-primary text-primary-foreground'
											: 'border-input text-muted-foreground hover:text-foreground'}"
										onclick={() => toggleOrg(org.id)}
									>
										{org.name}
									</button>
								{/each}
							{/if}
						</div>
						<TransferList.Root direction="horizontal">
							<TransferList.Container>
								<TransferList.Title
									title={
										selectedOrgIds.length > 0
											? `All players (filtered) — ${filteredMemberCount} member${filteredMemberCount === 1 ? '' : 's'}`
											: 'All players'
									}
								/>
								<TransferList.Toolbar
									variant="source"
									core={transferCore}
									inputPlaceholder="Search players..."
								/>
								<TransferList.Body>
									{#each transferCore.filteredSource as row (row.id)}
										<TransferList.Item side="source" {row} core={transferCore}>
											{fullName(row)}
										</TransferList.Item>
									{/each}
								</TransferList.Body>
							</TransferList.Container>
							<TransferList.Container>
								<TransferList.Title title="Available for draft" />
								<TransferList.Toolbar
									variant="target"
									core={transferCore}
									inputPlaceholder="Search players..."
								/>
								<TransferList.Body>
									{#each transferCore.filteredTarget as row (row.id)}
										<TransferList.Item side="target" {row} core={transferCore}>
											{fullName(row)}
										</TransferList.Item>
									{/each}
								</TransferList.Body>
							</TransferList.Container>
						</TransferList.Root>
						<div class="mt-4 flex items-center gap-3">
							<Button
								type="button"
								onclick={handleSavePlayers}
								disabled={savingPlayers}
							>
								{savingPlayers ? 'Saving...' : 'Save players'}
							</Button>
							<span class="text-sm text-muted-foreground">
								{transferCore.target.length} player{transferCore.target.length === 1 ? '' : 's'}
								in pool
							</span>
						</div>
					</div>
				{:else}
					<p class="mt-4 text-sm text-muted-foreground">Loading players...</p>
				{/if}
			</CardContent>
		</Card>
		</TabsContent>

		<TabsContent value="cutoffs" class="flex-1">
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
		</TabsContent>
	</Tabs>
{/if}
