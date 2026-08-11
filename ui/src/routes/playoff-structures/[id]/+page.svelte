<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		getPlayoffStructure,
		updatePlayoffStructure,
		deletePlayoffStructure
	} from '$lib/playoffStructure';
	import type { PlayoffStructure } from '$lib/playoffStructure';

	const id = () => page.params.id as string;

	let structure = $state<PlayoffStructure | null>(null);
	let loadError = $state('');
	let byes = $state<string>('');
	let numberOfTeams = $state<string>('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);

	onMount(load);

	async function load() {
		loadError = '';
		try {
			const s = await getPlayoffStructure(id());
			structure = s;
			byes = String(s.byes);
			numberOfTeams = String(s.number_of_teams);
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load playoff structure';
		}
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const updated = await updatePlayoffStructure(id(), {
				byes: parseInt(byes || '0', 10),
				number_of_teams: parseInt(numberOfTeams, 10)
			});
			structure = updated;
			byes = String(updated.byes);
			numberOfTeams = String(updated.number_of_teams);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update playoff structure';
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		if (!confirm('Delete this playoff structure?')) return;
		error = '';
		deleting = true;
		try {
			await deletePlayoffStructure(id());
			await goto('/playoff-structures');
		} catch (err) {
			// e.g. the structure is referenced by a Season and PreDelete blocks it
			error = err instanceof Error ? err.message : 'Failed to delete playoff structure';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>Playoff Structure</title>
</svelte:head>

{#if loadError}
	<h1>Playoff Structure</h1>
	<p class="error">{loadError}</p>
	<a href="/playoff-structures">&larr; Back to playoff structures</a>
{:else if !structure}
	<h1>Playoff Structure</h1>
	<p>Loading...</p>
{:else}
	<h1>Playoff Structure</h1>
	<a href="/playoff-structures">&larr; Back to playoff structures</a>

	<dl class="meta">
		<div>
			<dt>Byes</dt>
			<dd>{structure.byes}</dd>
		</div>
		<div>
			<dt>Number of teams</dt>
			<dd>{structure.number_of_teams}</dd>
		</div>
	</dl>

	<h2>Edit</h2>
	<form onsubmit={handleSave}>
		<label>
			Byes
			<input type="number" min="0" bind:value={byes} required />
		</label>
		<label>
			Number of teams
			<input type="number" min="2" bind:value={numberOfTeams} required />
		</label>
		<button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save changes'}</button>
	</form>

	<button type="button" onclick={handleDelete} disabled={deleting} class="danger">
		{deleting ? 'Deleting...' : 'Delete playoff structure'}
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
	input {
		padding: 0.35rem;
	}
	.danger {
		margin-top: 1rem;
		color: #c00;
	}
</style>
