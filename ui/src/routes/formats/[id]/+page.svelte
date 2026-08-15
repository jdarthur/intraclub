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
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		NativeSelect,
		NativeSelectOption
	} from '$lib/components/ui/native-select/index.js';
	import {
		Popover,
		PopoverClose,
		PopoverContent,
		PopoverHeader,
		PopoverTitle,
		PopoverTrigger
	} from '$lib/components/ui/popover/index.js';

	const id = () => page.params.id as string;

	let format = $state<Format | null>(null);
	let loadError = $state('');
	let name = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);

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
		deleteOpen = false;
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
	<title>Intraclub | {format ? format.name : 'Format'}</title>
</svelte:head>

{#if loadError}
	<h1 class="text-2xl font-semibold tracking-tight">Format</h1>
	<p class="text-sm font-medium text-destructive">{loadError}</p>
	<a href="/formats" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to formats</a>
{:else if !format}
	<h1 class="text-2xl font-semibold tracking-tight">Format</h1>
	<p class="text-muted-foreground">Loading...</p>
{:else}
	<div class="flex items-center gap-4">
		<h1 class="text-2xl font-semibold tracking-tight">{format.name}</h1>
		<a href="/formats" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to formats</a>
	</div>

	<Card class="mt-6 max-w-md">
		<CardHeader>
			<CardTitle>Format details</CardTitle>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSave} class="flex flex-col gap-4">
				<div class="flex flex-col gap-2">
					<Label for="name">Name</Label>
					<Input id="name" type="text" bind:value={name} required />
				</div>
				<Button type="submit" disabled={saving} class="w-fit">
					{saving ? 'Saving...' : 'Save changes'}
				</Button>
			</form>
		</CardContent>
	</Card>

	<section class="mt-8 max-w-md">
		<h2 class="text-xl font-semibold tracking-tight">Possible ratings</h2>
		<p class="mt-1 text-sm text-muted-foreground">
			The skill ratings that players can be assigned to in this format, ordered
			highest-skill to lowest-skill.
		</p>

		{#if possibleRatings.length === 0}
			<p class="mt-4 text-muted-foreground">No ratings assigned yet.</p>
		{:else}
			<ul class="mt-4 flex flex-col gap-2">
				{#each possibleRatings as rating}
					<li class="flex items-center justify-between gap-2 rounded-lg border p-2 pl-3">
						<Badge variant="secondary" class="rating-name">{rating.name}</Badge>
						<Button
							type="button"
							variant="outline"
							size="sm"
							onclick={() => handleRemoveRating(rating.id)}
							disabled={ratingsSaving || possibleRatings.length === 1}
							title={possibleRatings.length === 1
								? 'A format must keep at least one rating'
								: undefined}
						>
							Remove
						</Button>
					</li>
				{/each}
			</ul>
			{#if possibleRatings.length === 1}
				<p class="mt-2 text-sm text-muted-foreground">A format must keep at least one rating.</p>
			{/if}
		{/if}

		{#if unassignedRatings.length > 0}
			<form onsubmit={handleAddRating} class="mt-4 flex items-center gap-2">
				<NativeSelect
					bind:value={selectedRatingId}
					aria-label="Rating to assign"
					class="flex-1"
				>
					<NativeSelectOption value="" disabled>Select a rating…</NativeSelectOption>
					{#each unassignedRatings as rating}
						<NativeSelectOption value={rating.id}>{rating.name}</NativeSelectOption>
					{/each}
				</NativeSelect>
				<Button type="submit" disabled={ratingsSaving || !selectedRatingId}>
					Add rating
				</Button>
			</form>
		{:else if allRatings.length > 0}
			<p class="mt-4 text-sm text-muted-foreground">All ratings are already assigned.</p>
		{/if}

		{#if ratingsError}
			<p class="mt-3 text-sm font-medium text-destructive">{ratingsError}</p>
		{/if}
	</section>

	<div class="mt-8">
		<Popover bind:open={deleteOpen}>
			<PopoverTrigger disabled={deleting} class={buttonVariants({ variant: 'destructive' })}>
				{deleting ? 'Deleting...' : 'Delete format'}
			</PopoverTrigger>
			<PopoverContent class="w-80">
				<PopoverHeader>
					<PopoverTitle>Delete format?</PopoverTitle>
					<p class="text-sm text-muted-foreground">
						This permanently removes this format and cannot be undone.
					</p>
				</PopoverHeader>
				<div class="flex justify-end gap-2">
					<PopoverClose class={buttonVariants({ variant: 'outline', size: 'sm' })}>Cancel</PopoverClose>
					<Button variant="destructive" size="sm" onclick={handleDelete}>Delete</Button>
				</div>
			</PopoverContent>
		</Popover>
	</div>

	{#if error}
		<p class="mt-4 text-sm font-medium text-destructive">{error}</p>
	{/if}
{/if}
