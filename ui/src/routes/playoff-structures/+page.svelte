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
	import { listPlayoffStructures } from '$lib/playoffStructure';
	import type { PlayoffStructure } from '$lib/playoffStructure';
	import { Button } from '$lib/components/ui/button/index.js';
	import TrophyIcon from '@lucide/svelte/icons/trophy';

	const structures = new Async<PlayoffStructure[]>();
	onMount(() => structures.run(() => listPlayoffStructures()));

	const columns: Column<PlayoffStructure>[] = [
		{ key: 'name', header: 'Playoff structure', sortable: true, cell: nameCell }
	];
</script>

{#snippet nameCell(s: PlayoffStructure)}
	<a
		href={`/playoff-structures/${s.id}`}
		class="font-medium text-primary underline-offset-4 hover:underline dark:text-brand"
	>
		{s.byes} byes / {s.number_of_teams} teams
	</a>
{/snippet}

<svelte:head>
	<title>Intraclub | Playoff Structures</title>
</svelte:head>

<PageHeader
	title="Playoff Structures"
	description="Bracket shapes for the end-of-season tournament"
	icon={TrophyIcon}
>
	{#snippet actions()}
		<Button href="/playoff-structures/new">New playoff structure</Button>
	{/snippet}
</PageHeader>

<AsyncSection state={structures} isEmpty={(s) => s.length === 0}>
	{#snippet loading()}
		<DataTable rows={[]} {columns} getKey={(s) => s.id} loading />
	{/snippet}
	{#snippet empty()}
		<EmptyState
			title="No playoff structures yet."
			description="Create a playoff structure to define how the post-season bracket is seeded."
		/>
	{/snippet}
	{#snippet children(ss)}
		<DataTable
			rows={ss}
			{columns}
			getKey={(s) => s.id}
			caption="Playoff structures"
			filter
			filterLabel="Filter playoff structures"
		/>
	{/snippet}
</AsyncSection>
