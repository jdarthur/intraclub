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
	import { listTeams } from '$lib/team';
	import type { TeamRoster } from '$lib/team';
	import ShieldIcon from '@lucide/svelte/icons/shield';

	const rosters = new Async<TeamRoster[]>();
	onMount(() => rosters.run(() => listTeams()));

	const columns: Column<TeamRoster>[] = [
		{ key: 'name', header: 'Team', sortable: true, cell: nameCell },
		{
			key: 'members',
			header: 'Members',
			align: 'right',
			value: (r) => r.assignments.length,
			class: 'text-muted-foreground'
		}
	];
</script>

{#snippet nameCell(r: TeamRoster)}
	<a href={`/teams/${r.team.id}`} class="font-medium text-primary underline-offset-4 hover:underline">
		{r.team.name}
	</a>
{/snippet}

<svelte:head>
	<title>Intraclub | Teams</title>
</svelte:head>

<PageHeader title="Teams" description="Rosters formed by finalized drafts" icon={ShieldIcon} />

<AsyncSection state={rosters} isEmpty={(r) => r.length === 0}>
	{#snippet loading()}
		<DataTable rows={[]} {columns} getKey={(r) => r.team.id} loading />
	{/snippet}
	{#snippet empty()}
		<EmptyState
			title="You aren't on any teams."
			description="Teams are created when a draft is finalized and are only visible to their members."
		/>
	{/snippet}
	{#snippet children(rs)}
		<DataTable
			rows={rs}
			{columns}
			getKey={(r) => r.team.id}
			caption="Teams"
			filter
			filterLabel="Filter teams"
		/>
	{/snippet}
</AsyncSection>
