<script lang="ts">
	import { createOrganization } from '$lib/organization';
	import { goto } from '$app/navigation';
	import { PageHeader } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { toast } from '$lib/toast';
	import BuildingIcon from '@lucide/svelte/icons/building';

	let name = $state('');
	let error = $state('');
	let submitting = $state(false);
	let attempted = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		attempted = true;
		error = '';
		submitting = true;
		try {
			const created = await createOrganization({ name: name.trim() });
			toast.success('Organization created');
			await goto(`/organizations/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create organization';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | New organization</title>
</svelte:head>

<PageHeader
	title="New organization"
	icon={BuildingIcon}
	backHref="/organizations"
	backLabel="Back to organizations"
/>

<Card class="mt-6 max-w-md">
	<CardHeader>
		<CardTitle>Organization details</CardTitle>
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
			<Button type="submit" disabled={submitting} class="w-fit">
				{submitting ? 'Creating...' : 'Create organization'}
			</Button>
		</form>
	</CardContent>
</Card>
