<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { getRuleset, amendRulesetName, deleteRuleset } from '$lib/ruleset';
	import type { Ruleset } from '$lib/ruleset';

	// Derived from the route param so the page re-loads when navigating between
	// ruleset revisions (same [id] route, different id).
	const currentId = $derived(page.params.id as string);

	let ruleset = $state<Ruleset | null>(null);
	let loadError = $state('');
	let name = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);

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
		if (!confirm('Delete this ruleset?')) return;
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
	<h1>Ruleset</h1>
	<p class="error">{loadError}</p>
	<a href="/rulesets">&larr; Back to rulesets</a>
{:else if !ruleset}
	<h1>Ruleset</h1>
	<p>Loading...</p>
{:else}
	<h1>{ruleset.name}</h1>
	<a href="/rulesets">&larr; Back to rulesets</a>

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

	<h2>Edit</h2>
	<p class="hint">Saving edits amends this ruleset and creates a new revision.</p>
	<form onsubmit={handleSave}>
		<label>
			Name
			<input type="text" bind:value={name} required />
		</label>
		<button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Save changes'}</button>
	</form>

	<button type="button" onclick={handleDelete} disabled={deleting} class="danger">
		{deleting ? 'Deleting...' : 'Delete ruleset'}
	</button>

	{#if error}
		<p class="error">{error}</p>
	{/if}
{/if}

<style>
	.error {
		color: #c00;
	}
	.hint {
		color: #666;
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
	form {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		max-width: 24rem;
		margin-top: 0.5rem;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	input {
		padding: 0.35rem;
	}
	.danger {
		margin-top: 1rem;
		color: #c00;
	}
</style>
