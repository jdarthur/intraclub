<script lang="ts">
	import { onMount } from 'svelte';
	import { createScoringStructure, getScoreCountingTypes } from '$lib/scoringStructure';
	import type { ScoreCountingType } from '$lib/scoringStructure';
	import { goto } from '$app/navigation';

	let countingTypes = $state<ScoreCountingType[]>([]);
	let name = $state('');
	let countingType = $state<number>(0);
	let winThreshold = $state<string>('1');
	let mustWinBy = $state<string>('1');
	let instantWinThreshold = $state<string>('0');
	let error = $state('');
	let submitting = $state(false);

	onMount(async () => {
		try {
			countingTypes = await getScoreCountingTypes();
			if (countingTypes.length > 0) countingType = countingTypes[0].type;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load score counting types';
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		submitting = true;
		try {
			const created = await createScoringStructure({
				name: name.trim(),
				win_condition_counting_type: countingType,
				win_condition: {
					win_threshold: parseInt(winThreshold, 10),
					must_win_by: parseInt(mustWinBy, 10),
					instant_win_threshold: parseInt(instantWinThreshold || '0', 10)
				}
			});
			await goto(`/scoring-structures/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create scoring structure';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>New scoring structure</title>
</svelte:head>

<h1>New scoring structure</h1>
<a href="/scoring-structures">&larr; Back to scoring structures</a>

<form onsubmit={handleSubmit}>
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
	<button type="submit" disabled={submitting}>{submitting ? 'Creating...' : 'Create scoring structure'}</button>
</form>

{#if error}
	<p class="error">{error}</p>
{/if}

<style>
	.error {
		color: #c00;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		max-width: 24rem;
		margin-top: 1rem;
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
</style>
