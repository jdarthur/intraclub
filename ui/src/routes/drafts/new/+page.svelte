<script lang="ts">
	import { onMount } from 'svelte';
	import { createDraft, listDraftOrderPatterns, setDraftOrderPattern } from '$lib/draft';
	import type { DraftOrderPattern } from '$lib/draft';
	import { listFormats } from '$lib/format';
	import type { Format } from '$lib/format';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		NativeSelect,
		NativeSelectOption
	} from '$lib/components/ui/native-select/index.js';

	let formats = $state<Format[]>([]);
	let patterns = $state<DraftOrderPattern[]>([]);
	let loadError = $state('');
	let loading = $state(true);

	let name = $state('');
	let format = $state('');
	let pattern = $state('');
	let error = $state('');
	let submitting = $state(false);

	// The backend defaults a new draft's pattern to Snake; only send the
	// explicit pattern set when the user picks something else.
	const DEFAULT_PATTERN = 'Snake';

	onMount(async () => {
		try {
			const [formatList, patternList] = await Promise.all([
				listFormats(),
				listDraftOrderPatterns()
			]);
			formats = formatList;
			patterns = patternList;
			pattern = DEFAULT_PATTERN;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load draft options';
		} finally {
			loading = false;
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		submitting = true;
		try {
			const created = await createDraft({ name: name.trim(), format });
			// If the pattern set fails the draft still exists, so navigate to it
			// rather than stranding the user on the create form. The detail page
			// defaults the pattern to Snake, which is the safe fallback.
			if (pattern && pattern !== DEFAULT_PATTERN) {
				try {
					await setDraftOrderPattern(created.id, pattern);
				} catch {
					// pattern selection failed; proceed to the draft anyway
				}
			}
			await goto(`/drafts/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create draft';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | New draft</title>
</svelte:head>

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">New draft</h1>
	<a href="/drafts" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to drafts</a>
</div>

{#if loading}
	<p class="mt-4 text-muted-foreground">Loading...</p>
{:else if loadError}
	<p class="mt-4 text-sm font-medium text-destructive">{loadError}</p>
{:else}
	<Card class="mt-6 max-w-md">
		<CardHeader>
			<CardTitle class="text-base">Draft details</CardTitle>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSubmit} class="flex flex-col gap-4">
				<div class="flex flex-col gap-2">
					<Label for="name">Name</Label>
					<Input id="name" type="text" bind:value={name} required />
				</div>
				<div class="flex flex-col gap-2">
					<Label for="format">Format</Label>
					<NativeSelect id="format" bind:value={format} required class="w-full">
						<NativeSelectOption value="" disabled>Select a format…</NativeSelectOption>
						{#each formats as f}
							<NativeSelectOption value={f.id}>{f.name}</NativeSelectOption>
						{/each}
					</NativeSelect>
				</div>
				<div class="flex flex-col gap-2">
					<Label for="pattern">Draft order pattern</Label>
					<NativeSelect id="pattern" bind:value={pattern} class="w-full">
						{#each patterns as p}
							<NativeSelectOption value={p.name}>{p.name}</NativeSelectOption>
						{/each}
					</NativeSelect>
				</div>
				<Button type="submit" disabled={submitting} class="w-fit">
					{submitting ? 'Creating...' : 'Create draft'}
				</Button>
			</form>
		</CardContent>
	</Card>
{/if}

{#if error}
	<p class="mt-4 text-sm font-medium text-destructive">{error}</p>
{/if}
