<script lang="ts">
	import { onMount } from 'svelte';
	import { listRatings } from '$lib/rating';
	import type { Rating } from '$lib/rating';

	let ratings = $state<Rating[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			ratings = await listRatings();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load ratings';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Ratings</title>
</svelte:head>

<h1>Ratings</h1>
<a href="/ratings/new">New rating</a>

{#if loading}
	<p>Loading...</p>
{:else if error}
	<p class="error">{error}</p>
{:else if ratings.length === 0}
	<p>No ratings yet.</p>
{:else}
	<table>
		<thead>
			<tr>
				<th>Name</th>
				<th>Description</th>
			</tr>
		</thead>
		<tbody>
			{#each ratings as rating}
				<tr>
					<td><a href={`/ratings/${rating.id}`}>{rating.name}</a></td>
					<td>{rating.description}</td>
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
