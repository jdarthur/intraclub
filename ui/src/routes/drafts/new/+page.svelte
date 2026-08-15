<script lang="ts">
	import { onMount } from 'svelte';
	import { createDraft, listDraftOrderPatterns, setDraftOrderPattern } from '$lib/draft';
	import type { DraftOrderPattern } from '$lib/draft';
	import { listFormats } from '$lib/format';
	import type { Format } from '$lib/format';
	import { goto } from '$app/navigation';
	import { Async } from '$lib/async.svelte';
	import { AsyncSection, PageHeader } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		NativeSelect,
		NativeSelectOption
	} from '$lib/components/ui/native-select/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { toast } from '$lib/toast';
	import ListOrderedIcon from '@lucide/svelte/icons/list-ordered';

	type DraftOptions = { formats: Format[]; patterns: DraftOrderPattern[] };

	const options = new Async<DraftOptions>();
	let name = $state('');
	let format = $state('');
	let pattern = $state('');
	let error = $state('');
	let submitting = $state(false);
	let attempted = $state(false);

	// The backend defaults a new draft's pattern to Snake; only send the
	// explicit pattern set when the user picks something else.
	const DEFAULT_PATTERN = 'Snake';

	onMount(() =>
		options.run(async () => {
			const [formatList, patternList] = await Promise.all([
				listFormats(),
				listDraftOrderPatterns()
			]);
			pattern = DEFAULT_PATTERN;
			return { formats: formatList, patterns: patternList };
		})
	);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		attempted = true;
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
			toast.success('Draft created');
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

<PageHeader title="New draft" icon={ListOrderedIcon} backHref="/drafts" backLabel="Back to drafts" />

<AsyncSection state={options}>
	{#snippet loading()}
		<Skeleton class="mt-6 h-64 w-full max-w-md" />
	{/snippet}
	{#snippet children(opts)}
		<Card class="mt-6 max-w-md">
			<CardHeader>
				<CardTitle>Draft details</CardTitle>
			</CardHeader>
			<CardContent>
				{#if error}
					<Alert variant="destructive" class="mb-4">
						<AlertDescription>{error}</AlertDescription>
					</Alert>
				{/if}
				<form onsubmit={handleSubmit} class="flex flex-col gap-4">
					<div class="flex flex-col gap-2">
						<Label for="name">Name</Label>
						<Input
							id="name"
							type="text"
							bind:value={name}
							required
							aria-invalid={attempted && !name.trim()}
						/>
					</div>
					<div class="flex flex-col gap-2">
						<Label for="format">Format</Label>
						<NativeSelect
							id="format"
							bind:value={format}
							required
							class="w-full"
							aria-invalid={attempted && !format}
						>
							<NativeSelectOption value="" disabled>Select a format…</NativeSelectOption>
							{#each opts.formats as f}
								<NativeSelectOption value={f.id}>{f.name}</NativeSelectOption>
							{/each}
						</NativeSelect>
					</div>
					<div class="flex flex-col gap-2">
						<Label for="pattern">Draft order pattern</Label>
						<NativeSelect id="pattern" bind:value={pattern} class="w-full">
							{#each opts.patterns as p}
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
	{/snippet}
</AsyncSection>
