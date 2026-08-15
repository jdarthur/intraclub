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
	import { listFormats } from '$lib/format';
	import type { Format } from '$lib/format';
	import { Button } from '$lib/components/ui/button/index.js';
	import ListTreeIcon from '@lucide/svelte/icons/list-tree';

	const formats = new Async<Format[]>();
	onMount(() => formats.run(() => listFormats()));

	const columns: Column<Format>[] = [
		{ key: 'name', header: 'Name', sortable: true, cell: nameCell }
	];
</script>

{#snippet nameCell(f: Format)}
	<a href={`/formats/${f.id}`} class="font-medium text-primary underline-offset-4 hover:underline">
		{f.name}
	</a>
{/snippet}

<svelte:head>
	<title>Intraclub | Formats</title>
</svelte:head>

<PageHeader title="Formats" description="How a season structures its ratings and play" icon={ListTreeIcon}>
	{#snippet actions()}
		<Button href="/formats/new">New format</Button>
	{/snippet}
</PageHeader>

<AsyncSection state={formats} isEmpty={(f) => f.length === 0}>
	{#snippet loading()}
		<DataTable rows={[]} {columns} getKey={(f) => f.id} loading />
	{/snippet}
	{#snippet empty()}
		<EmptyState title="No formats yet." description="Create a format to define how players are rated and matched." />
	{/snippet}
	{#snippet children(fs)}
		<DataTable
			rows={fs}
			{columns}
			getKey={(f) => f.id}
			caption="Formats"
			filter
			filterLabel="Filter formats"
		/>
	{/snippet}
</AsyncSection>
