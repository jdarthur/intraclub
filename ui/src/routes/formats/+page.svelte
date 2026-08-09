<script lang="ts">
	import { onMount } from 'svelte';
	import { listFormats } from '$lib/format';
	import type { Format } from '$lib/format';

	let formats = $state<Format[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			formats = await listFormats();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load formats';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Formats</title>
</svelte:head>

<h1>Formats</h1>
<a href="/formats/new">New format</a>

{#if loading}
	<p>Loading...</p>
{:else if error}
	<p class="error">{error}</p>
{:else if formats.length === 0}
	<p>No formats yet.</p>
{:else}
	<table>
		<thead>
			<tr>
				<th>Name</th>
			</tr>
		</thead>
		<tbody>
			{#each formats as format}
				<tr>
					<td><a href={`/formats/${format.id}`}>{format.name}</a></td>
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
