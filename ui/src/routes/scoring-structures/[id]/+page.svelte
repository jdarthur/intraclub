<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		getScoringStructure,
		getScoreCountingTypes,
		updateScoringStructure,
		deleteScoringStructure
	} from '$lib/scoringStructure';
	import type { ScoringStructure, ScoreCountingType } from '$lib/scoringStructure';
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { NativeSelect, NativeSelectOption } from '$lib/components/ui/native-select/index.js';
	import {
		Popover,
		PopoverClose,
		PopoverContent,
		PopoverHeader,
		PopoverTitle,
		PopoverTrigger
	} from '$lib/components/ui/popover/index.js';

	const id = () => page.params.id as string;

	let structure = $state<ScoringStructure | null>(null);
	let countingTypes = $state<ScoreCountingType[]>([]);
	let loadError = $state('');
	let name = $state('');
	let countingType = $state<number>(0);
	let winThreshold = $state<string>('');
	let mustWinBy = $state<string>('');
	let instantWinThreshold = $state<string>('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);

	function countingTypeName(type: number): string {
		return countingTypes.find((t) => t.type === type)?.name ?? String(type);
	}

	onMount(async () => {
		try {
			countingTypes = await getScoreCountingTypes();
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load score counting types';
			return;
		}
		load();
	});

	async function load() {
		loadError = '';
		try {
			const s = await getScoringStructure(id());
			structure = s;
			name = s.name;
			countingType = s.win_condition_counting_type;
			winThreshold = String(s.win_condition.win_threshold);
			mustWinBy = String(s.win_condition.must_win_by);
			instantWinThreshold = String(s.win_condition.instant_win_threshold);
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load scoring structure';
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
			structure = updated;
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
			await goto('/scoring-structures');
		} catch (err) {
			// e.g. the structure is referenced by a Schedule/PlayoffStructure and PreDelete blocks it
			error = err instanceof Error ? err.message : 'Failed to delete scoring structure';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | {structure ? structure.name : 'Scoring Structure'}</title>
</svelte:head>

{#if loadError}
	<h1 class="text-2xl font-semibold tracking-tight">Scoring Structure</h1>
	<p class="text-sm font-medium text-destructive">{loadError}</p>
	<a href="/scoring-structures" class="text-sm text-muted-foreground hover:text-foreground"
		>&larr; Back to scoring structures</a
	>
{:else if !structure}
	<h1 class="text-2xl font-semibold tracking-tight">Scoring Structure</h1>
	<p class="text-muted-foreground">Loading...</p>
{:else}
	<div class="flex items-center gap-4">
		<h1 class="text-2xl font-semibold tracking-tight">{structure.name}</h1>
		<a href="/scoring-structures" class="text-sm text-muted-foreground hover:text-foreground"
			>&larr; Back to scoring structures</a
		>
	</div>

	<Card class="mt-6 max-w-md">
		<CardHeader>
			<CardTitle class="text-base">Scoring structure details</CardTitle>
		</CardHeader>
		<CardContent>
			<dl class="flex flex-col gap-3">
				<div class="flex flex-col gap-1">
					<dt class="text-sm text-muted-foreground">Counting type</dt>
					<dd>{countingTypeName(structure.win_condition_counting_type)}</dd>
				</div>
				<div class="flex flex-col gap-1">
					<dt class="text-sm text-muted-foreground">Win threshold</dt>
					<dd>{structure.win_condition.win_threshold}</dd>
				</div>
				<div class="flex flex-col gap-1">
					<dt class="text-sm text-muted-foreground">Must win by</dt>
					<dd>{structure.win_condition.must_win_by}</dd>
				</div>
				<div class="flex flex-col gap-1">
					<dt class="text-sm text-muted-foreground">Instant win threshold</dt>
					<dd>{structure.win_condition.instant_win_threshold}</dd>
				</div>
			</dl>
		</CardContent>
	</Card>

	<Card class="mt-6 max-w-md">
		<CardHeader>
			<CardTitle class="text-base">Edit scoring structure</CardTitle>
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

	<div class="mt-4">
		<Popover bind:open={deleteOpen}>
			<PopoverTrigger disabled={deleting} class={buttonVariants({ variant: 'destructive' })}>
				{deleting ? 'Deleting...' : 'Delete scoring structure'}
			</PopoverTrigger>
			<PopoverContent class="w-80">
				<PopoverHeader>
					<PopoverTitle>Delete scoring structure?</PopoverTitle>
					<p class="text-sm text-muted-foreground">
						This permanently removes this scoring structure and cannot be undone.
					</p>
				</PopoverHeader>
				<div class="flex justify-end gap-2">
					<PopoverClose class={buttonVariants({ variant: 'outline', size: 'sm' })}>Cancel</PopoverClose>
					<Button variant="destructive" size="sm" onclick={handleDelete}>Delete</Button>
				</div>
			</PopoverContent>
		</Popover>
	</div>

	{#if error}
		<p class="mt-4 text-sm font-medium text-destructive">{error}</p>
	{/if}
{/if}
