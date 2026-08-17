<script lang="ts">
	import { onMount } from 'svelte';
	import { Async } from '$lib/async.svelte';
	import {
		AsyncSection,
		DataTable,
		EmptyState,
		PageHeader,
		StatusBadge
	} from '$lib/components/app/index.js';
	import type { Column } from '$lib/components/app/data-table.svelte';
	import { listDrafts } from '$lib/draft';
	import type { Draft } from '$lib/draft';
	import { listFormats } from '$lib/format';
	import { listUsers, fullName } from '$lib/user';
	import { Button } from '$lib/components/ui/button/index.js';
	import ListOrderedIcon from '@lucide/svelte/icons/list-ordered';

	let formats = $state<Record<string, string>>({});
	let users = $state<Record<string, string>>({});

	const drafts = new Async<Draft[]>();
	onMount(() =>
		drafts.run(async () => {
			const [draftList, formatList, userList] = await Promise.all([
				listDrafts(),
				listFormats(),
				listUsers()
			]);
			formats = Object.fromEntries(formatList.map((f) => [f.id, f.name]));
			users = Object.fromEntries(userList.map((u) => [u.id, fullName(u)]));
			return draftList;
		})
	);

	// The zero Go time marshals to this literal; treat it as "not set".
	const ZERO_TIME = '0001-01-01T00:00:00Z';

	function isCompleted(draft: Draft): boolean {
		return !!draft.completed_at && draft.completed_at !== ZERO_TIME;
	}

	const columns: Column<Draft>[] = [
		{ key: 'name', header: 'Name', sortable: true, cell: nameCell },
		{
			key: 'owner',
			header: 'Owner',
			hideBelow: 'sm',
			value: (d) => users[d.owner] ?? d.owner
		},
		{
			key: 'format',
			header: 'Format',
			hideBelow: 'md',
			value: (d) => formats[d.format] ?? d.format
		},
		{ key: 'status', header: 'Status', cell: statusCell }
	];
</script>

{#snippet nameCell(d: Draft)}
	<a href={`/drafts/${d.id}`} class="font-medium text-primary underline-offset-4 hover:underline dark:text-brand">
		{d.name}
	</a>
{/snippet}

{#snippet statusCell(d: Draft)}
	<StatusBadge
		status={isCompleted(d) ? 'complete' : 'open'}
		label={isCompleted(d) ? 'Completed' : 'In progress'}
	/>
{/snippet}

<svelte:head>
	<title>Intraclub | Drafts</title>
</svelte:head>

<PageHeader title="Drafts" description="Player drafts that build teams for a season" icon={ListOrderedIcon}>
	{#snippet actions()}
		<Button href="/drafts/new">New draft</Button>
	{/snippet}
</PageHeader>

<AsyncSection state={drafts} isEmpty={(d) => d.length === 0}>
	{#snippet loading()}
		<DataTable rows={[]} {columns} getKey={(d) => d.id} loading />
	{/snippet}
	{#snippet empty()}
		<EmptyState title="No drafts yet." description="Create a draft to split players into teams." />
	{/snippet}
	{#snippet children(ds)}
		<DataTable
			rows={ds}
			{columns}
			getKey={(d) => d.id}
			caption="Drafts"
			filter
			filterLabel="Filter drafts"
		/>
	{/snippet}
</AsyncSection>
