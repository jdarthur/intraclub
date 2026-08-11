<script lang="ts">
	import { onMount } from 'svelte';
	import { listFormats } from '$lib/format';
	import type { Format } from '$lib/format';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let formats = $state<Format[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			formats = await listFormats();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load formats';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Formats</title>
</svelte:head>

<div class="flex items-center justify-between gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Formats</h1>
	<Button href="/formats/new">New format</Button>
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{:else if formats.length === 0}
	<p class="text-muted-foreground">No formats yet.</p>
{:else}
	<div class="mt-4 overflow-hidden rounded-lg border">
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead>Name</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each formats as format}
					<TableRow>
						<TableCell>
							<a href={`/formats/${format.id}`} class="font-medium text-primary underline-offset-4 hover:underline">
								{format.name}
							</a>
						</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
