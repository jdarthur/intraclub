<script lang="ts">
	import { onMount } from 'svelte';
	import { listRulesets } from '$lib/ruleset';
	import type { Ruleset } from '$lib/ruleset';

	let rulesets = $state<Ruleset[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			rulesets = await listRulesets();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load rulesets';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Rulesets</title>
</svelte:head>

<h1>Rulesets</h1>
<a href="/rulesets/new">New ruleset</a>

{#if loading}
	<p>Loading...</p>
{:else if error}
	<p class="error">{error}</p>
{:else if rulesets.length === 0}
	<p>No rulesets yet.</p>
{:else}
	<table>
		<thead>
			<tr>
				<th>Name</th>
				<th>Revision</th>
			</tr>
		</thead>
		<tbody>
			{#each rulesets as ruleset}
				<tr>
					<td><a href={`/rulesets/${ruleset.id}`}>{ruleset.name}</a></td>
					<td>{ruleset.revision}</td>
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
