<script lang="ts">
	import { createFormat } from '$lib/format';
	import { goto } from '$app/navigation';

	let name = $state('');
	let error = $state('');
	let submitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		submitting = true;
		try {
			const created = await createFormat({ name: name.trim() });
			await goto(`/formats/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create format';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>New format</title>
</svelte:head>

<h1>New format</h1>
<a href="/formats">&larr; Back to formats</a>

<form onsubmit={handleSubmit}>
	<label>
		Name
		<input type="text" bind:value={name} required />
	</label>
	<button type="submit" disabled={submitting}>{submitting ? 'Creating...' : 'Create format'}</button>
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
