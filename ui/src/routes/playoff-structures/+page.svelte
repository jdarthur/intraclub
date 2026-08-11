<script lang="ts">
	import { onMount } from 'svelte';
	import { listPlayoffStructures } from '$lib/playoffStructure';
	import type { PlayoffStructure } from '$lib/playoffStructure';

	let structures = $state<PlayoffStructure[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			structures = await listPlayoffStructures();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load playoff structures';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Playoff Structures</title>
</svelte:head>

<h1>Playoff Structures</h1>
<a href="/playoff-structures/new">New playoff structure</a>

{#if loading}
	<p>Loading...</p>
{:else if error}
	<p class="error">{error}</p>
{:else if structures.length === 0}
	<p>No playoff structures yet.</p>
{:else}
	<table>
		<thead>
			<tr>
				<th>Playoff structure</th>
			</tr>
		</thead>
		<tbody>
			{#each structures as structure}
				<tr>
					<td><a href={`/playoff-structures/${structure.id}`}>{structure.byes} byes / {structure.number_of_teams} teams</a></td>
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
