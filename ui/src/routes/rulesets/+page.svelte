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
	import { listRulesets } from '$lib/ruleset';
	import type { Ruleset } from '$lib/ruleset';
	import { Button } from '$lib/components/ui/button/index.js';
	import ScrollTextIcon from '@lucide/svelte/icons/scroll-text';

	const rulesets = new Async<Ruleset[]>();
	onMount(() => rulesets.run(() => listRulesets()));

	const columns: Column<Ruleset>[] = [
		{ key: 'name', header: 'Name', sortable: true, cell: nameCell },
		{ key: 'revision', header: 'Revision', value: (r) => r.revision }
	];
</script>

{#snippet nameCell(r: Ruleset)}
	<a href={`/rulesets/${r.id}`} class="font-medium text-primary underline-offset-4 hover:underline">
		{r.name}
	</a>
{/snippet}

<svelte:head>
	<title>Intraclub | Rulesets</title>
</svelte:head>

<PageHeader title="Rulesets" description="The club's house rules, versioned by amendment" icon={ScrollTextIcon}>
	{#snippet actions()}
		<Button href="/rulesets/new">New ruleset</Button>
	{/snippet}
</PageHeader>

<AsyncSection state={rulesets} isEmpty={(r) => r.length === 0}>
	{#snippet loading()}
		<DataTable rows={[]} {columns} getKey={(r) => r.id} loading />
	{/snippet}
	{#snippet empty()}
		<EmptyState title="No rulesets yet." description="Create a ruleset to codify the club's rules." />
	{/snippet}
	{#snippet children(rs)}
		<DataTable
			rows={rs}
			{columns}
			getKey={(r) => r.id}
			caption="Rulesets"
			filter
			filterLabel="Filter rulesets"
		/>
	{/snippet}
</AsyncSection>
