<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { getFacility, updateFacility, deleteFacility } from '$lib/facility';
	import type { Facility } from '$lib/facility';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';

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
	<h1 class="text-2xl font-semibold tracking-tight">Facility</h1>
	<p class="text-sm font-medium text-destructive">{loadError}</p>
	<a href="/facilities" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to facilities</a>
{:else if !facility}
	<h1 class="text-2xl font-semibold tracking-tight">Facility</h1>
	<p class="text-muted-foreground">Loading...</p>
{:else}
	<div class="flex items-center gap-4">
		<h1 class="text-2xl font-semibold tracking-tight">{facility.name}</h1>
		<a href="/facilities" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to facilities</a>
	</div>

	<Card class="mt-6 max-w-md">
		<CardHeader>
			<CardTitle class="text-base">Facility details</CardTitle>
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
					<Label for="layoutPhoto">Layout photo ID (optional)</Label>
					<Input id="layoutPhoto" type="text" bind:value={layoutPhoto} placeholder="16-char hex ID" />
				</div>
				<Button type="submit" disabled={saving} class="w-fit">
					{saving ? 'Saving...' : 'Save changes'}
				</Button>
			</form>
		</CardContent>
	</Card>

	<div class="mt-4">
		<Button variant="destructive" onclick={handleDelete} disabled={deleting}>
			{deleting ? 'Deleting...' : 'Delete facility'}
		</Button>
	</div>

	{#if error}
		<p class="mt-4 text-sm font-medium text-destructive">{error}</p>
	{/if}
{/if}
