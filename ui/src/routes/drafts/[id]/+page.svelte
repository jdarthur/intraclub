<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getDraft } from '$lib/draft';
	import type { Draft } from '$lib/draft';

	const id = () => page.params.id as string;

	let draft = $state<Draft | null>(null);
	let error = $state('');

	onMount(async () => {
		try {
			draft = await getDraft(id());
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load draft';
		}
	});
</script>

<svelte:head>
	<title>{draft ? draft.name : 'Draft'}</title>
</svelte:head>

{#if error}
	<h1 class="text-2xl font-semibold tracking-tight">Draft</h1>
	<p class="text-sm font-medium text-destructive">{error}</p>
	<a href="/drafts" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to drafts</a>
{:else if !draft}
	<h1 class="text-2xl font-semibold tracking-tight">Draft</h1>
	<p class="text-muted-foreground">Loading...</p>
{:else}
	<div class="flex items-center gap-4">
		<h1 class="text-2xl font-semibold tracking-tight">{draft.name}</h1>
		<a href="/drafts" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to drafts</a>
	</div>

	<p class="mt-4 text-muted-foreground">
		This draft's setup is not yet available. Check back after the draft setup feature ships.
	</p>
{/if}
