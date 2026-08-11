<script lang="ts">
	import { createPhoto, photoTypeFromExtension, photoTypeLabels } from '$lib/photo';
	import { goto } from '$app/navigation';
	import type { PhotoType } from '$lib/photo';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { NativeSelect, NativeSelectOption } from '$lib/components/ui/native-select/index.js';

	let altText = $state('');
	let contents = $state('');
	let fileType = $state<PhotoType>(0);
	let error = $state('');
	let submitting = $state(false);

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

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		if (!contents) {
			error = 'Please choose an image file to upload.';
			return;
		}
		submitting = true;
		try {
			const created = await createPhoto({ alt_text: altText, contents, file_type: fileType });
			await goto(`/photos/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create photo';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>New photo</title>
</svelte:head>

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">New photo</h1>
	<a href="/photos" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to photos</a>
</div>

<Card class="mt-6 max-w-md">
	<CardHeader>
		<CardTitle class="text-base">Upload a photo</CardTitle>
	</CardHeader>
	<CardContent>
		<form onsubmit={handleSubmit} class="flex flex-col gap-4">
			<div class="flex flex-col gap-2">
				<Label for="file">File</Label>
				<Input id="file" type="file" accept="image/*" onchange={onFileChange} required />
			</div>
			<div class="flex flex-col gap-2">
				<Label for="altText">Alt text</Label>
				<Input id="altText" type="text" bind:value={altText} />
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
			<Button type="submit" disabled={submitting} class="w-fit">
				{submitting ? 'Creating...' : 'Create photo'}
			</Button>
		</form>
	</CardContent>
</Card>

{#if error}
	<p class="mt-4 text-sm font-medium text-destructive">{error}</p>
{/if}
