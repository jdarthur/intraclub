<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		getRuleset,
		amendRulesetName,
		deleteRuleset,
		getRulesetSections,
		amendSections,
		type Ruleset,
		type RuleSection,
		RULE_AMENDMENT_ADD,
		RULE_AMENDMENT_REMOVE,
		RULE_AMENDMENT_MODIFY,
		RULE_AMENDMENT_REORDER
	} from '$lib/ruleset';
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
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

	// Section management state.
	let sections = $state<RuleSection[]>([]);
	let sectionError = $state('');
	let sectionSaving = $state(false);
	let addTitle = $state('');
	let addMarkdown = $state('');
	let editingId = $state<string | null>(null);
	let editTitle = $state('');
	let editMarkdown = $state('');

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
			await loadSections(idToLoad);
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load ruleset';
		}
	}

	async function loadSections(idToLoad: string) {
		sectionError = '';
		try {
			sections = await getRulesetSections(idToLoad);
		} catch (e) {
			sectionError = e instanceof Error ? e.message : 'Failed to load sections';
		}
	}

	// After an amend, the ruleset may have been superseded by a new revision
	// (add/remove/reorder) or updated in place (content edit). Navigate to a
	// new revision when the ID changed, otherwise just reload the sections.
	async function afterAmend(amended: Ruleset) {
		if (amended.id !== currentId) {
			await goto(`/rulesets/${amended.id}`);
		} else {
			await loadSections(currentId);
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

	async function handleAdd(e: Event) {
		e.preventDefault();
		sectionError = '';
		sectionSaving = true;
		try {
			// Fetch the current sections so we append after the authoritative
			// last section rather than a possibly-stale in-memory list: the
			// previous add amends the ruleset into a new revision and navigates,
			// and `sections` may not have caught up to the new revision yet.
			const current = await getRulesetSections(currentId);
			const lastId = current.length ? current[current.length - 1].section_id : '';
			const amended = await amendSections(currentId, {
				type: RULE_AMENDMENT_ADD,
				new_section: { title: addTitle, markdown: addMarkdown },
				after: lastId
			});
			addTitle = '';
			addMarkdown = '';
			await afterAmend(amended);
		} catch (err) {
			sectionError = err instanceof Error ? err.message : 'Failed to add section';
		} finally {
			sectionSaving = false;
		}
	}

	function startEdit(section: RuleSection) {
		editingId = section.section_id;
		editTitle = section.title;
		editMarkdown = section.markdown;
	}

	function cancelEdit() {
		editingId = null;
	}

	async function handleEditSave() {
		if (editingId === null) return;
		sectionError = '';
		sectionSaving = true;
		try {
			const amended = await amendSections(currentId, {
				type: RULE_AMENDMENT_MODIFY,
				target_section: editingId,
				new_section: { title: editTitle, markdown: editMarkdown }
			});
			editingId = null;
			await afterAmend(amended);
		} catch (err) {
			sectionError = err instanceof Error ? err.message : 'Failed to update section';
		} finally {
			sectionSaving = false;
		}
	}

	async function handleRemove(section: RuleSection) {
		sectionError = '';
		sectionSaving = true;
		try {
			const amended = await amendSections(currentId, {
				type: RULE_AMENDMENT_REMOVE,
				target_section: section.section_id
			});
			await afterAmend(amended);
		} catch (err) {
			sectionError = err instanceof Error ? err.message : 'Failed to remove section';
		} finally {
			sectionSaving = false;
		}
	}

	async function doReorder(section: RuleSection, after: string) {
		sectionError = '';
		sectionSaving = true;
		try {
			const amended = await amendSections(currentId, {
				type: RULE_AMENDMENT_REORDER,
				target_section: section.section_id,
				after
			});
			await afterAmend(amended);
		} catch (err) {
			sectionError = err instanceof Error ? err.message : 'Failed to reorder section';
		} finally {
			sectionSaving = false;
		}
	}

	// Move a section up one position. Moving to the front uses an empty "after".
	function moveUp(section: RuleSection) {
		const i = sections.findIndex((s) => s.section_id === section.section_id);
		if (i <= 0) return;
		const after = i - 1 === 0 ? '' : sections[i - 2].section_id;
		return doReorder(section, after);
	}

	function moveDown(section: RuleSection) {
		const i = sections.findIndex((s) => s.section_id === section.section_id);
		if (i < 0 || i >= sections.length - 1) return;
		return doReorder(section, sections[i + 1].section_id);
	}
</script>

<svelte:head>
	<title>Intraclub | {ruleset ? ruleset.name : 'Ruleset'}</title>
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
			<CardTitle>Edit</CardTitle>
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

	<Card class="mt-6">
		<CardHeader>
			<CardTitle>Sections</CardTitle>
		</CardHeader>
		<CardContent class="flex flex-col gap-4">
			{#if sections.length === 0}
				<p class="text-sm text-muted-foreground">This ruleset has no sections yet.</p>
			{:else}
				<ol class="flex flex-col gap-3">
					{#each sections as section, i (section.section_id)}
						<li class="section-card">
							<div class="flex items-center justify-between gap-2">
								<span class="font-medium">
									{i + 1}. {section.title || '(untitled)'}
								</span>
								<div class="flex items-center gap-1">
									<Button
										variant="outline"
										size="sm"
										disabled={sectionSaving || i === 0}
										onclick={() => moveUp(section)}
									>
										&uarr;
									</Button>
									<Button
										variant="outline"
										size="sm"
										disabled={sectionSaving || i === sections.length - 1}
										onclick={() => moveDown(section)}
									>
										&darr;
									</Button>
									<Button variant="outline" size="sm" onclick={() => startEdit(section)}>
										Edit
									</Button>
									<Button
										variant="destructive"
										size="sm"
										disabled={sectionSaving}
										onclick={() => handleRemove(section)}
									>
										Remove
									</Button>
								</div>
							</div>

							{#if editingId === section.section_id}
								<form
									onsubmit={(e) => {
										e.preventDefault();
										handleEditSave();
									}}
									class="mt-3 flex flex-col gap-3"
								>
									<div class="flex flex-col gap-2">
										<Label for={`edit-title-${section.section_id}`}>Title</Label>
										<Input
											id={`edit-title-${section.section_id}`}
											type="text"
											bind:value={editTitle}
										/>
									</div>
									<div class="flex flex-col gap-2">
										<Label for={`edit-markdown-${section.section_id}`}>Contents</Label>
										<Textarea
											id={`edit-markdown-${section.section_id}`}
											bind:value={editMarkdown}
											rows={4}
										/>
									</div>
									<div class="flex gap-2">
										<Button type="submit" size="sm" disabled={sectionSaving}>
											Save section
										</Button>
										<Button variant="outline" size="sm" type="button" onclick={cancelEdit}>
											Cancel
										</Button>
									</div>
								</form>
							{:else}
								{#if section.markdown}
									<p class="section-contents">{section.markdown}</p>
								{/if}
							{/if}
						</li>
					{/each}
				</ol>
			{/if}

			<form onsubmit={handleAdd} class="flex flex-col gap-3 border-t pt-4">
				<p class="text-sm font-medium">Add a section</p>
				<div class="flex flex-col gap-2">
					<Label for="add-title">Title</Label>
					<Input id="add-title" type="text" bind:value={addTitle} placeholder="e.g. Eligibility" />
				</div>
				<div class="flex flex-col gap-2">
					<Label for="add-markdown">Contents</Label>
					<Textarea
						id="add-markdown"
						bind:value={addMarkdown}
						rows={4}
						placeholder="Rule text"
						required
					/>
				</div>
				<Button type="submit" class="w-fit" disabled={sectionSaving}>
					{sectionSaving ? 'Saving...' : 'Add section'}
				</Button>
			</form>

			{#if sectionError}
				<p class="text-sm font-medium text-destructive">{sectionError}</p>
			{/if}
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
	.section-card {
		border: 1px solid var(--border);
		border-radius: var(--radius);
		padding: 0.75rem;
	}
	.section-contents {
		margin-top: 0.5rem;
		white-space: pre-wrap;
		font-size: 0.875rem;
		color: #555;
	}
</style>
