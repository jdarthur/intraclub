<script lang="ts">
	import { onMount } from 'svelte';
	import { listScoringStructures, getScoreCountingTypes } from '$lib/scoringStructure';
	import type { ScoringStructure, ScoreCountingType } from '$lib/scoringStructure';

	let structures = $state<ScoringStructure[]>([]);
	let countingTypes = $state<ScoreCountingType[]>([]);
	let loading = $state(true);
	let error = $state('');

	function countingTypeName(type: number): string {
		return countingTypes.find((t) => t.type === type)?.name ?? String(type);
	}

	onMount(async () => {
		try {
			[structures, countingTypes] = await Promise.all([
				listScoringStructures(),
				getScoreCountingTypes()
			]);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load scoring structures';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Scoring Structures</title>
</svelte:head>

<h1>Scoring Structures</h1>
<a href="/scoring-structures/new">New scoring structure</a>

{#if loading}
	<p>Loading...</p>
{:else if error}
	<p class="error">{error}</p>
{:else if structures.length === 0}
	<p>No scoring structures yet.</p>
{:else}
	<table>
		<thead>
			<tr>
				<th>Name</th>
				<th>Counting type</th>
				<th>Win threshold</th>
			</tr>
		</thead>
		<tbody>
			{#each structures as structure}
				<tr>
					<td><a href={`/scoring-structures/${structure.id}`}>{structure.name}</a></td>
					<td>{countingTypeName(structure.win_condition_counting_type)}</td>
					<td>{structure.win_condition.win_threshold}</td>
				</tr>
			{/each}
		</tbody>
	</table>
{/if}

<style>
	.error {
		color: #c00;
	}
	table {
		border-collapse: collapse;
		margin-top: 1rem;
	}
	th,
	td {
		border: 1px solid #ccc;
		padding: 0.4rem 0.8rem;
		text-align: left;
	}
</style>
