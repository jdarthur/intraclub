<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Async } from '$lib/async.svelte';
	import { AsyncSection, PageHeader } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import {
		AlertDialog,
		AlertDialogAction,
		AlertDialogCancel,
		AlertDialogContent,
		AlertDialogDescription,
		AlertDialogFooter,
		AlertDialogHeader,
		AlertDialogTitle,
		AlertDialogTrigger
	} from '$lib/components/ui/alert-dialog/index.js';
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
	import { toast } from '$lib/toast';
	import ImageIcon from '@lucide/svelte/icons/image';

	const id = () => page.params.id as string;

	const photo = new Async<Photo>();
	let altText = $state('');
	let fileType = $state<PhotoType>(0);
	let contents = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);

	onMount(() => photo.run(load));

	async function load(): Promise<Photo> {
		const p = await getPhoto(id());
		altText = p.alt_text;
		fileType = p.file_type;
		contents = p.contents;
		return p;
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
			photo.data = updated;
			altText = updated.alt_text;
			fileType = updated.file_type;
			contents = updated.contents;
			toast.success('Photo saved');
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
			toast.success('Photo deleted');
			await goto('/photos');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete photo';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | Photo</title>
</svelte:head>

<PageHeader title="Photo" icon={ImageIcon} backHref="/photos" backLabel="Back to photos" />

<AsyncSection state={photo}>
	{#snippet children(p)}
		<img
			class="detail mt-2 block max-h-60 max-w-[360px] rounded-md border border-border"
			src={dataUrlFor(p)}
			alt={p.alt_text || 'Photo'}
		/>

		<dl class="my-4 flex gap-6">
			<div>
				<dt class="text-sm text-muted-foreground">Alt text</dt>
				<dd class="m-0">{p.alt_text || '(none)'}</dd>
			</div>
			<div>
				<dt class="text-sm text-muted-foreground">File type</dt>
				<dd class="m-0">{photoTypeLabels[p.file_type] ?? 'unknown'}</dd>
			</div>
			<div>
				<dt class="text-sm text-muted-foreground">Owner</dt>
				<dd class="m-0">{p.owner}</dd>
			</div>
		</dl>

		<Card class="mt-6 max-w-md">
			<CardHeader>
				<CardTitle>Edit photo</CardTitle>
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

		{#if error}
			<Alert variant="destructive" class="mt-4 max-w-md">
				<AlertDescription>{error}</AlertDescription>
			</Alert>
		{/if}

		<div class="mt-4">
			<AlertDialog bind:open={deleteOpen}>
				<AlertDialogTrigger
					disabled={deleting}
					class={buttonVariants({ variant: 'destructive' })}
				>
					{deleting ? 'Deleting...' : 'Delete photo'}
				</AlertDialogTrigger>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Delete photo?</AlertDialogTitle>
						<AlertDialogDescription>
							This permanently removes this photo and cannot be undone.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>Cancel</AlertDialogCancel>
						<AlertDialogAction variant="destructive" onclick={handleDelete}>
							Delete
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</div>
	{/snippet}
</AsyncSection>
