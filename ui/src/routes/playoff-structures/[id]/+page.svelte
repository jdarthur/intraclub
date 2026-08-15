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
		getPlayoffStructure,
		updatePlayoffStructure,
		deletePlayoffStructure
	} from '$lib/playoffStructure';
	import type { PlayoffStructure } from '$lib/playoffStructure';
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { toast } from '$lib/toast';
	import TrophyIcon from '@lucide/svelte/icons/trophy';

	const id = () => page.params.id as string;

	const structure = new Async<PlayoffStructure>();
	let byes = $state<string>('');
	let numberOfTeams = $state<string>('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);

	onMount(() => structure.run(load));

	async function load(): Promise<PlayoffStructure> {
		const s = await getPlayoffStructure(id());
		byes = String(s.byes);
		numberOfTeams = String(s.number_of_teams);
		return s;
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
			structure.data = updated;
			byes = String(updated.byes);
			numberOfTeams = String(updated.number_of_teams);
			toast.success('Playoff structure saved');
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
			toast.success('Playoff structure deleted');
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

<PageHeader
	title="Playoff Structure"
	icon={TrophyIcon}
	backHref="/playoff-structures"
	backLabel="Back to playoff structures"
/>

<AsyncSection state={structure}>
	{#snippet children(s)}
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
						<Input
							id="numberOfTeams"
							type="number"
							min="2"
							bind:value={numberOfTeams}
							required
						/>
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
					{deleting ? 'Deleting...' : 'Delete playoff structure'}
				</AlertDialogTrigger>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>Delete playoff structure?</AlertDialogTitle>
						<AlertDialogDescription>
							This permanently removes this playoff structure and cannot be undone.
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
