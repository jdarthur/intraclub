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
	import { listScoringStructures, getScoreCountingTypes } from '$lib/scoringStructure';
	import type { ScoringStructure, ScoreCountingType } from '$lib/scoringStructure';
	import { Button } from '$lib/components/ui/button/index.js';
	import GaugeIcon from '@lucide/svelte/icons/gauge';

	let countingTypes = $state<ScoreCountingType[]>([]);
	const structures = new Async<ScoringStructure[]>();
	onMount(() =>
		structures.run(async () => {
			const [ss, cts] = await Promise.all([listScoringStructures(), getScoreCountingTypes()]);
			countingTypes = cts;
			return ss;
		})
	);

	function countingTypeName(type: number): string {
		return countingTypes.find((t) => t.type === type)?.name ?? String(type);
	}

	const columns: Column<ScoringStructure>[] = [
		{ key: 'name', header: 'Name', sortable: true, cell: nameCell },
		{
			key: 'counting_type',
			header: 'Counting type',
			hideBelow: 'sm',
			value: (s) => countingTypeName(s.win_condition_counting_type)
		},
		{
			key: 'win_threshold',
			header: 'Win threshold',
			align: 'right',
			value: (s) => s.win_condition.win_threshold
		}
	];
</script>

{#snippet nameCell(s: ScoringStructure)}
	<a href={`/scoring-structures/${s.id}`} class="font-medium text-primary underline-offset-4 hover:underline dark:text-brand">
		{s.name}
	</a>
{/snippet}

<svelte:head>
	<title>Intraclub | Scoring Structures</title>
</svelte:head>

<PageHeader
	title="Scoring Structures"
	description="How matches are scored and won"
	icon={GaugeIcon}
>
	{#snippet actions()}
		<Button href="/scoring-structures/new">New scoring structure</Button>
	{/snippet}
</PageHeader>

<AsyncSection state={structures} isEmpty={(s) => s.length === 0}>
	{#snippet loading()}
		<DataTable rows={[]} {columns} getKey={(s) => s.id} loading />
	{/snippet}
	{#snippet empty()}
		<EmptyState
			title="No scoring structures yet."
			description="Create a scoring structure to define how matches are won."
		/>
	{/snippet}
	{#snippet children(ss)}
		<DataTable
			rows={ss}
			{columns}
			getKey={(s) => s.id}
			caption="Scoring structures"
			filter
			filterLabel="Filter scoring structures"
		/>
	{/snippet}
</AsyncSection>
