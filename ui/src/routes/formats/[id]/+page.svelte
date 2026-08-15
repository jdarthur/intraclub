<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Async } from '$lib/async.svelte';
	import { AsyncSection, PageHeader } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import {
		AlertDialog,
		AlertDialogAction,
		AlertDialogCancel,
		AlertDialogContent,
		AlertDialogDescription,
		AlertDialogFooter,
		AlertDialogHeader,
		AlertDialogTitle,
		AlertDialogTrigger
	} from '$lib/components/ui/alert-dialog/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		NativeSelect,
		NativeSelectOption
	} from '$lib/components/ui/native-select/index.js';
	import { toast } from '$lib/toast';
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
	import ListTreeIcon from '@lucide/svelte/icons/list-tree';

	const id = () => page.params.id as string;

	const format = new Async<Format>();
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

	onMount(() => format.run(load));

	async function load(): Promise<Format> {
		const f = await getFormat(id());
		name = f.name;
		await loadRatings();
		return f;
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
			format.data = updated;
			toast.success('Format saved');
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
			toast.success('Format deleted');
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
	<title>Intraclub | {format.data ? format.data.name : 'Format'}</title>
</svelte:head>

<PageHeader
	title={format.data?.name}
	icon={ListTreeIcon}
	backHref="/formats"
	backLabel="Back to formats"
/>

<AsyncSection state={format}>
	{#snippet children(f)}
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
				<Alert variant="destructive" class="mt-3">
					<AlertDescription>{ratingsError}</AlertDescription>
				</Alert>
			{/if}
		</section>

		{#if error}
			<Alert variant="destructive" class="mt-4 max-w-md">
				<AlertDescription>{error}</AlertDescription>
			</Alert>
		{/if}

		<div class="mt-8">
			<AlertDialog bind:open={deleteOpen}>
				<AlertDialogTrigger
					disabled={deleting}
					class={buttonVariants({ variant: 'destructive' })}
				>
					{deleting ? 'Deleting...' : 'Delete format'}
				</AlertDialogTrigger>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Delete format?</AlertDialogTitle>
						<AlertDialogDescription>
							This permanently removes this format and cannot be undone.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Cancel</AlertDialogCancel>
						<AlertDialogAction variant="destructive" onclick={handleDelete}>
							Delete
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</div>
	{/snippet}
</AsyncSection>
