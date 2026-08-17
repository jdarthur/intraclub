<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Async } from '$lib/async.svelte';
	import { AsyncSection, PageHeader, PhotoPicker } from '$lib/components/app/index.js';
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
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { toast } from '$lib/toast';
	import { getFacility, updateFacility, deleteFacility } from '$lib/facility';
	import type { Facility } from '$lib/facility';
	import Building2Icon from '@lucide/svelte/icons/building-2';

	const id = () => page.params.id as string;

	const facility = new Async<Facility>();
	let name = $state('');
	let address = $state('');
	let courts = $state(1);
	let layoutPhoto = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);

	onMount(() => facility.run(load));

	async function load(): Promise<Facility> {
		const f = await getFacility(id());
		name = f.name;
		address = f.address;
		courts = f.courts;
		layoutPhoto = f.layout_photo;
		return f;
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
			facility.data = updated;
			toast.success('Facility saved');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update facility';
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		deleteOpen = false;
		error = '';
		deleting = true;
		try {
			await deleteFacility(id());
			toast.success('Facility deleted');
			await goto('/facilities');
		} catch (err) {
			// e.g. the facility is assigned to a Season and PreDelete blocks it
			error = err instanceof Error ? err.message : 'Failed to delete facility';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | {facility.data ? facility.data.name : 'Facility'}</title>
</svelte:head>

<PageHeader
	title={facility.data?.name}
	icon={Building2Icon}
	backHref="/facilities"
	backLabel="Back to facilities"
/>

<AsyncSection state={facility}>
	{#snippet children(f)}
		<Card class="mt-6 max-w-md">
			<CardHeader>
				<CardTitle>Facility details</CardTitle>
			</CardHeader>
			<CardContent>
				<form onsubmit={handleSave} class="flex flex-col gap-4">
					<div class="flex flex-col gap-2">
						<Label for="name">Name</Label>
						<Input id="name" type="text" bind:value={name} required />
					</div>
					<div class="flex flex-col gap-2">
						<Label for="address">Address</Label>
						<Input id="address" type="text" bind:value={address} required />
					</div>
					<div class="flex flex-col gap-2">
						<Label for="courts">Number of courts</Label>
						<Input id="courts" type="number" bind:value={courts} min="1" required />
					</div>
					<div class="flex flex-col gap-2">
						<Label>Layout photo (optional)</Label>
						<PhotoPicker label="Layout photo" bind:value={layoutPhoto} />
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

		<div class="mt-4">
			<AlertDialog bind:open={deleteOpen}>
				<AlertDialogTrigger
					disabled={deleting}
					class={buttonVariants({ variant: 'destructive' })}
				>
					{deleting ? 'Deleting...' : 'Delete facility'}
				</AlertDialogTrigger>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Delete facility?</AlertDialogTitle>
						<AlertDialogDescription>
							This permanently removes this facility and cannot be undone.
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
