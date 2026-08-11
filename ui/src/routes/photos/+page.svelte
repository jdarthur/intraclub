<script lang="ts">
	import { onMount } from 'svelte';
	import { listPhotos, dataUrlFor, photoTypeLabels } from '$lib/photo';
	import type { Photo } from '$lib/photo';

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

<h1>Photos</h1>
<a href="/photos/new">New photo</a>

{#if loading}
	<p>Loading...</p>
{:else if error}
	<p class="error">{error}</p>
{:else if photos.length === 0}
	<p>No photos yet.</p>
{:else}
	<table>
		<thead>
			<tr>
				<th>Thumbnail</th>
				<th>Alt text</th>
				<th>Type</th>
			</tr>
		</thead>
		<tbody>
			{#each photos as photo}
				<tr>
					<td class="thumb">
						<a href={`/photos/${photo.id}`}>
							<img src={dataUrlFor(photo)} alt={photo.alt_text || 'Photo thumbnail'} />
						</a>
					</td>
					<td><a href={`/photos/${photo.id}`}>{photo.alt_text || '(no alt text)'}</a></td>
					<td>{photoTypeLabel(photo.file_type)}</td>
				</tr>
			{/each}
		</tbody>
	</table>
{/if}

<style>
	.error {
		color: #c00;
	}
	table {
		border-collapse: collapse;
		margin-top: 1rem;
	}
	th,
	td {
		border: 1px solid #ccc;
		padding: 0.4rem 0.8rem;
		text-align: left;
	}
	.thumb img {
		display: block;
		width: 48px;
		height: 48px;
		object-fit: cover;
		border-radius: 4px;
	}
</style>
