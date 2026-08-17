<script lang="ts">
	import { onMount } from 'svelte';
	import { Async } from '$lib/async.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import {
		Popover,
		PopoverContent,
		PopoverTrigger
	} from '$lib/components/ui/popover/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { createPhoto, dataUrlFor, listPhotos, photoTypeFromExtension, photoTypeLabels } from '$lib/photo';
	import type { Photo, PhotoType } from '$lib/photo';
	import { toast } from '$lib/toast';
	import { cn } from '$lib/utils.js';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
	import ImageIcon from '@lucide/svelte/icons/image';
	import ImagePlusIcon from '@lucide/svelte/icons/image-plus';
	import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle';
	import UploadIcon from '@lucide/svelte/icons/upload';
	import XIcon from '@lucide/svelte/icons/x';

	let {
		value = $bindable(''),
		label = 'Photo'
	}: {
		/** Currently selected photo id (16-char hex string; '' = none). */
		value?: string;
		/** Accessible name for the trigger, e.g. "Layout photo". */
		label?: string;
	} = $props();

	const photos = new Async<Photo[]>();
	let open = $state(false);

	// Inline upload state (upload a photo without leaving the form).
	let uploadAlt = $state('');
	let uploadContents = $state('');
	let uploadType = $state<PhotoType | null>(null);
	let uploading = $state(false);
	let uploadError = $state('');

	const selected = $derived(photos.data?.find((p) => p.id === value) ?? null);

	onMount(() => photos.run(() => listPhotos()));

	function selectPhoto(p: Photo) {
		value = p.id;
		open = false;
	}

	function clear() {
		value = '';
		open = false;
	}

	function onFileChange(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		const type = photoTypeFromExtension(file.name);
		if (type === null) {
			uploadError = `Unsupported file type. Use one of: ${Object.values(photoTypeLabels).join(', ')}.`;
			uploadType = null;
			uploadContents = '';
			return;
		}
		uploadType = type;
		uploadError = '';
		const reader = new FileReader();
		reader.onload = () => {
			const result = reader.result as string;
			uploadContents = result.split(',')[1] ?? '';
		};
		reader.readAsDataURL(file);
	}

	async function handleUpload() {
		uploadError = '';
		if (!uploadContents || uploadType === null) {
			uploadError = 'Choose an image file to upload first.';
			return;
		}
		uploading = true;
		try {
			const created = await createPhoto({
				alt_text: uploadAlt,
				contents: uploadContents,
				file_type: uploadType
			});
			toast.success('Photo uploaded');
			// Refresh the list so the new photo appears in the grid, then select it.
			await photos.run(() => listPhotos());
			value = created.id;
			open = false;
			uploadAlt = '';
			uploadContents = '';
			uploadType = null;
		} catch (err) {
			uploadError = err instanceof Error ? err.message : 'Failed to upload photo';
		} finally {
			uploading = false;
		}
	}
</script>

<Popover bind:open>
	<PopoverTrigger
		type="button"
		aria-label={`Select ${label}`}
		class="flex h-8 w-full min-w-0 items-center gap-2 rounded-lg border border-input bg-transparent px-2.5 text-sm transition-colors select-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring data-[state=open]:border-ring aria-expanded:bg-muted dark:bg-input/30 dark:hover:bg-input/50 dark:aria-expanded:bg-input/50 outline-none disabled:pointer-events-none disabled:opacity-50"
	>
		{#if selected}
			<img
				class="size-6 shrink-0 rounded object-cover"
				src={dataUrlFor(selected)}
				alt=""
			/>
			<span class="min-w-0 flex-1 truncate text-left">{selected.alt_text || selected.id}</span>
		{:else}
			<ImageIcon class="size-4 shrink-0 text-muted-foreground" aria-hidden />
			<span class="min-w-0 flex-1 truncate text-left text-muted-foreground">
				{label ? `Select ${label.toLowerCase()}…` : 'Select photo…'}
			</span>
		{/if}
		<ChevronsUpDownIcon class="size-4 shrink-0 text-muted-foreground" aria-hidden />
	</PopoverTrigger>

	<PopoverContent class="w-80" align="start">
		{#if photos.status === 'error'}
			<p class="px-1 py-2 text-sm text-destructive">Failed to load photos: {photos.error}</p>
		{:else if photos.status === 'loading'}
			<div class="grid grid-cols-3 gap-2">
				{#each Array(6) as _, i}
					<div class="flex flex-col gap-1">
						<Skeleton class="h-16 w-full rounded-md" />
						<Skeleton class="h-3 w-full" />
					</div>
				{/each}
			</div>
		{:else if photos.data && photos.data.length > 0}
			<div class="flex max-h-64 flex-col gap-1 overflow-y-auto pr-1">
				<div class="grid grid-cols-3 gap-2">
					{#each photos.data as p (p.id)}
						<button
							type="button"
							title={p.alt_text || p.id}
							onclick={() => selectPhoto(p)}
							class={cn(
								'group flex flex-col items-center gap-1 rounded-md p-1 text-left transition-colors outline-none hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring',
								p.id === value && 'bg-muted'
							)}
						>
							<span
								class={cn(
									'relative block w-full overflow-hidden rounded-md ring-1 ring-foreground/10',
									p.id === value && 'ring-2 ring-primary dark:ring-brand'
								)}
							>
								<img
									class="h-16 w-full object-cover"
									src={dataUrlFor(p)}
									alt={p.alt_text || 'Photo'}
								/>
								{#if p.id === value}
									<span
										class="absolute top-1 right-1 flex size-4 items-center justify-center rounded-full bg-primary text-primary-foreground"
									>
										<CheckIcon class="size-3" aria-hidden />
									</span>
								{/if}
							</span>
							<span class="w-full truncate text-xs text-muted-foreground group-hover:text-foreground">
								{p.alt_text || 'No alt text'}
							</span>
						</button>
					{/each}
				</div>
			</div>
			{#if value}
				<button
					type="button"
					onclick={clear}
					class="mt-1 flex w-full items-center justify-center gap-1 rounded-md px-2 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring"
				>
					<XIcon class="size-4" aria-hidden />
					Clear selection
				</button>
			{/if}
		{:else}
			<p class="px-1 py-2 text-sm text-muted-foreground">
				No photos yet — upload one below to attach it.
			</p>
		{/if}

		<Separator class="my-2" />

		<div class="flex flex-col gap-2">
			<p class="flex items-center gap-1.5 px-1 text-sm font-medium">
				<ImagePlusIcon class="size-4" aria-hidden />
				Upload a new photo
			</p>
			<Input id="photoPickerFile" type="file" accept="image/*" onchange={onFileChange} />
			<Input
				type="text"
				bind:value={uploadAlt}
				placeholder="Alt text (optional)"
				aria-label="Alt text"
			/>
			{#if uploadError}
				<p class="px-1 text-sm text-destructive" role="alert">{uploadError}</p>
			{/if}
			<Button
				type="button"
				size="sm"
				class="w-fit"
				onclick={handleUpload}
				disabled={uploading || !uploadContents}
			>
				{#if uploading}
					<LoaderCircleIcon class="size-4 animate-spin" aria-hidden />
					Uploading…
				{:else}
					<UploadIcon class="size-4" aria-hidden />
					Upload
				{/if}
			</Button>
		</div>
	</PopoverContent>
</Popover>
