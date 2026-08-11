<script lang="ts">
	import { onMount } from 'svelte';
	import { listRatings } from '$lib/rating';
	import type { Rating } from '$lib/rating';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let ratings = $state<Rating[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			ratings = await listRatings();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load ratings';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Ratings</title>
</svelte:head>

<div class="flex items-center justify-between gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Ratings</h1>
	<Button href="/ratings/new">New rating</Button>
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{:else if ratings.length === 0}
	<p class="text-muted-foreground">No ratings yet.</p>
{:else}
	<div class="mt-4 overflow-hidden rounded-lg border">
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead>Name</TableHead>
					<TableHead>Description</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each ratings as rating}
					<TableRow>
						<TableCell>
							<a href={`/ratings/${rating.id}`} class="font-medium text-primary underline-offset-4 hover:underline">
								{rating.name}
							</a>
						</TableCell>
						<TableCell>{rating.description}</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
