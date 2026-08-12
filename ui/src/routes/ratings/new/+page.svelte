<script lang="ts">
	import { createRating } from '$lib/rating';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';

	let name = $state('');
	let description = $state('');
	let error = $state('');
	let submitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		submitting = true;
		try {
			const created = await createRating({ name: name.trim(), description: description.trim() });
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

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">New rating</h1>
	<a href="/ratings" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to ratings</a>
</div>

<Card class="mt-6 max-w-md">
	<CardHeader>
		<CardTitle class="text-base">Rating details</CardTitle>
	</CardHeader>
	<CardContent>
		<form onsubmit={handleSubmit} class="flex flex-col gap-4">
			<div class="flex flex-col gap-2">
				<Label for="name">Name</Label>
				<Input id="name" type="text" bind:value={name} required />
			</div>
			<div class="flex flex-col gap-2">
				<Label for="description">Description</Label>
				<Textarea id="description" bind:value={description} required />
			</div>
			<Button type="submit" disabled={submitting} class="w-fit">
				{submitting ? 'Creating...' : 'Create rating'}
			</Button>
		</form>
	</CardContent>
</Card>

{#if error}
	<p class="mt-4 text-sm font-medium text-destructive">{error}</p>
{/if}
