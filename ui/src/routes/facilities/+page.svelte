<script lang="ts">
	import { onMount } from 'svelte';
	import { listFacilities } from '$lib/facility';
	import type { Facility } from '$lib/facility';

	let facilities = $state<Facility[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			facilities = await listFacilities();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load facilities';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Facilities</title>
</svelte:head>

<h1>Facilities</h1>
<a href="/facilities/new">New facility</a>

{#if loading}
	<p>Loading...</p>
{:else if error}
	<p class="error">{error}</p>
{:else if facilities.length === 0}
	<p>No facilities yet.</p>
{:else}
	<table>
		<thead>
			<tr>
				<th>Name</th>
				<th>Address</th>
				<th>Courts</th>
			</tr>
		</thead>
		<tbody>
			{#each facilities as facility}
				<tr>
					<td><a href={`/facilities/${facility.id}`}>{facility.name}</a></td>
					<td>{facility.address}</td>
					<td>{facility.courts}</td>
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
