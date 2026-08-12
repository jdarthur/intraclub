<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getDraft } from '$lib/draft';
	import type { Draft } from '$lib/draft';
	import { listFormats, getFormatPossibleRatings } from '$lib/format';
	import type { Format } from '$lib/format';
	import type { Rating } from '$lib/rating';
	import { listUsers, fullName } from '$lib/user';
	import type { User } from '$lib/user';
	import { getCurrentUserId } from '$lib/auth';
	import { listDraftAvailablePlayers } from '$lib/draftAvailablePlayer';
	import type { DraftAvailablePlayer } from '$lib/draftAvailablePlayer';
	import {
		listPreDraftGrades,
		createPreDraftGrade,
		updatePreDraftGrade
	} from '$lib/preDraftGrade';
	import type { PreDraftGrade, PreDraftModifier } from '$lib/preDraftGrade';
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

	// Modifier values: 0 = weak, 1 = average, 2 = strong.
	const MODIFIERS: { value: string; label: string }[] = [
		{ value: '0', label: 'Weak' },
		{ value: '1', label: 'Average' },
		{ value: '2', label: 'Strong' }
	];

	let draft = $state<Draft | null>(null);
	let format = $state<Format | null>(null);
	let possibleRatings = $state<Rating[]>([]);
	let users = $state<User[]>([]);
	// The current user is the grader; their id comes from the JWT's `sub`.
	const currentUserId = getCurrentUserId();

	let availablePlayers = $state<DraftAvailablePlayer[]>([]);
	let grades = $state<PreDraftGrade[]>([]);

	let loading = $state(true);
	let loadError = $state('');
	let actionError = $state('');
	let saving = $state(false);

	// One grade entry per available player, seeded from the current user's
	// existing grade (there is at most one grade per player/grader/draft).
	interface GradeEntry {
		playerId: string;
		modifier: string;
		ratingId: string;
		gradeId: string | null;
	}
	let entries = $state<GradeEntry[]>([]);

	async function load() {
		loading = true;
		loadError = '';
		actionError = '';
		try {
			const draftData = await getDraft(id());
			const [formatList, userList, availableList, gradeList] = await Promise.all([
				listFormats(),
				listUsers(),
				listDraftAvailablePlayers(),
				listPreDraftGrades()
			]);

			draft = draftData;
			users = userList;

			const fmt = formatList.find((f) => f.id === draftData.format) ?? null;
			format = fmt;
			possibleRatings = fmt ? await getFormatPossibleRatings(fmt.id) : [];

			const forThisDraft = availableList.filter((a) => a.draft_id === draftData.id);
			availablePlayers = forThisDraft;
			grades = gradeList.filter((g) => g.DraftId === draftData.id);

			// Seed an entry per available player from the current user's grade.
			entries = forThisDraft.map((a) => {
				const mine = grades.find(
					(g) => g.PlayerId === a.player_id && g.GraderId === currentUserId
				);
				return {
					playerId: a.player_id,
					modifier: mine ? String(mine.Modifier) : '',
					ratingId: mine ? mine.Rating : '',
					gradeId: mine ? mine.id : null
				};
			});
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load grading page';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function userName(userId: string): string {
		const u = users.find((x) => x.id === userId);
		return u ? fullName(u) : userId;
	}

	const isCompleted = $derived(
		!!draft && !!draft.completed_at && draft.completed_at !== ZERO_TIME
	);

	// --- Aggregate ranking ---
	// Mirrors model.PreDraftGrade.NumericRating / GetDraftAggregateForPlayer /
	// GetSortedListOfAllPreDraftGradesDescending so captains can compare players.
	const possibleRatingIds = $derived(possibleRatings.map((r) => r.id));

	function numericRating(grade: PreDraftGrade, ratingIds: string[]): number {
		const idx = ratingIds.indexOf(grade.Rating);
		if (idx === -1) return -1;
		const base = (ratingIds.length - idx - 1) * 3 + 1;
		return base + grade.Modifier;
	}

	const ranking = $derived(
		availablePlayers
			.map((a) => {
				const forPlayer = grades.filter((g) => g.PlayerId === a.player_id);
				const sum = forPlayer.reduce((acc, g) => acc + numericRating(g, possibleRatingIds), 0);
				const aggregate = forPlayer.length > 0 ? sum / forPlayer.length : 0;
				return { playerId: a.player_id, aggregate, graded: forPlayer.length > 0 };
			})
			.sort((a, b) => b.aggregate - a.aggregate)
	);

	// --- Saving grades ---

	function entryError(): string | null {
		for (const entry of entries) {
			const name = userName(entry.playerId);
			if (entry.ratingId === '') return `Pick a rating for ${name}.`;
			if (entry.modifier !== '0' && entry.modifier !== '1' && entry.modifier !== '2') {
				return `Pick a modifier for ${name}.`;
			}
		}
		return null;
	}

	async function handleSave() {
		actionError = '';
		if (isCompleted) {
			actionError = 'This draft is already completed, so grades can no longer be changed.';
			return;
		}
		const error = entryError();
		if (error) {
			actionError = error;
			return;
		}
		saving = true;
		try {
			for (const entry of entries) {
				const input = {
					PlayerId: entry.playerId,
					DraftId: draft!.id,
					Modifier: Number(entry.modifier) as PreDraftModifier,
					Rating: entry.ratingId
				};
				if (entry.gradeId) {
					await updatePreDraftGrade(entry.gradeId, input);
				} else {
					await createPreDraftGrade(input);
				}
			}
			await load();
		} catch (e) {
			actionError = e instanceof Error ? e.message : 'Failed to save grades';
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>{draft ? `${draft.name} — Grades` : 'Pre-draft grades'}</title>
</svelte:head>

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">{draft?.name ?? 'Pre-draft grades'}</h1>
	<a href={`/drafts/${id()}`} class="text-sm text-muted-foreground hover:text-foreground">
		&larr; Draft setup
	</a>
</div>

{#if loading}
	<p class="mt-4 text-muted-foreground">Loading...</p>
{:else if loadError}
	<p class="mt-4 text-sm font-medium text-destructive">{loadError}</p>
{:else if draft}
	<div class="mt-4 flex items-center gap-4">
		<Badge variant={isCompleted ? 'secondary' : 'default'}>
			{isCompleted ? 'Draft completed' : 'Draft in progress'}
		</Badge>
		{#if isCompleted}
			<p class="text-sm text-muted-foreground">
				Grades can no longer be changed once the draft is completed.
			</p>
		{/if}
	</div>

	{#if actionError}
		<p class="mt-4 text-sm font-medium text-destructive">{actionError}</p>
	{/if}

	<div class="mt-6 grid gap-6 lg:grid-cols-2">
		<!-- Grade assignment -->
		<Card>
			<CardHeader>
				<CardTitle class="text-base">Assign grades</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="text-sm text-muted-foreground">
					For each draftable player, pick a rating from this draft's format and mark them
					as a weak, average, or strong version of that rating.
				</p>
				{#if availablePlayers.length === 0}
					<p class="mt-4 text-sm text-muted-foreground">No draftable players yet.</p>
				{:else}
					<form
						class="mt-4 flex flex-col gap-4"
						onsubmit={(e) => {
							e.preventDefault();
							handleSave();
						}}
					>
						{#each entries as entry, i}
							<div class="rounded-lg border px-3 py-3">
								<div class="flex items-center justify-between gap-4">
									<Label for={`modifier-${i}`} class="font-medium">
										{userName(entry.playerId)}
									</Label>
									{#if entry.gradeId}
										<Badge variant="outline">saved</Badge>
									{/if}
								</div>
								<div class="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center">
									<NativeSelect
										id={`modifier-${i}`}
										class="sm:w-40"
										disabled={isCompleted}
										value={entry.modifier}
										onchange={(e) => {
											entry.modifier = (e.currentTarget as HTMLSelectElement).value;
										}}
									>
										<NativeSelectOption value="" disabled>Modifier…</NativeSelectOption>
										{#each MODIFIERS as m}
											<NativeSelectOption value={m.value}>{m.label}</NativeSelectOption>
										{/each}
									</NativeSelect>
									<NativeSelect
										id={`rating-${i}`}
										class="sm:w-48"
										disabled={isCompleted}
										value={entry.ratingId}
										onchange={(e) => {
											entry.ratingId = (e.currentTarget as HTMLSelectElement).value;
										}}
									>
										<NativeSelectOption value="" disabled>Rating…</NativeSelectOption>
										{#each possibleRatings as r}
											<NativeSelectOption value={r.id}>{r.name}</NativeSelectOption>
										{/each}
									</NativeSelect>
								</div>
							</div>
						{/each}
						<div class="mt-2">
							<Button type="submit" disabled={saving || isCompleted}>
								{saving ? 'Saving...' : 'Save grades'}
							</Button>
						</div>
					</form>
				{/if}
			</CardContent>
		</Card>

		<!-- Aggregate ranking -->
		<Card>
			<CardHeader>
				<CardTitle class="text-base">Ranking</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="text-sm text-muted-foreground">
					Aggregate of all grades, highest first, to help captains evaluate the field.
				</p>
				{#if ranking.length === 0}
					<p class="mt-4 text-sm text-muted-foreground">No draftable players to rank yet.</p>
				{:else}
					<div class="mt-4 overflow-hidden rounded-lg border">
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>Rank</TableHead>
									<TableHead>Player</TableHead>
									<TableHead>Score</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{#each ranking as row, i}
									<TableRow>
										<TableCell>{i + 1}</TableCell>
										<TableCell>{userName(row.playerId)}</TableCell>
										<TableCell>
											{row.graded ? row.aggregate.toFixed(2) : '—'}
										</TableCell>
									</TableRow>
								{/each}
							</TableBody>
						</Table>
					</div>
				{/if}
			</CardContent>
		</Card>
	</div>
{/if}
