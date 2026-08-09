<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { getFormat, updateFormat, deleteFormat } from '$lib/format';
	import type { Format } from '$lib/format';

	const id = () => page.params.id as string;

	let format = $state<Format | null>(null);
	let loadError = $state('');
	let name = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);

	onMount(load);

	async function load() {
		loadError = '';
		try {
			const f = await getFormat(id());
			format = f;
			name = f.name;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load format';
		}
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const updated = await updateFormat(id(), { name: name.trim() });
			format = updated;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update format';
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		if (!confirm('Delete this format?')) return;
		error = '';
		deleting = true;
		try {
			await deleteFormat(id());
			await goto('/formats');
		} catch (err) {
			// e.g. the format is assigned to a Draft and PreDelete blocks it
			error = err instanceof Error ? err.message : 'Failed to delete format';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>{format ? format.name : 'Format'}</title>
</svelte:head>

{#if loadError}
	<h1>Format</h1>
	<p class="error">{loadError}</p>
	<a href="/formats">&larr; Back to formats</a>
{:else if !format}
	<h1>Format</h1>
	<p>Loading...</p>
{:else}
	<h1>{format.name}</h1>
	<a href="/formats">&larr; Back to formats</a>

	<form onsubmit={handleSave}>
		<label>
			Name
			<input type="text" bind:value={name} required />
		</label>
		<button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save changes'}</button>
	</form>

	<button type="button" onclick={handleDelete} disabled={deleting} class="danger">
		{deleting ? 'Deleting...' : 'Delete format'}
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
	input {
		padding: 0.35rem;
	}
	.danger {
		margin-top: 1rem;
		color: #c00;
	}
</style>
