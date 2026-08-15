<script lang="ts">
	import { createFacility } from '$lib/facility';
	import { goto } from '$app/navigation';
	import { PageHeader } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { toast } from '$lib/toast';
	import Building2Icon from '@lucide/svelte/icons/building-2';

	let name = $state('');
	let address = $state('');
	let courts = $state(1);
	let layoutPhoto = $state('');
	let error = $state('');
	let submitting = $state(false);
	let attempted = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		attempted = true;
		error = '';
		submitting = true;
		try {
			const created = await createFacility({
				name: name.trim(),
				address: address.trim(),
				courts,
				layout_photo: layoutPhoto.trim()
			});
			toast.success('Facility created');
			await goto(`/facilities/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create facility';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | New facility</title>
</svelte:head>

<PageHeader title="New facility" icon={Building2Icon} backHref="/facilities" backLabel="Back to facilities" />

<Card class="mt-6 max-w-md">
	<CardHeader>
		<CardTitle>Facility details</CardTitle>
	</CardHeader>
	<CardContent>
		{#if error}
			<Alert variant="destructive" class="mb-4">
				<AlertDescription>{error}</AlertDescription>
			</Alert>
		{/if}
		<form onsubmit={handleSubmit} class="flex flex-col gap-4">
			<div class="flex flex-col gap-2">
				<Label for="name">Name</Label>
				<Input
					id="name"
					type="text"
					bind:value={name}
					required
					aria-invalid={attempted && !name.trim()}
				/>
			</div>
			<div class="flex flex-col gap-2">
				<Label for="address">Address</Label>
				<Input
					id="address"
					type="text"
					bind:value={address}
					required
					aria-invalid={attempted && !address.trim()}
				/>
			</div>
			<div class="flex flex-col gap-2">
				<Label for="courts">Number of courts</Label>
				<Input
					id="courts"
					type="number"
					bind:value={courts}
					min="1"
					required
					aria-invalid={attempted && !courts}
				/>
			</div>
			<div class="flex flex-col gap-2">
				<Label for="layoutPhoto">Layout photo ID (optional)</Label>
				<Input
					id="layoutPhoto"
					type="text"
					bind:value={layoutPhoto}
					placeholder="16-char hex ID"
				/>
			</div>
			<Button type="submit" disabled={submitting} class="w-fit">
				{submitting ? 'Creating...' : 'Create facility'}
			</Button>
		</form>
	</CardContent>
</Card>
