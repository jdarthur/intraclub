<script lang="ts">
	import { onMount } from 'svelte';
	import { listFacilities } from '$lib/facility';
	import type { Facility } from '$lib/facility';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let facilities = $state<Facility[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			facilities = await listFacilities();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load facilities';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Intraclub | Facilities</title>
</svelte:head>

<div class="flex items-center justify-between gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Facilities</h1>
	<Button href="/facilities/new">New facility</Button>
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{:else if facilities.length === 0}
	<p class="text-muted-foreground">No facilities yet.</p>
{:else}
	<div class="mt-4 overflow-hidden rounded-lg border">
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead>Name</TableHead>
					<TableHead>Address</TableHead>
					<TableHead>Courts</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each facilities as facility}
					<TableRow>
						<TableCell>
							<a href={`/facilities/${facility.id}`} class="font-medium text-primary underline-offset-4 hover:underline">
								{facility.name}
							</a>
						</TableCell>
						<TableCell>{facility.address}</TableCell>
						<TableCell>{facility.courts}</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
