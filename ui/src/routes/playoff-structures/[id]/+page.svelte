<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		getPlayoffStructure,
		updatePlayoffStructure,
		deletePlayoffStructure
	} from '$lib/playoffStructure';
	import type { PlayoffStructure } from '$lib/playoffStructure';
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

	const id = () => page.params.id as string;

	let structure = $state<PlayoffStructure | null>(null);
	let loadError = $state('');
	let byes = $state<string>('');
	let numberOfTeams = $state<string>('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);

	onMount(load);

	async function load() {
		loadError = '';
		try {
			const s = await getPlayoffStructure(id());
			structure = s;
			byes = String(s.byes);
			numberOfTeams = String(s.number_of_teams);
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load playoff structure';
		}
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const updated = await updatePlayoffStructure(id(), {
				byes: parseInt(byes || '0', 10),
				number_of_teams: parseInt(numberOfTeams, 10)
			});
			structure = updated;
			byes = String(updated.byes);
			numberOfTeams = String(updated.number_of_teams);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update playoff structure';
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		deleteOpen = false;
		error = '';
		deleting = true;
		try {
			await deletePlayoffStructure(id());
			await goto('/playoff-structures');
		} catch (err) {
			// e.g. the structure is referenced by a Season and PreDelete blocks it
			error = err instanceof Error ? err.message : 'Failed to delete playoff structure';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | Playoff Structure</title>
</svelte:head>

{#if loadError}
	<h1 class="text-2xl font-semibold tracking-tight">Playoff Structure</h1>
	<p class="text-sm font-medium text-destructive">{loadError}</p>
	<a
		href="/playoff-structures"
		class="text-sm text-muted-foreground hover:text-foreground"
	>&larr; Back to playoff structures</a>
{:else if !structure}
	<h1 class="text-2xl font-semibold tracking-tight">Playoff Structure</h1>
	<p class="text-muted-foreground">Loading...</p>
{:else}
	<div class="flex items-center gap-4">
		<h1 class="text-2xl font-semibold tracking-tight">Playoff Structure</h1>
		<a
			href="/playoff-structures"
			class="text-sm text-muted-foreground hover:text-foreground"
		>&larr; Back to playoff structures</a>
	</div>

	<Card class="mt-6 max-w-md">
		<CardHeader>
			<CardTitle>Playoff structure details</CardTitle>
		</CardHeader>
		<CardContent>
			<form onsubmit={handleSave} class="flex flex-col gap-4">
				<div class="flex flex-col gap-2">
					<Label for="byes">Byes</Label>
					<Input id="byes" type="number" min="0" bind:value={byes} required />
				</div>
				<div class="flex flex-col gap-2">
					<Label for="numberOfTeams">Number of teams</Label>
					<Input id="numberOfTeams" type="number" min="2" bind:value={numberOfTeams} required />
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
				{deleting ? 'Deleting...' : 'Delete playoff structure'}
			</PopoverTrigger>
			<PopoverContent class="w-80">
				<PopoverHeader>
					<PopoverTitle>Delete playoff structure?</PopoverTitle>
					<p class="text-sm text-muted-foreground">
						This permanently removes this playoff structure and cannot be undone.
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
