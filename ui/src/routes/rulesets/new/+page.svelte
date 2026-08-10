<script lang="ts">
	import { createRuleset } from '$lib/ruleset';
	import { goto } from '$app/navigation';

	let name = $state('');
	let error = $state('');
	let submitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		submitting = true;
		try {
			const created = await createRuleset(name.trim());
			await goto(`/rulesets/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create ruleset';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>New ruleset</title>
</svelte:head>

<h1>New ruleset</h1>
<a href="/rulesets">&larr; Back to rulesets</a>

<form onsubmit={handleSubmit}>
	<label>
		Name
		<input type="text" bind:value={name} required />
	</label>
	<button type="submit" disabled={submitting}>{submitting ? 'Creating...' : 'Create ruleset'}</button>
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
	input {
		padding: 0.35rem;
	}
</style>
