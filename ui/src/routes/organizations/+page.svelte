<script lang="ts">
	import { onMount } from 'svelte';
	import { listOrganizations } from '$lib/organization';
	import type { Organization } from '$lib/organization';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let organizations = $state<Organization[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			organizations = await listOrganizations();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load organizations';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Intraclub | Organizations</title>
</svelte:head>

<div class="flex items-center justify-between gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Organizations</h1>
	<Button href="/organizations/new">New organization</Button>
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{:else if organizations.length === 0}
	<p class="text-muted-foreground">No organizations yet.</p>
{:else}
	<div class="mt-4 overflow-hidden rounded-lg border">
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead>Name</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each organizations as organization}
					<TableRow>
						<TableCell>
							<a href={`/organizations/${organization.id}`} class="font-medium text-primary underline-offset-4 hover:underline">
								{organization.name}
							</a>
						</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
