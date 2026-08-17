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
	import { listFacilities } from '$lib/facility';
	import type { Facility } from '$lib/facility';
	import { Button } from '$lib/components/ui/button/index.js';
	import Building2Icon from '@lucide/svelte/icons/building-2';

	const facilities = new Async<Facility[]>();
	onMount(() => facilities.run(() => listFacilities()));

	const columns: Column<Facility>[] = [
		{ key: 'name', header: 'Name', sortable: true, cell: nameCell },
		{ key: 'address', header: 'Address', hideBelow: 'sm', value: (f) => f.address },
		{ key: 'courts', header: 'Courts', align: 'right', value: (f) => f.courts }
	];
</script>

{#snippet nameCell(f: Facility)}
	<a href={`/facilities/${f.id}`} class="font-medium text-primary underline-offset-4 hover:underline dark:text-brand">
		{f.name}
	</a>
{/snippet}

<svelte:head>
	<title>Intraclub | Facilities</title>
</svelte:head>

<PageHeader title="Facilities" description="Courts and venues available to the league" icon={Building2Icon}>
	{#snippet actions()}
		<Button href="/facilities/new">New facility</Button>
	{/snippet}
</PageHeader>

<AsyncSection state={facilities} isEmpty={(f) => f.length === 0}>
	{#snippet loading()}
		<DataTable rows={[]} {columns} getKey={(f) => f.id} loading />
	{/snippet}
	{#snippet empty()}
		<EmptyState title="No facilities yet." description="Add a venue so seasons have somewhere to play." />
	{/snippet}
	{#snippet children(fs)}
		<DataTable
			rows={fs}
			{columns}
			getKey={(f) => f.id}
			caption="Facilities"
			filter
			filterLabel="Filter facilities"
		/>
	{/snippet}
</AsyncSection>
