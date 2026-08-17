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
	import { listPhotos, dataUrlFor, photoTypeLabels } from '$lib/photo';
	import type { Photo } from '$lib/photo';
	import { Button } from '$lib/components/ui/button/index.js';
	import ImageIcon from '@lucide/svelte/icons/image';

	const photos = new Async<Photo[]>();
	onMount(() => photos.run(() => listPhotos()));

	function photoTypeLabel(type: number): string {
		return photoTypeLabels[type] ?? 'unknown';
	}

	const columns: Column<Photo>[] = [
		{ key: 'thumb', header: 'Thumbnail', cell: thumbCell },
		{ key: 'alt', header: 'Alt text', sortable: true, cell: altCell },
		{ key: 'type', header: 'Type', value: (p) => photoTypeLabel(p.file_type) }
	];
</script>

{#snippet thumbCell(p: Photo)}
	<a href={`/photos/${p.id}`}>
		<img
			class="block size-12 rounded object-cover"
			src={dataUrlFor(p)}
			alt={p.alt_text || 'Photo thumbnail'}
		/>
	</a>
{/snippet}

{#snippet altCell(p: Photo)}
	<a
		href={`/photos/${p.id}`}
		class="font-medium text-primary underline-offset-4 hover:underline dark:text-brand"
	>
		{p.alt_text || '(no alt text)'}
	</a>
{/snippet}

<svelte:head>
	<title>Intraclub | Photos</title>
</svelte:head>

<PageHeader title="Photos" description="Images uploaded to the league" icon={ImageIcon}>
	{#snippet actions()}
		<Button href="/photos/new">New photo</Button>
	{/snippet}
</PageHeader>

<AsyncSection state={photos} isEmpty={(p) => p.length === 0}>
	{#snippet loading()}
		<DataTable rows={[]} {columns} getKey={(p) => p.id} loading />
	{/snippet}
	{#snippet empty()}
		<EmptyState title="No photos yet." description="Upload a photo to share with the league." />
	{/snippet}
	{#snippet children(ps)}
		<DataTable
			rows={ps}
			{columns}
			getKey={(p) => p.id}
			caption="Photos"
			filter
			filterLabel="Filter photos"
		/>
	{/snippet}
</AsyncSection>
