<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { getRuleset, amendRulesetName, deleteRuleset } from '$lib/ruleset';
	import type { Ruleset } from '$lib/ruleset';
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		Popover,
		PopoverClose,
		PopoverContent,
		PopoverHeader,
		PopoverTitle,
		PopoverTrigger
	} from '$lib/components/ui/popover/index.js';

	// Derived from the route param so the page re-loads when navigating between
	// ruleset revisions (same [id] route, different id).
	const currentId = $derived(page.params.id as string);

	let ruleset = $state<Ruleset | null>(null);
	let loadError = $state('');
	let name = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);

	$effect(() => {
		load(currentId);
	});

	async function load(idToLoad: string) {
		loadError = '';
		ruleset = null;
		try {
			const r = await getRuleset(idToLoad);
			ruleset = r;
			name = r.name;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load ruleset';
		}
	}

	// Editing a Ruleset's name amends it: the backend creates a new revision
	// that supersedes this one, so we navigate to the new revision.
	async function handleSave(e: Event) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const amended = await amendRulesetName(currentId, name.trim());
			await goto(`/rulesets/${amended.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update ruleset';
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		deleteOpen = false;
		error = '';
		deleting = true;
		try {
			await deleteRuleset(currentId);
			await goto('/rulesets');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete ruleset';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>{ruleset ? ruleset.name : 'Ruleset'}</title>
</svelte:head>

{#if loadError}
	<h1 class="text-2xl font-semibold tracking-tight">Ruleset</h1>
	<p class="text-sm font-medium text-destructive">{loadError}</p>
	<a href="/rulesets" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to rulesets</a>
{:else if !ruleset}
	<h1 class="text-2xl font-semibold tracking-tight">Ruleset</h1>
	<p class="text-muted-foreground">Loading...</p>
{:else}
	<div class="flex items-center gap-4">
		<h1 class="text-2xl font-semibold tracking-tight">{ruleset.name}</h1>
		<a href="/rulesets" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to rulesets</a>
	</div>

	<dl class="meta">
		<div>
			<dt>Revision</dt>
			<dd>{ruleset.revision}</dd>
		</div>
		<div>
			<dt>Date</dt>
			<dd>{new Date(ruleset.date).toLocaleString()}</dd>
		</div>
		<div>
			<dt>Superseded by</dt>
			<dd>
				{#if ruleset.superseded_by}
					<a href={`/rulesets/${ruleset.superseded_by}`}>{ruleset.superseded_by}</a>
				{:else}
					— (current revision)
				{/if}
			</dd>
		</div>
	</dl>

	<Card class="mt-6 max-w-md">
		<CardHeader>
			<CardTitle class="text-base">Edit</CardTitle>
		</CardHeader>
		<CardContent>
			<p class="hint text-sm text-muted-foreground">
				Saving edits amends this ruleset and creates a new revision.
			</p>
			<form onsubmit={handleSave} class="flex flex-col gap-4">
				<div class="flex flex-col gap-2">
					<Label for="name">Name</Label>
					<Input id="name" type="text" bind:value={name} required />
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
				{deleting ? 'Deleting...' : 'Delete ruleset'}
			</PopoverTrigger>
			<PopoverContent class="w-80">
				<PopoverHeader>
					<PopoverTitle>Delete ruleset?</PopoverTitle>
					<p class="text-sm text-muted-foreground">
						This permanently removes this ruleset and cannot be undone.
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
