<script lang="ts">
	import { createRating } from '$lib/rating';
	import { goto } from '$app/navigation';
	import { PageHeader } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { toast } from '$lib/toast';
	import StarIcon from '@lucide/svelte/icons/star';

	let name = $state('');
	let description = $state('');
	let error = $state('');
	let submitting = $state(false);
	let attempted = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		attempted = true;
		error = '';
		submitting = true;
		try {
			const created = await createRating({ name: name.trim(), description: description.trim() });
			toast.success('Rating created');
			await goto(`/ratings/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create rating';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | New rating</title>
</svelte:head>

<PageHeader title="New rating" icon={StarIcon} backHref="/ratings" backLabel="Back to ratings" />

<Card class="mt-6 max-w-md">
	<CardHeader>
		<CardTitle>Rating details</CardTitle>
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
				<Label for="description">Description</Label>
				<Textarea
					id="description"
					bind:value={description}
					required
					aria-invalid={attempted && !description.trim()}
				/>
			</div>
			<Button type="submit" disabled={submitting} class="w-fit">
				{submitting ? 'Creating...' : 'Create rating'}
			</Button>
		</form>
	</CardContent>
</Card>
