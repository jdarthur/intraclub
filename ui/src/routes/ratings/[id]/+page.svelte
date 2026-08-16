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
	import { getRating, updateRating, deleteRating } from '$lib/rating';
	import type { Rating } from '$lib/rating';
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { toast } from '$lib/toast';
	import StarIcon from '@lucide/svelte/icons/star';

	const id = () => page.params.id as string;

	const rating = new Async<Rating>();
	let name = $state('');
	let description = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);

	onMount(() => rating.run(load));

	async function load(): Promise<Rating> {
		const r = await getRating(id());
		name = r.name;
		description = r.description;
		return r;
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const updated = await updateRating(id(), {
				name: name.trim(),
				description: description.trim()
			});
			rating.data = updated;
			toast.success('Rating saved');
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
			toast.success('Rating deleted');
			await goto('/ratings');
		} catch (err) {
			// e.g. the rating is assigned to a Format and PreDelete blocks it
			error = err instanceof Error ? err.message : 'Failed to delete rating';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | {rating.data ? rating.data.name : 'Rating'}</title>
</svelte:head>

<PageHeader
	title={rating.data?.name}
	icon={StarIcon}
	backHref="/ratings"
	backLabel="Back to ratings"
/>

<AsyncSection state={rating}>
	{#snippet children(r)}
		<Card class="mt-6 max-w-md">
			<CardHeader>
				<CardTitle>Rating details</CardTitle>
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
					{deleting ? 'Deleting...' : 'Delete rating'}
				</AlertDialogTrigger>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Delete rating?</AlertDialogTitle>
						<AlertDialogDescription>
							This permanently removes this rating and cannot be undone.
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
