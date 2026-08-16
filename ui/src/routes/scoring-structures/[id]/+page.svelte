<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Async } from '$lib/async.svelte';
	import { AsyncSection, PageHeader } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import {
		AlertDialog,
		AlertDialogAction,
		AlertDialogCancel,
		AlertDialogContent,
		AlertDialogDescription,
		AlertDialogFooter,
		AlertDialogHeader,
		AlertDialogTitle,
		AlertDialogTrigger
	} from '$lib/components/ui/alert-dialog/index.js';
	import {
		getScoringStructure,
		getScoreCountingTypes,
		updateScoringStructure,
		deleteScoringStructure,
		getScoringStructureSecondaries,
		setScoringStructureSecondaries,
		listScoringStructures,
		secondaryCountingTypeFor,
		maximumScoreCountingUnits,
		ScoreCountingTypes
	} from '$lib/scoringStructure';
	import type { ScoringStructure, ScoreCountingType, WinCondition } from '$lib/scoringStructure';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { NativeSelect, NativeSelectOption } from '$lib/components/ui/native-select/index.js';
	import { toast } from '$lib/toast';
	import GaugeIcon from '@lucide/svelte/icons/gauge';

	const id = () => page.params.id as string;

	const structure = new Async<ScoringStructure>();
	let countingTypes = $state<ScoreCountingType[]>([]);
	let name = $state('');
	let countingType = $state<number>(0);
	let winThreshold = $state<string>('');
	let mustWinBy = $state<string>('');
	let instantWinThreshold = $state<string>('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);

	// Secondary (tie-breaker) scoring structures assigned to this structure,
	// ordered by SecondaryIndex.
	let secondaries = $state<ScoringStructure[]>([]);
	// Local draft of secondary edits, built incrementally. The backend requires
	// a composite to have exactly the required number of secondaries, so edits
	// accumulate here and are committed in one call via the Save button.
	let draftSecondaries = $state<ScoringStructure[]>([]);
	let allStructures = $state<ScoringStructure[]>([]);
	let selectedSecondaryId = $state('');
	let secondariesError = $state('');
	let secondariesSaving = $state(false);

	// The win condition as currently being edited in the form (what a save
	// would validate against).
	const formWinCondition = $derived<WinCondition>({
		win_threshold: parseInt(winThreshold, 10) || 0,
		must_win_by: parseInt(mustWinBy, 10) || 0,
		instant_win_threshold: parseInt(instantWinThreshold || '0', 10) || 0
	});

	// The counting type secondaries must use (null = cannot be composite), and
	// the number of secondaries this win condition requires (null = cannot be
	// composite).
	const secondaryType = $derived(secondaryCountingTypeFor(countingType));
	const requiredSecondaryCount = $derived(
		secondaryType === null ? null : maximumScoreCountingUnits(formWinCondition)
	);
	const eligibleStructures = $derived(
		secondaryType === null
			? []
			: allStructures.filter(
					(s) =>
						s.win_condition_counting_type === secondaryType &&
						!draftSecondaries.some((d) => d.id === s.id)
				)
	);
	// The backend requires a composite to have exactly this many secondaries.
	const canSaveSecondaries = $derived(
		secondaryType !== null &&
			requiredSecondaryCount !== null &&
			draftSecondaries.length === requiredSecondaryCount
	);
	const secondaryTypeName = $derived(
		countingTypes.find((t) => t.type === secondaryType)?.name.toLowerCase() ?? ''
	);
	const unitNoun = $derived(
		countingType === ScoreCountingTypes.Set
			? 'set'
			: countingType === ScoreCountingTypes.Game
				? 'game'
				: ''
	);

	function countingTypeName(type: number): string {
		return countingTypes.find((t) => t.type === type)?.name ?? String(type);
	}

	onMount(async () => {
		try {
			countingTypes = await getScoreCountingTypes();
		} catch (e) {
			structure.error = e instanceof Error ? e.message : 'Failed to load score counting types';
			structure.status = 'error';
			return;
		}
		structure.run(load);
	});

	async function load(): Promise<ScoringStructure> {
		const s = await getScoringStructure(id());
		name = s.name;
		countingType = s.win_condition_counting_type;
		winThreshold = String(s.win_condition.win_threshold);
		mustWinBy = String(s.win_condition.must_win_by);
		instantWinThreshold = String(s.win_condition.instant_win_threshold);
		await loadSecondaries();
		return s;
	}

	async function loadSecondaries() {
		try {
			secondaries = await getScoringStructureSecondaries(id());
			draftSecondaries = secondaries;
		} catch (e) {
			secondariesError = e instanceof Error ? e.message : 'Failed to load secondary scoring structures';
		}
		// Refresh the full catalog so the add dropdown reflects available structures.
		try {
			allStructures = await listScoringStructures();
		} catch {
			// ignore; the add dropdown just stays empty
		}
	}

	async function saveSecondaries() {
		secondariesError = '';
		secondariesSaving = true;
		try {
			secondaries = await setScoringStructureSecondaries(
				id(),
				draftSecondaries.map((s) => s.id)
			);
			draftSecondaries = secondaries;
			toast.success('Secondaries saved');
		} catch (err) {
			secondariesError = err instanceof Error ? err.message : 'Failed to update secondary scoring structures';
		} finally {
			secondariesSaving = false;
		}
	}

	function addToDraft() {
		const chosen = allStructures.find((s) => s.id === selectedSecondaryId);
		if (!chosen) return;
		draftSecondaries = [...draftSecondaries, chosen];
		selectedSecondaryId = '';
	}

	function removeFromDraft(index: number) {
		draftSecondaries = draftSecondaries.filter((_, i) => i !== index);
	}

	function moveDraft(index: number, direction: -1 | 1) {
		const target = index + direction;
		if (target < 0 || target >= draftSecondaries.length) return;
		const next = [...draftSecondaries];
		[next[index], next[target]] = [next[target], next[index]];
		draftSecondaries = next;
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const updated = await updateScoringStructure(id(), {
				name: name.trim(),
				win_condition_counting_type: countingType,
				win_condition: {
					win_threshold: parseInt(winThreshold, 10),
					must_win_by: parseInt(mustWinBy, 10),
					instant_win_threshold: parseInt(instantWinThreshold || '0', 10)
				}
			});
			structure.data = updated;
			toast.success('Scoring structure saved');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update scoring structure';
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		deleteOpen = false;
		error = '';
		deleting = true;
		try {
			await deleteScoringStructure(id());
			toast.success('Scoring structure deleted');
			await goto('/scoring-structures');
		} catch (err) {
			// e.g. the structure is referenced by a Schedule/PlayoffStructure and PreDelete blocks it
			error = err instanceof Error ? err.message : 'Failed to delete scoring structure';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | {structure.data ? structure.data.name : 'Scoring Structure'}</title>
</svelte:head>

<PageHeader
	title={structure.data?.name}
	icon={GaugeIcon}
	backHref="/scoring-structures"
	backLabel="Back to scoring structures"
/>

<AsyncSection state={structure}>
	{#snippet children(s)}
		<Card class="mt-6 max-w-md">
			<CardHeader>
				<CardTitle>Scoring structure details</CardTitle>
			</CardHeader>
			<CardContent>
				<dl class="flex flex-col gap-3">
					<div class="flex flex-col gap-1">
						<dt class="text-sm text-muted-foreground">Counting type</dt>
						<dd>{countingTypeName(s.win_condition_counting_type)}</dd>
					</div>
					<div class="flex flex-col gap-1">
						<dt class="text-sm text-muted-foreground">Win threshold</dt>
						<dd>{s.win_condition.win_threshold}</dd>
					</div>
					<div class="flex flex-col gap-1">
						<dt class="text-sm text-muted-foreground">Must win by</dt>
						<dd>{s.win_condition.must_win_by}</dd>
					</div>
					<div class="flex flex-col gap-1">
						<dt class="text-sm text-muted-foreground">Instant win threshold</dt>
						<dd>{s.win_condition.instant_win_threshold}</dd>
					</div>
				</dl>
			</CardContent>
		</Card>

		<Card class="mt-6 max-w-md">
			<CardHeader>
				<CardTitle>Edit scoring structure</CardTitle>
			</CardHeader>
			<CardContent>
				<form onsubmit={handleSave} class="flex flex-col gap-4">
					<div class="flex flex-col gap-2">
						<Label for="name">Name</Label>
						<Input id="name" type="text" bind:value={name} required />
					</div>
					<div class="flex flex-col gap-2">
						<Label for="countingType">Score counting type</Label>
						<NativeSelect id="countingType" bind:value={countingType} class="w-full">
							{#each countingTypes as ct}
								<NativeSelectOption value={ct.type}>{ct.name}</NativeSelectOption>
							{/each}
						</NativeSelect>
					</div>
					<div class="flex flex-col gap-2">
						<Label for="winThreshold">Win threshold</Label>
						<Input id="winThreshold" type="number" min="1" bind:value={winThreshold} required />
					</div>
					<div class="flex flex-col gap-2">
						<Label for="mustWinBy">Must win by</Label>
						<Input id="mustWinBy" type="number" min="1" bind:value={mustWinBy} required />
					</div>
					<div class="flex flex-col gap-2">
						<Label for="instantWinThreshold">Instant win threshold</Label>
						<Input
							id="instantWinThreshold"
							type="number"
							min="0"
							bind:value={instantWinThreshold}
							placeholder="0 (disabled)"
						/>
					</div>
					<Button type="submit" disabled={saving} class="w-fit">
						{saving ? 'Saving...' : 'Save changes'}
					</Button>
				</form>
			</CardContent>
		</Card>

		<section class="mt-6 max-w-md">
			<h2 class="text-xl font-semibold tracking-tight">Secondary scoring structures</h2>

			{#if draftSecondaries.length > 0}
				<ul class="mt-4 flex flex-col gap-2">
					{#each draftSecondaries as secondary, i}
						<li class="flex items-center justify-between gap-2 rounded-lg border p-2 pl-3">
							<span class="flex items-center gap-2">
								<span class="text-sm text-muted-foreground">{i + 1}.</span>
								<Badge variant="secondary" class="secondary-name">{secondary.name}</Badge>
							</span>
							<span class="flex items-center gap-1">
								<Button
									type="button"
									variant="outline"
									size="sm"
									onclick={() => moveDraft(i, -1)}
									disabled={i === 0}
								>
									↑
								</Button>
								<Button
									type="button"
									variant="outline"
									size="sm"
									onclick={() => moveDraft(i, 1)}
									disabled={i === draftSecondaries.length - 1}
								>
									↓
								</Button>
								<Button
									type="button"
									variant="outline"
									size="sm"
									onclick={() => removeFromDraft(i)}
								>
									Remove
								</Button>
							</span>
						</li>
					{/each}
				</ul>
			{/if}

			{#if secondaryType === null}
				<p class="mt-2 text-sm text-muted-foreground">
					Point-based scoring structures cannot have secondary (tie-breaker) scoring structures
					{#if draftSecondaries.length > 0}
						— remove the {draftSecondaries.length} assigned secondaries (or change the counting type)
						before saving.
					{/if}
				</p>
			{:else if requiredSecondaryCount === null}
				<p class="mt-2 text-sm text-muted-foreground">
					Win conditions that require winning by 2 or more without an instant-win threshold cannot
					be composite
					{#if draftSecondaries.length > 0}
						— remove the {draftSecondaries.length} assigned secondaries (or change the win condition)
						before saving.
					{/if}
				</p>
			{:else}
				<p class="mt-1 text-sm text-muted-foreground">
					The {secondaryTypeName}-based structures used to score each {unitNoun} in this structure,
					in tie-breaker order.
				</p>

				{#if draftSecondaries.length === 0}
					<p class="mt-4 text-muted-foreground">No secondary scoring structures assigned yet.</p>
				{/if}

				{#if draftSecondaries.length !== requiredSecondaryCount}
					<p class="mt-2 text-sm text-muted-foreground">
						This win condition can play at most {requiredSecondaryCount} {unitNoun}s, so a
						composite structure needs exactly {requiredSecondaryCount} secondary scoring
						structures ({draftSecondaries.length} currently assigned).
					</p>
				{:else if draftSecondaries.length > 0}
					<p class="mt-2 text-sm text-muted-foreground">
						All {requiredSecondaryCount} required secondary scoring structures assigned.
					</p>
				{/if}

				{#if eligibleStructures.length > 0}
					<div class="mt-4 flex items-center gap-2">
						<NativeSelect
							bind:value={selectedSecondaryId}
							aria-label="Secondary structure to assign"
							class="flex-1"
						>
							<NativeSelectOption value="" disabled
								>Select a {secondaryTypeName} structure…</NativeSelectOption
							>
							{#each eligibleStructures as s}
								<NativeSelectOption value={s.id}>{s.name}</NativeSelectOption>
							{/each}
						</NativeSelect>
						<Button
							type="button"
							onclick={addToDraft}
							disabled={
								!selectedSecondaryId || draftSecondaries.length >= requiredSecondaryCount
							}
						>
							Add secondary
						</Button>
					</div>
				{:else}
					<p class="mt-4 text-sm text-muted-foreground">
						No {secondaryTypeName}-based scoring structures available to assign.
					</p>
				{/if}

				<Button
					type="button"
					onclick={saveSecondaries}
					disabled={secondariesSaving || !canSaveSecondaries}
					class="mt-4"
				>
					{secondariesSaving
						? 'Saving…'
						: canSaveSecondaries
							? 'Save secondaries'
							: `Add ${requiredSecondaryCount - draftSecondaries.length} more to save`}
				</Button>
			{/if}

			{#if secondariesError}
				<Alert variant="destructive" class="mt-3">
					<AlertDescription>{secondariesError}</AlertDescription>
				</Alert>
			{/if}
		</section>

		{#if error}
			<Alert variant="destructive" class="mt-4 max-w-md">
				<AlertDescription>{error}</AlertDescription>
			</Alert>
		{/if}

		<div class="mt-4">
			<AlertDialog bind:open={deleteOpen}>
				<AlertDialogTrigger
					disabled={deleting}
					class={buttonVariants({ variant: 'destructive' })}
				>
					{deleting ? 'Deleting...' : 'Delete scoring structure'}
				</AlertDialogTrigger>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Delete scoring structure?</AlertDialogTitle>
						<AlertDialogDescription>
							This permanently removes this scoring structure and cannot be undone.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Cancel</AlertDialogCancel>
						<AlertDialogAction variant="destructive" onclick={handleDelete}>
							Delete
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</div>
	{/snippet}
</AsyncSection>
