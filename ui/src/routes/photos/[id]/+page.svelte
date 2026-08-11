<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		getPhoto,
		updatePhoto,
		deletePhoto,
		dataUrlFor,
		photoTypeFromExtension,
		photoTypeLabels,
		type Photo,
		type PhotoType
	} from '$lib/photo';

	const id = () => page.params.id as string;

	let photo = $state<Photo | null>(null);
	let loadError = $state('');
	let altText = $state('');
	let fileType = $state<PhotoType>(0);
	let contents = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);

	onMount(load);

	async function load() {
		loadError = '';
		try {
			const p = await getPhoto(id());
			photo = p;
			altText = p.alt_text;
			fileType = p.file_type;
			contents = p.contents;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load photo';
		}
	}

	function onFileChange(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		const type = photoTypeFromExtension(file.name);
		if (type === null) {
			error = `Unsupported file type. Use one of: ${Object.values(photoTypeLabels).join(', ')}.`;
			return;
		}
		fileType = type;
		error = '';
		const reader = new FileReader();
		reader.onload = () => {
			const result = reader.result as string;
			contents = result.split(',')[1] ?? '';
		};
		reader.readAsDataURL(file);
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const updated = await updatePhoto(id(), {
				alt_text: altText,
				contents,
				file_type: fileType
			});
			photo = updated;
			altText = updated.alt_text;
			fileType = updated.file_type;
			contents = updated.contents;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update photo';
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		if (!confirm('Delete this photo?')) return;
		error = '';
		deleting = true;
		try {
			await deletePhoto(id());
			await goto('/photos');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete photo';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>Photo</title>
</svelte:head>

{#if loadError}
	<h1>Photo</h1>
	<p class="error">{loadError}</p>
	<a href="/photos">&larr; Back to photos</a>
{:else if !photo}
	<h1>Photo</h1>
	<p>Loading...</p>
{:else}
	<h1>Photo</h1>
	<a href="/photos">&larr; Back to photos</a>

	<img class="detail" src={dataUrlFor(photo)} alt={photo.alt_text || 'Photo'} />

	<dl class="meta">
		<div>
			<dt>Alt text</dt>
			<dd>{photo.alt_text || '(none)'}</dd>
		</div>
		<div>
			<dt>File type</dt>
			<dd>{photoTypeLabels[photo.file_type] ?? 'unknown'}</dd>
		</div>
		<div>
			<dt>Owner</dt>
			<dd>{photo.owner}</dd>
		</div>
	</dl>

	<h2>Edit</h2>
	<form onsubmit={handleSave}>
		<label>
			Alt text
			<input type="text" bind:value={altText} />
		</label>
		<label>
			File
			<input type="file" accept="image/*" onchange={onFileChange} />
		</label>
		<label>
			File type
			<select bind:value={fileType}>
				<option value={0}>png</option>
				<option value={1}>jpg</option>
				<option value={2}>jpeg</option>
				<option value={3}>gif</option>
				<option value={4}>webp</option>
			</select>
		</label>
		<button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save changes'}</button>
	</form>

	<button type="button" onclick={handleDelete} disabled={deleting} class="danger">
		{deleting ? 'Deleting...' : 'Delete photo'}
	</button>

	{#if error}
		<p class="error">{error}</p>
	{/if}
{/if}

<style>
	.error {
		color: #c00;
	}
	.detail {
		max-width: 360px;
		max-height: 240px;
		border: 1px solid #ccc;
		border-radius: 6px;
		margin-top: 0.5rem;
	}
	.meta {
		display: flex;
		gap: 1.5rem;
		margin: 1rem 0;
	}
	.meta dt {
		font-size: 0.85rem;
		color: #666;
	}
	.meta dd {
		margin: 0;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		max-width: 24rem;
		margin-top: 0.5rem;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	input,
	select {
		padding: 0.35rem;
	}
	.danger {
		margin-top: 1rem;
		color: #c00;
	}
</style>
