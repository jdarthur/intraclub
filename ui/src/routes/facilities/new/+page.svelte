<script lang="ts">
	import { createFacility } from '$lib/facility';
	import { goto } from '$app/navigation';

	let name = $state('');
	let address = $state('');
	let courts = $state(1);
	let layoutPhoto = $state('');
	let error = $state('');
	let submitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		submitting = true;
		try {
			const created = await createFacility({
				name: name.trim(),
				address: address.trim(),
				courts,
				layout_photo: layoutPhoto.trim()
			});
			await goto(`/facilities/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create facility';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>New facility</title>
</svelte:head>

<h1>New facility</h1>
<a href="/facilities">&larr; Back to facilities</a>

<form onsubmit={handleSubmit}>
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
	<button type="submit" disabled={submitting}>{submitting ? 'Creating...' : 'Create facility'}</button>
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
