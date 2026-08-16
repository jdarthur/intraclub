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
	import { listSeasons } from '$lib/season';
	import type { Season } from '$lib/season';
	import { listFacilities } from '$lib/facility';
	import { listDrafts } from '$lib/draft';
	import { listPlayoffStructures } from '$lib/playoffStructure';
	import CalendarIcon from '@lucide/svelte/icons/calendar';

	let facilities = $state<Record<string, string>>({});
	let drafts = $state<Record<string, string>>({});
	let playoffs = $state<Record<string, string>>({});

	const seasons = new Async<Season[]>();
	onMount(() =>
		seasons.run(async () => {
			const [seasonList, facilityList, draftList, playoffList] = await Promise.all([
				listSeasons(),
				listFacilities(),
				listDrafts(),
				listPlayoffStructures()
			]);
			facilities = Object.fromEntries(facilityList.map((f) => [f.id, f.name]));
			drafts = Object.fromEntries(draftList.map((d) => [d.id, d.name]));
			playoffs = Object.fromEntries(playoffList.map((p) => [p.id, playoffLabel(p)]));
			return seasonList;
		})
	);

	function playoffLabel(p: { byes: number; number_of_teams: number }): string {
		if (p.number_of_teams === 0 && p.byes === 0) return '';
		return `${p.number_of_teams} teams${p.byes > 0 ? ` / ${p.byes} bye${p.byes > 1 ? 's' : ''}` : ''}`;
	}

	const columns: Column<Season>[] = [
		{ key: 'name', header: 'Name', sortable: true, cell: nameCell },
		{
			key: 'facility',
			header: 'Facility',
			hideBelow: 'sm',
			value: (s) => facilities[s.facility] ?? s.facility
		},
		{ key: 'start_time', header: 'Start time', hideBelow: 'sm', value: (s) => s.start_time },
		{
			key: 'draft',
			header: 'Draft',
			hideBelow: 'md',
			value: (s) => drafts[s.draft_id] ?? s.draft_id
		},
		{
			key: 'playoff',
			header: 'Playoff structure',
			hideBelow: 'md',
			value: (s) => playoffs[s.playoff_structure] ?? s.playoff_structure
		}
	];
</script>

{#snippet nameCell(s: Season)}
	<a href={`/seasons/${s.id}`} class="font-medium text-primary underline-offset-4 hover:underline">
		{s.name}
	</a>
{/snippet}

<svelte:head>
	<title>Intraclub | Seasons</title>
</svelte:head>

<PageHeader title="Seasons" description="Weekly head-to-head play, draft to playoffs" icon={CalendarIcon} />

<AsyncSection state={seasons} isEmpty={(s) => s.length === 0}>
	{#snippet loading()}
		<DataTable rows={[]} {columns} getKey={(s) => s.id} loading />
	{/snippet}
	{#snippet empty()}
		<EmptyState
			title="No seasons yet."
			description="Complete a draft to create one."
		/>
	{/snippet}
	{#snippet children(ss)}
		<DataTable
			rows={ss}
			{columns}
			getKey={(s) => s.id}
			caption="Seasons"
			filter
			filterLabel="Filter seasons"
		/>
	{/snippet}
</AsyncSection>
