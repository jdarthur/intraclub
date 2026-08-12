<script lang="ts">
	import { createPlayoffStructure } from '$lib/playoffStructure';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';

	let byes = $state<string>('0');
	let numberOfTeams = $state<string>('8');
	let error = $state('');
	let submitting = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		submitting = true;
		try {
			const created = await createPlayoffStructure({
				byes: parseInt(byes || '0', 10),
				number_of_teams: parseInt(numberOfTeams, 10)
			});
			await goto(`/playoff-structures/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create playoff structure';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | New playoff structure</title>
</svelte:head>

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">New playoff structure</h1>
	<a
		href="/playoff-structures"
		class="text-sm text-muted-foreground hover:text-foreground"
	>&larr; Back to playoff structures</a>
</div>

<Card class="mt-6 max-w-md">
	<CardHeader>
		<CardTitle class="text-base">Playoff structure details</CardTitle>
	</CardHeader>
	<CardContent>
		<form onsubmit={handleSubmit} class="flex flex-col gap-4">
			<div class="flex flex-col gap-2">
				<Label for="byes">Byes</Label>
				<Input id="byes" type="number" min="0" bind:value={byes} required />
			</div>
			<div class="flex flex-col gap-2">
				<Label for="numberOfTeams">Number of teams</Label>
				<Input id="numberOfTeams" type="number" min="2" bind:value={numberOfTeams} required />
			</div>
			<Button type="submit" disabled={submitting} class="w-fit">
				{submitting ? 'Creating...' : 'Create playoff structure'}
			</Button>
		</form>
	</CardContent>
</Card>

{#if error}
	<p class="mt-4 text-sm font-medium text-destructive">{error}</p>
{/if}
