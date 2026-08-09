<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { getRating, updateRating, deleteRating } from '$lib/rating';
	import type { Rating } from '$lib/rating';

	const id = () => page.params.id as string;

	let rating = $state<Rating | null>(null);
	let loadError = $state('');
	let name = $state('');
	let description = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);

	onMount(load);

	async function load() {
		loadError = '';
		try {
			const r = await getRating(id());
			rating = r;
			name = r.name;
			description = r.description;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load rating';
		}
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const updated = await updateRating(id(), { name: name.trim(), description: description.trim() });
			rating = updated;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update rating';
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		if (!confirm('Delete this rating?')) return;
		error = '';
		deleting = true;
		try {
			await deleteRating(id());
			await goto('/ratings');
		} catch (err) {
			// e.g. the rating is assigned to a Format and PreDelete blocks it
			error = err instanceof Error ? err.message : 'Failed to delete rating';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>{rating ? rating.name : 'Rating'}</title>
</svelte:head>

{#if loadError}
	<h1>Rating</h1>
	<p class="error">{loadError}</p>
	<a href="/ratings">&larr; Back to ratings</a>
{:else if !rating}
	<h1>Rating</h1>
	<p>Loading...</p>
{:else}
	<h1>{rating.name}</h1>
	<a href="/ratings">&larr; Back to ratings</a>

	<form onsubmit={handleSave}>
		<label>
			Name
			<input type="text" bind:value={name} required />
		</label>
		<label>
			Description
			<textarea bind:value={description} required></textarea>
		</label>
		<button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save changes'}</button>
	</form>

	<button type="button" onclick={handleDelete} disabled={deleting} class="danger">
		{deleting ? 'Deleting...' : 'Delete rating'}
	</button>

	{#if error}
		<p class="error">{error}</p>
	{/if}
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
	.danger {
		margin-top: 1rem;
		color: #c00;
	}
</style>
