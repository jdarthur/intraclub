<script lang="ts">
	import { createPlayoffStructure } from '$lib/playoffStructure';
	import { goto } from '$app/navigation';
	import { PageHeader } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { toast } from '$lib/toast';
	import TrophyIcon from '@lucide/svelte/icons/trophy';

	let byes = $state<string>('0');
	let numberOfTeams = $state<string>('8');
	let error = $state('');
	let submitting = $state(false);
	let attempted = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		attempted = true;
		error = '';
		submitting = true;
		try {
			const created = await createPlayoffStructure({
				byes: parseInt(byes || '0', 10),
				number_of_teams: parseInt(numberOfTeams, 10)
			});
			toast.success('Playoff structure created');
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

<PageHeader
	title="New playoff structure"
	icon={TrophyIcon}
	backHref="/playoff-structures"
	backLabel="Back to playoff structures"
/>

<Card class="mt-6 max-w-md">
	<CardHeader>
		<CardTitle>Playoff structure details</CardTitle>
	</CardHeader>
	<CardContent>
		{#if error}
			<Alert variant="destructive" class="mb-4">
				<AlertDescription>{error}</AlertDescription>
			</Alert>
		{/if}
		<form onsubmit={handleSubmit} class="flex flex-col gap-4">
			<div class="flex flex-col gap-2">
				<Label for="byes">Byes</Label>
				<Input
					id="byes"
					type="number"
					min="0"
					bind:value={byes}
					required
					aria-invalid={attempted && !byes}
				/>
			</div>
			<div class="flex flex-col gap-2">
				<Label for="numberOfTeams">Number of teams</Label>
				<Input
					id="numberOfTeams"
					type="number"
					min="2"
					bind:value={numberOfTeams}
					required
					aria-invalid={attempted && !numberOfTeams}
				/>
			</div>
			<Button type="submit" disabled={submitting} class="w-fit">
				{submitting ? 'Creating...' : 'Create playoff structure'}
			</Button>
		</form>
	</CardContent>
</Card>
