<script lang="ts">
	import { onMount } from 'svelte';
	import { listPhotos, dataUrlFor, photoTypeLabels } from '$lib/photo';
	import type { Photo } from '$lib/photo';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let photos = $state<Photo[]>([]);
	let loading = $state(true);
	let error = $state('');

	function photoTypeLabel(type: number): string {
		return photoTypeLabels[type] ?? 'unknown';
	}

	onMount(async () => {
		try {
			photos = await listPhotos();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load photos';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Photos</title>
</svelte:head>

<div class="flex items-center justify-between gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Photos</h1>
	<Button href="/photos/new">New photo</Button>
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{:else if photos.length === 0}
	<p class="text-muted-foreground">No photos yet.</p>
{:else}
	<div class="mt-4 overflow-hidden rounded-lg border">
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead>Thumbnail</TableHead>
					<TableHead>Alt text</TableHead>
					<TableHead>Type</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each photos as photo}
					<TableRow>
						<TableCell>
							<a href={`/photos/${photo.id}`}>
								<img class="thumb-img" src={dataUrlFor(photo)} alt={photo.alt_text || 'Photo thumbnail'} />
							</a>
						</TableCell>
						<TableCell>
							<a
								href={`/photos/${photo.id}`}
								class="font-medium text-primary underline-offset-4 hover:underline"
							>
								{photo.alt_text || '(no alt text)'}
							</a>
						</TableCell>
						<TableCell>{photoTypeLabel(photo.file_type)}</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}

<style>
	.thumb-img {
		display: block;
		width: 48px;
		height: 48px;
		object-fit: cover;
		border-radius: 4px;
	}
</style>
