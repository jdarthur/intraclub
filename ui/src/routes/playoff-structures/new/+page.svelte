<script lang="ts">
	import { createPlayoffStructure } from '$lib/playoffStructure';
	import { goto } from '$app/navigation';

	let byes = $state<string>('0');
	let numberOfTeams = $state<string>('8');
	let error = $state('');
	let submitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		submitting = true;
		try {
			const created = await createPlayoffStructure({
				byes: parseInt(byes || '0', 10),
				number_of_teams: parseInt(numberOfTeams, 10)
			});
			await goto(`/playoff-structures/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create playoff structure';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>New playoff structure</title>
</svelte:head>

<h1>New playoff structure</h1>
<a href="/playoff-structures">&larr; Back to playoff structures</a>

<form onsubmit={handleSubmit}>
	<label>
		Byes
		<input type="number" min="0" bind:value={byes} required />
	</label>
	<label>
		Number of teams
		<input type="number" min="2" bind:value={numberOfTeams} required />
	</label>
	<button type="submit" disabled={submitting}>{submitting ? 'Creating...' : 'Create playoff structure'}</button>
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
