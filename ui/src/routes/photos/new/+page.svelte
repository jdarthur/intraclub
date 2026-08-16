<script lang="ts">
	import { createPhoto, photoTypeFromExtension, photoTypeLabels } from '$lib/photo';
	import { goto } from '$app/navigation';
	import type { PhotoType } from '$lib/photo';
	import { PageHeader } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { NativeSelect, NativeSelectOption } from '$lib/components/ui/native-select/index.js';
	import { toast } from '$lib/toast';
	import ImageIcon from '@lucide/svelte/icons/image';

	let altText = $state('');
	let contents = $state('');
	let fileType = $state<PhotoType>(0);
	let error = $state('');
	let submitting = $state(false);
	let attempted = $state(false);

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
		attempted = true;
		error = '';
		if (!contents) {
			error = 'Please choose an image file to upload.';
			return;
		}
		submitting = true;
		try {
			const created = await createPhoto({ alt_text: altText, contents, file_type: fileType });
			toast.success('Photo created');
			await goto(`/photos/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create photo';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | New photo</title>
</svelte:head>

<PageHeader title="New photo" icon={ImageIcon} backHref="/photos" backLabel="Back to photos" />

<Card class="mt-6 max-w-md">
	<CardHeader>
		<CardTitle>Upload a photo</CardTitle>
	</CardHeader>
	<CardContent>
		{#if error}
			<Alert variant="destructive" class="mb-4">
				<AlertDescription>{error}</AlertDescription>
			</Alert>
		{/if}
		<form onsubmit={handleSubmit} class="flex flex-col gap-4">
			<div class="flex flex-col gap-2">
				<Label for="file">File</Label>
				<Input
					id="file"
					type="file"
					accept="image/*"
					onchange={onFileChange}
					required
					aria-invalid={attempted && !contents}
				/>
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
