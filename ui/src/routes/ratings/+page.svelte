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
	import { listRatings } from '$lib/rating';
	import type { Rating } from '$lib/rating';
	import { Button } from '$lib/components/ui/button/index.js';
	import StarIcon from '@lucide/svelte/icons/star';

	const ratings = new Async<Rating[]>();
	onMount(() => ratings.run(() => listRatings()));

	const columns: Column<Rating>[] = [
		{ key: 'name', header: 'Name', sortable: true, cell: nameCell },
		{ key: 'description', header: 'Description', hideBelow: 'sm', value: (r) => r.description }
	];
</script>

{#snippet nameCell(r: Rating)}
	<a href={`/ratings/${r.id}`} class="font-medium text-primary underline-offset-4 hover:underline dark:text-brand">
		{r.name}
	</a>
{/snippet}

<svelte:head>
	<title>Intraclub | Ratings</title>
</svelte:head>

<PageHeader title="Ratings" description="The skill levels players can be assigned" icon={StarIcon}>
	{#snippet actions()}
		<Button href="/ratings/new">New rating</Button>
	{/snippet}
</PageHeader>

<AsyncSection state={ratings} isEmpty={(r) => r.length === 0}>
	{#snippet loading()}
		<DataTable rows={[]} {columns} getKey={(r) => r.id} loading />
	{/snippet}
	{#snippet empty()}
		<EmptyState title="No ratings yet." description="Create ratings so formats can define player skill levels." />
	{/snippet}
	{#snippet children(rs)}
		<DataTable
			rows={rs}
			{columns}
			getKey={(r) => r.id}
			caption="Ratings"
			filter
			filterLabel="Filter ratings"
		/>
	{/snippet}
</AsyncSection>
