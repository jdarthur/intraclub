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
	import * as TransferList from '$lib/components/ui/transfer-list/index.js';
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
	let allStructures = $state<ScoringStructure[]>([]);
	let secondariesError = $state('');
	let secondariesSaving = $state(false);

	// The transfer list core holds the in-memory source (eligible structures)
	// / target (assigned secondaries in tie-breaker order) lists and
	// selection. The backend requires a composite to have exactly the required
	// number of secondaries, so edits accumulate in the core and are committed
	// in one call via the Save button.
	let transferCore = $state<TransferList.Core<ScoringStructure> | null>(null);

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
	// All catalog structures of the required secondary type, regardless of
	// assignment (used to explain an empty source list).
	const typeStructures = $derived(
		secondaryType === null
			? []
			: allStructures.filter((s) => s.win_condition_counting_type === secondaryType)
	);
	// The backend requires a composite to have exactly this many secondaries.
	const canSaveSecondaries = $derived(
		secondaryType !== null &&
			requiredSecondaryCount !== null &&
			transferCore !== null &&
			transferCore.target.length === requiredSecondaryCount
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

	// Keep the transfer list's source in sync with the catalog and the form's
	// counting type: only structures of the required secondary type that are
	// not already in the target are available to assign. Recomputing here
	// (rather than only on load) means editing the counting type in the form
	// immediately refilters the available structures.
	$effect(() => {
		const core = transferCore;
		if (!core) return;
		const targetIds = new Set(core.target.map((s) => s.id));
		core.source =
			secondaryType === null
				? []
				: allStructures.filter(
						(s) =>
							s.win_condition_counting_type === secondaryType && !targetIds.has(s.id)
					);
	});

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

	function buildTransferCore(target: ScoringStructure[]): TransferList.Core<ScoringStructure> {
		const type = secondaryCountingTypeFor(countingType);
		const targetIds = new Set(target.map((s) => s.id));
		return new TransferList.Core<ScoringStructure>({
			initialSource:
				type === null
					? []
					: allStructures.filter(
							(s) => s.win_condition_counting_type === type && !targetIds.has(s.id)
						),
			initialTarget: target,
			filterPredicate: (s, search) => s.name.toLowerCase().includes(search.toLowerCase())
		});
	}

	async function loadSecondaries() {
		try {
			secondaries = await getScoringStructureSecondaries(id());
		} catch (e) {
			secondariesError = e instanceof Error ? e.message : 'Failed to load secondary scoring structures';
		}
		// Refresh the full catalog so the transfer-list source reflects the
		// structures available to assign.
		try {
			allStructures = await listScoringStructures();
		} catch {
			// ignore; the source just stays empty
		}
		transferCore = buildTransferCore(secondaries);
	}

	async function saveSecondaries() {
		secondariesError = '';
		if (!transferCore) return;
		secondariesSaving = true;
		try {
			secondaries = await setScoringStructureSecondaries(
				id(),
				transferCore.target.map((s) => s.id)
			);
			// Rebuild the transfer list from the server-confirmed ordering
			// (SecondaryIndex), clearing any in-progress selections.
			transferCore = buildTransferCore(secondaries);
			toast.success('Secondaries saved');
		} catch (err) {
			secondariesError = err instanceof Error ? err.message : 'Failed to update secondary scoring structures';
		} finally {
			secondariesSaving = false;
		}
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

		<section class="mt-6 max-w-2xl">
			<h2 class="text-xl font-semibold tracking-tight">Secondary scoring structures</h2>

			{#if secondaryType === null}
				<p class="mt-2 text-sm text-muted-foreground">
					Point-based scoring structures cannot have secondary (tie-breaker) scoring structures
					{#if transferCore && transferCore.target.length > 0}
						— move the {transferCore.target.length} assigned secondaries back to the
						available list (or change the counting type) before saving.
					{/if}
				</p>
			{:else if requiredSecondaryCount === null}
				<p class="mt-2 text-sm text-muted-foreground">
					Win conditions that require winning by 2 or more without an instant-win threshold cannot
					be composite
					{#if transferCore && transferCore.target.length > 0}
						— move the {transferCore.target.length} assigned secondaries back to the
						available list (or change the win condition) before saving.
					{/if}
				</p>
			{:else}
				<p class="mt-1 text-sm text-muted-foreground">
					The {secondaryTypeName}-based structures used to score each {unitNoun} in this
					structure, in tie-breaker order. Changes apply when you save.
				</p>
			{/if}

			{#if transferCore}
				<div class="mt-4">
					<TransferList.Root direction="horizontal">
						<TransferList.Container>
							<TransferList.Title title="Available secondaries" />
							<TransferList.Toolbar
								variant="source"
								core={transferCore}
								inputPlaceholder="Search secondaries..."
							/>
							<TransferList.Body>
								{#each transferCore.filteredSource as row (row.id)}
									<TransferList.Item side="source" {row} core={transferCore}>
										<Badge variant="secondary" class="secondary-name">{row.name}</Badge>
									</TransferList.Item>
								{/each}
							</TransferList.Body>
						</TransferList.Container>
						<TransferList.Container class="secondary-target">
							<TransferList.Title title="Assigned secondaries" />
							<TransferList.Toolbar
								variant="target"
								core={transferCore}
								inputPlaceholder="Search secondaries..."
							/>
							<TransferList.Body>
								{#each transferCore.filteredTarget as row (row.id)}
									<TransferList.Item side="target" {row} core={transferCore}>
										<Badge variant="secondary" class="secondary-name">{row.name}</Badge>
									</TransferList.Item>
								{/each}
							</TransferList.Body>
						</TransferList.Container>
					</TransferList.Root>

					{#if secondaryType !== null && requiredSecondaryCount !== null}
						<div class="mt-4 flex flex-col gap-2">
							{#if transferCore.target.length === 0}
								<p class="text-sm text-muted-foreground">
									No secondary scoring structures assigned yet.
								</p>
							{/if}

							{#if transferCore.target.length !== requiredSecondaryCount}
								<p class="text-sm text-muted-foreground">
									This win condition can play at most {requiredSecondaryCount} {unitNoun}s,
									so a composite structure needs exactly {requiredSecondaryCount} secondary
									scoring structures ({transferCore.target.length} currently assigned).
								</p>
							{:else if transferCore.target.length > 0}
								<p class="text-sm text-muted-foreground">
									All {requiredSecondaryCount} required secondary scoring structures assigned.
								</p>
							{/if}

							{#if typeStructures.length === 0}
								<p class="text-sm text-muted-foreground">
									No {secondaryTypeName}-based scoring structures available to assign.
								</p>
							{/if}

							<div class="mt-1 flex items-center gap-3">
								<Button
									type="button"
									onclick={saveSecondaries}
									disabled={secondariesSaving || !canSaveSecondaries}
								>
									{secondariesSaving ? 'Saving…' : 'Save secondaries'}
								</Button>
								<span class="text-sm text-muted-foreground">
									{transferCore.target.length} of {requiredSecondaryCount} required
								</span>
							</div>
						</div>
					{/if}
				</div>
			{:else}
				<p class="mt-4 text-sm text-muted-foreground">Loading secondaries...</p>
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
