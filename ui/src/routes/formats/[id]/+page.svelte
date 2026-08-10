<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		getFormat,
		updateFormat,
		deleteFormat,
		getFormatPossibleRatings,
		setFormatPossibleRatings
	} from '$lib/format';
	import type { Format } from '$lib/format';
	import { listRatings } from '$lib/rating';
	import type { Rating } from '$lib/rating';

	const id = () => page.params.id as string;

	let format = $state<Format | null>(null);
	let loadError = $state('');
	let name = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);

	// Possible ratings assigned to this format.
	let possibleRatings = $state<Rating[]>([]);
	let allRatings = $state<Rating[]>([]);
	let selectedRatingId = $state('');
	let ratingsError = $state('');
	let ratingsSaving = $state(false);

	onMount(load);

	async function load() {
		loadError = '';
		try {
			const f = await getFormat(id());
			format = f;
			name = f.name;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load format';
			return;
		}
		await loadRatings();
	}

	async function loadRatings() {
		try {
			possibleRatings = await getFormatPossibleRatings(id());
		} catch (e) {
			ratingsError = e instanceof Error ? e.message : 'Failed to load possible ratings';
		}
		// Refresh the full catalog so the add dropdown reflects current assignments.
		try {
			allRatings = await listRatings();
		} catch {
			// ignore; the add dropdown just stays empty
		}
	}

	const unassignedRatings = $derived(
		allRatings.filter((r) => !possibleRatings.some((p) => p.id === r.id))
	);

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

	async function saveRatings(ratingIds: string[]) {
		ratingsError = '';
		ratingsSaving = true;
		try {
			possibleRatings = await setFormatPossibleRatings(id(), ratingIds);
			selectedRatingId = '';
		} catch (err) {
			ratingsError = err instanceof Error ? err.message : 'Failed to update possible ratings';
		} finally {
			ratingsSaving = false;
		}
	}

	function handleAddRating(e: Event) {
		e.preventDefault();
		if (!selectedRatingId) return;
		const next = [...possibleRatings.map((r) => r.id), selectedRatingId];
		saveRatings(next);
	}

	function handleRemoveRating(ratingId: string) {
		const next = possibleRatings.filter((r) => r.id !== ratingId).map((r) => r.id);
		saveRatings(next);
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

	<section class="ratings">
		<h2>Possible ratings</h2>
		<p class="hint">
			The skill ratings that players can be assigned to in this format, ordered
			highest-skill to lowest-skill.
		</p>

		{#if possibleRatings.length === 0}
			<p>No ratings assigned yet.</p>
		{:else}
			<ul class="rating-list">
				{#each possibleRatings as rating}
					<li>
						<span class="rating-name">{rating.name}</span>
						<button
							type="button"
							class="remove"
							onclick={() => handleRemoveRating(rating.id)}
							disabled={ratingsSaving || possibleRatings.length === 1}
							title={possibleRatings.length === 1
								? 'A format must keep at least one rating'
								: undefined}
						>
							Remove
						</button>
					</li>
				{/each}
			</ul>
			{#if possibleRatings.length === 1}
				<p class="hint">A format must keep at least one rating.</p>
			{/if}
		{/if}

		{#if unassignedRatings.length > 0}
			<form class="add" onsubmit={handleAddRating}>
				<select bind:value={selectedRatingId} aria-label="Rating to assign">
					<option value="" disabled>Select a rating…</option>
					{#each unassignedRatings as rating}
						<option value={rating.id}>{rating.name}</option>
					{/each}
				</select>
				<button type="submit" disabled={ratingsSaving || !selectedRatingId}>
					Add rating
				</button>
			</form>
		{:else if allRatings.length > 0}
			<p class="hint">All ratings are already assigned.</p>
		{/if}

		{#if ratingsError}
			<p class="error">{ratingsError}</p>
		{/if}
	</section>

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
	.hint {
		color: #666;
		font-size: 0.9rem;
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
	.ratings {
		margin-top: 1.5rem;
		max-width: 24rem;
	}
	.rating-list {
		list-style: none;
		padding: 0;
		margin: 0.5rem 0;
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}
	.rating-list li {
		display: flex;
		align-items: center;
		justify-content: space-between;
		border: 1px solid #ccc;
		padding: 0.4rem 0.6rem;
	}
	.rating-name {
		font-weight: 600;
	}
	.remove {
		color: #c00;
	}
	.add {
		flex-direction: row;
		align-items: center;
	}
	select {
		padding: 0.35rem;
		flex: 1;
	}
	.danger {
		margin-top: 1.5rem;
		color: #c00;
	}
</style>
