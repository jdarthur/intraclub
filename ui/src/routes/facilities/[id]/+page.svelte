<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { getFacility, updateFacility, deleteFacility } from '$lib/facility';
	import type { Facility } from '$lib/facility';

	const id = () => page.params.id as string;

	let facility = $state<Facility | null>(null);
	let loadError = $state('');
	let name = $state('');
	let address = $state('');
	let courts = $state(1);
	let layoutPhoto = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);

	onMount(load);

	async function load() {
		loadError = '';
		try {
			const f = await getFacility(id());
			facility = f;
			name = f.name;
			address = f.address;
			courts = f.courts;
			layoutPhoto = f.layout_photo;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load facility';
		}
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const updated = await updateFacility(id(), {
				name: name.trim(),
				address: address.trim(),
				courts,
				layout_photo: layoutPhoto.trim()
			});
			facility = updated;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update facility';
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		if (!confirm('Delete this facility?')) return;
		error = '';
		deleting = true;
		try {
			await deleteFacility(id());
			await goto('/facilities');
		} catch (err) {
			// e.g. the facility is assigned to a Season and PreDelete blocks it
			error = err instanceof Error ? err.message : 'Failed to delete facility';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>{facility ? facility.name : 'Facility'}</title>
</svelte:head>

{#if loadError}
	<h1>Facility</h1>
	<p class="error">{loadError}</p>
	<a href="/facilities">&larr; Back to facilities</a>
{:else if !facility}
	<h1>Facility</h1>
	<p>Loading...</p>
{:else}
	<h1>{facility.name}</h1>
	<a href="/facilities">&larr; Back to facilities</a>

	<form onsubmit={handleSave}>
		<label>
			Name
			<input type="text" bind:value={name} required />
		</label>
		<label>
			Address
			<input type="text" bind:value={address} required />
		</label>
		<label>
			Number of courts
			<input type="number" bind:value={courts} min="1" required />
		</label>
		<label>
			Layout photo ID (optional)
			<input type="text" bind:value={layoutPhoto} placeholder="16-char hex ID" />
		</label>
		<button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save changes'}</button>
	</form>

	<button type="button" onclick={handleDelete} disabled={deleting} class="danger">
		{deleting ? 'Deleting...' : 'Delete facility'}
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
