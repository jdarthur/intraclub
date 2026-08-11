<script lang="ts">
	import { createFacility } from '$lib/facility';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';

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

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">New facility</h1>
	<a href="/facilities" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to facilities</a>
</div>

<Card class="mt-6 max-w-md">
	<CardHeader>
		<CardTitle class="text-base">Facility details</CardTitle>
	</CardHeader>
	<CardContent>
		<form onsubmit={handleSubmit} class="flex flex-col gap-4">
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
			<Button type="submit" disabled={submitting} class="w-fit">
				{submitting ? 'Creating...' : 'Create facility'}
			</Button>
		</form>
	</CardContent>
</Card>

{#if error}
	<p class="mt-4 text-sm font-medium text-destructive">{error}</p>
{/if}
