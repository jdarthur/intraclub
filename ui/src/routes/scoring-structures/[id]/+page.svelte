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
		if (!confirm('Delete this scoring structure?')) return;
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
	<title>{structure ? structure.name : 'Scoring Structure'}</title>
</svelte:head>

{#if loadError}
	<h1>Scoring Structure</h1>
	<p class="error">{loadError}</p>
	<a href="/scoring-structures">&larr; Back to scoring structures</a>
{:else if !structure}
	<h1>Scoring Structure</h1>
	<p>Loading...</p>
{:else}
	<h1>{structure.name}</h1>
	<a href="/scoring-structures">&larr; Back to scoring structures</a>

	<dl class="meta">
		<div>
			<dt>Counting type</dt>
			<dd>{countingTypeName(structure.win_condition_counting_type)}</dd>
		</div>
		<div>
			<dt>Win threshold</dt>
			<dd>{structure.win_condition.win_threshold}</dd>
		</div>
		<div>
			<dt>Must win by</dt>
			<dd>{structure.win_condition.must_win_by}</dd>
		</div>
		<div>
			<dt>Instant win threshold</dt>
			<dd>{structure.win_condition.instant_win_threshold}</dd>
		</div>
	</dl>

	<h2>Edit</h2>
	<form onsubmit={handleSave}>
		<label>
			Name
			<input type="text" bind:value={name} required />
		</label>
		<label>
			Score counting type
			<select bind:value={countingType}>
				{#each countingTypes as ct}
					<option value={ct.type}>{ct.name}</option>
				{/each}
			</select>
		</label>
		<label>
			Win threshold
			<input type="number" min="1" bind:value={winThreshold} required />
		</label>
		<label>
			Must win by
			<input type="number" min="1" bind:value={mustWinBy} required />
		</label>
		<label>
			Instant win threshold
			<input type="number" min="0" bind:value={instantWinThreshold} placeholder="0 (disabled)" />
		</label>
		<button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save changes'}</button>
	</form>

	<button type="button" onclick={handleDelete} disabled={deleting} class="danger">
		{deleting ? 'Deleting...' : 'Delete scoring structure'}
	</button>

	{#if error}
		<p class="error">{error}</p>
	{/if}
{/if}

<style>
	.error {
		color: #c00;
	}
	.meta {
		display: flex;
		gap: 1.5rem;
		margin: 1rem 0;
	}
	.meta dt {
		font-size: 0.85rem;
		color: #666;
	}
	.meta dd {
		margin: 0;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		max-width: 24rem;
		margin-top: 0.5rem;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	input,
	select {
		padding: 0.35rem;
	}
	.danger {
		margin-top: 1rem;
		color: #c00;
	}
</style>
