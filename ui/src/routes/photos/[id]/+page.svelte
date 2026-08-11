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
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { NativeSelect, NativeSelectOption } from '$lib/components/ui/native-select/index.js';
	import {
		Popover,
		PopoverClose,
		PopoverContent,
		PopoverHeader,
		PopoverTitle,
		PopoverTrigger
	} from '$lib/components/ui/popover/index.js';

	const id = () => page.params.id as string;

	let photo = $state<Photo | null>(null);
	let loadError = $state('');
	let altText = $state('');
	let fileType = $state<PhotoType>(0);
	let contents = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);

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
		deleteOpen = false;
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
	<h1 class="text-2xl font-semibold tracking-tight">Photo</h1>
	<p class="text-sm font-medium text-destructive">{loadError}</p>
	<a href="/photos" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to photos</a>
{:else if !photo}
	<h1 class="text-2xl font-semibold tracking-tight">Photo</h1>
	<p class="text-muted-foreground">Loading...</p>
{:else}
	<div class="flex items-center gap-4">
		<h1 class="text-2xl font-semibold tracking-tight">Photo</h1>
		<a href="/photos" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to photos</a>
	</div>

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

	<Card class="mt-6 max-w-md">
		<CardHeader>
			<CardTitle class="text-base">Edit photo</CardTitle>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSave} class="flex flex-col gap-4">
				<div class="flex flex-col gap-2">
					<Label for="altText">Alt text</Label>
					<Input id="altText" type="text" bind:value={altText} />
				</div>
				<div class="flex flex-col gap-2">
					<Label for="file">File</Label>
					<Input id="file" type="file" accept="image/*" onchange={onFileChange} />
				</div>
				<div class="flex flex-col gap-2">
					<Label for="fileType">File type</Label>
					<NativeSelect id="fileType" bind:value={fileType} class="w-full">
						<NativeSelectOption value={0}>png</NativeSelectOption>
						<NativeSelectOption value={1}>jpg</NativeSelectOption>
						<NativeSelectOption value={2}>jpeg</NativeSelectOption>
						<NativeSelectOption value={3}>gif</NativeSelectOption>
						<NativeSelectOption value={4}>webp</NativeSelectOption>
					</NativeSelect>
				</div>
				<Button type="submit" disabled={saving} class="w-fit">
					{saving ? 'Saving...' : 'Save changes'}
				</Button>
			</form>
		</CardContent>
	</Card>

	<div class="mt-4">
		<Popover bind:open={deleteOpen}>
			<PopoverTrigger disabled={deleting} class={buttonVariants({ variant: 'destructive' })}>
				{deleting ? 'Deleting...' : 'Delete photo'}
			</PopoverTrigger>
			<PopoverContent class="w-80">
				<PopoverHeader>
					<PopoverTitle>Delete photo?</PopoverTitle>
					<p class="text-sm text-muted-foreground">
						This permanently removes this photo and cannot be undone.
					</p>
				</PopoverHeader>
				<div class="flex justify-end gap-2">
					<PopoverClose class={buttonVariants({ variant: 'outline', size: 'sm' })}>Cancel</PopoverClose>
					<Button variant="destructive" size="sm" onclick={handleDelete}>Delete</Button>
				</div>
			</PopoverContent>
		</Popover>
	</div>

	{#if error}
		<p class="mt-4 text-sm font-medium text-destructive">{error}</p>
	{/if}
{/if}

<style>
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
</style>
