<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { getRating, updateRating, deleteRating } from '$lib/rating';
	import type { Rating } from '$lib/rating';
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import {
		Popover,
		PopoverClose,
		PopoverContent,
		PopoverHeader,
		PopoverTitle,
		PopoverTrigger
	} from '$lib/components/ui/popover/index.js';

	const id = () => page.params.id as string;

	let rating = $state<Rating | null>(null);
	let loadError = $state('');
	let name = $state('');
	let description = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);

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
		deleteOpen = false;
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
	<title>Intraclub | {rating ? rating.name : 'Rating'}</title>
</svelte:head>

{#if loadError}
	<h1 class="text-2xl font-semibold tracking-tight">Rating</h1>
	<p class="text-sm font-medium text-destructive">{loadError}</p>
	<a href="/ratings" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to ratings</a>
{:else if !rating}
	<h1 class="text-2xl font-semibold tracking-tight">Rating</h1>
	<p class="text-muted-foreground">Loading...</p>
{:else}
	<div class="flex items-center gap-4">
		<h1 class="text-2xl font-semibold tracking-tight">{rating.name}</h1>
		<a href="/ratings" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to ratings</a>
	</div>

	<Card class="mt-6 max-w-md">
		<CardHeader>
			<CardTitle class="text-base">Rating details</CardTitle>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSave} class="flex flex-col gap-4">
				<div class="flex flex-col gap-2">
					<Label for="name">Name</Label>
					<Input id="name" type="text" bind:value={name} required />
				</div>
				<div class="flex flex-col gap-2">
					<Label for="description">Description</Label>
					<Textarea id="description" bind:value={description} required />
				</div>
				<Button type="submit" disabled={saving} class="w-fit">
					{saving ? 'Saving...' : 'Save changes'}
				</Button>
			</form>
		</CardContent>
	</Card>

	<div class="mt-8">
		<Popover bind:open={deleteOpen}>
			<PopoverTrigger disabled={deleting} class={buttonVariants({ variant: 'destructive' })}>
				{deleting ? 'Deleting...' : 'Delete rating'}
			</PopoverTrigger>
			<PopoverContent class="w-80">
				<PopoverHeader>
					<PopoverTitle>Delete rating?</PopoverTitle>
					<p class="text-sm text-muted-foreground">
						This permanently removes this rating and cannot be undone.
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
