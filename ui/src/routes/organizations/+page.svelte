<script lang="ts">
	import { onMount } from 'svelte';
	import { Async } from '$lib/async.svelte';
	import {
		AsyncSection,
		DataTable,
		EmptyState,
		PageHeader
	} from '$lib/components/app/index.js';
	import type { Column } from '$lib/components/app/data-table.svelte';
	import { listOrganizations } from '$lib/organization';
	import type { Organization } from '$lib/organization';
	import { Button } from '$lib/components/ui/button/index.js';
	import BuildingIcon from '@lucide/svelte/icons/building';

	const organizations = new Async<Organization[]>();
	onMount(() => organizations.run(() => listOrganizations()));

	const columns: Column<Organization>[] = [
		{ key: 'name', header: 'Name', sortable: true, cell: nameCell }
	];
</script>

{#snippet nameCell(o: Organization)}
	<a href={`/organizations/${o.id}`} class="font-medium text-primary underline-offset-4 hover:underline dark:text-brand">
		{o.name}
	</a>
{/snippet}

<svelte:head>
	<title>Intraclub | Organizations</title>
</svelte:head>

<PageHeader title="Organizations" description="Clubs and groups that play in the league" icon={BuildingIcon}>
	{#snippet actions()}
		<Button href="/organizations/new">New organization</Button>
	{/snippet}
</PageHeader>

<AsyncSection state={organizations} isEmpty={(o) => o.length === 0}>
	{#snippet loading()}
		<DataTable rows={[]} {columns} getKey={(o) => o.id} loading />
	{/snippet}
	{#snippet empty()}
		<EmptyState title="No organizations yet." description="Create an organization to group members and manage rosters." />
	{/snippet}
	{#snippet children(os)}
		<DataTable
			rows={os}
			{columns}
			getKey={(o) => o.id}
			caption="Organizations"
			filter
			filterLabel="Filter organizations"
		/>
	{/snippet}
</AsyncSection>
