<script lang="ts">
	import { createFormat } from '$lib/format';
	import { goto } from '$app/navigation';
	import { PageHeader } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { toast } from '$lib/toast';
	import ListTreeIcon from '@lucide/svelte/icons/list-tree';

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
			const created = await createFormat({ name: name.trim() });
			toast.success('Format created');
			await goto(`/formats/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create format';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | New format</title>
</svelte:head>

<PageHeader title="New format" icon={ListTreeIcon} backHref="/formats" backLabel="Back to formats" />

<Card class="mt-6 max-w-md">
	<CardHeader>
		<CardTitle>Format details</CardTitle>
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
				{submitting ? 'Creating...' : 'Create format'}
			</Button>
		</form>
	</CardContent>
</Card>
