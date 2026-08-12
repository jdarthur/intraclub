<script lang="ts">
	import { onMount } from 'svelte';
	import { listPlayoffStructures } from '$lib/playoffStructure';
	import type { PlayoffStructure } from '$lib/playoffStructure';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let structures = $state<PlayoffStructure[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			structures = await listPlayoffStructures();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load playoff structures';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Intraclub | Playoff Structures</title>
</svelte:head>

<div class="flex items-center justify-between gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Playoff Structures</h1>
	<Button href="/playoff-structures/new">New playoff structure</Button>
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{:else if structures.length === 0}
	<p class="text-muted-foreground">No playoff structures yet.</p>
{:else}
	<div class="mt-4 overflow-hidden rounded-lg border">
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead>Playoff structure</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each structures as structure}
					<TableRow>
						<TableCell>
							<a
								href={`/playoff-structures/${structure.id}`}
								class="font-medium text-primary underline-offset-4 hover:underline"
							>
								{structure.byes} byes / {structure.number_of_teams} teams
							</a>
						</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
