<script lang="ts">
	import { createRating } from '$lib/rating';
	import { goto } from '$app/navigation';

	let name = $state('');
	let description = $state('');
	let error = $state('');
	let submitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		submitting = true;
		try {
			const created = await createRating({ name: name.trim(), description: description.trim() });
			await goto(`/ratings/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create rating';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>New rating</title>
</svelte:head>

<h1>New rating</h1>
<a href="/ratings">&larr; Back to ratings</a>

<form onsubmit={handleSubmit}>
	<label>
		Name
		<input type="text" bind:value={name} required />
	</label>
	<label>
		Description
		<textarea bind:value={description} required></textarea>
	</label>
	<button type="submit" disabled={submitting}>{submitting ? 'Creating...' : 'Create rating'}</button>
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
	textarea {
		padding: 0.35rem;
	}
	textarea {
		min-height: 5rem;
		resize: vertical;
	}
</style>
