<script lang="ts">
	import { onMount } from 'svelte';
	import { createScoringStructure, getScoreCountingTypes } from '$lib/scoringStructure';
	import type { ScoreCountingType } from '$lib/scoringStructure';
	import { goto } from '$app/navigation';
	import { PageHeader } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { NativeSelect, NativeSelectOption } from '$lib/components/ui/native-select/index.js';
	import { toast } from '$lib/toast';
	import GaugeIcon from '@lucide/svelte/icons/gauge';

	let countingTypes = $state<ScoreCountingType[]>([]);
	let name = $state('');
	let countingType = $state<number>(0);
	let winThreshold = $state<string>('1');
	let mustWinBy = $state<string>('1');
	let instantWinThreshold = $state<string>('0');
	let error = $state('');
	let submitting = $state(false);
	let attempted = $state(false);

	onMount(async () => {
		try {
			countingTypes = await getScoreCountingTypes();
			if (countingTypes.length > 0) countingType = countingTypes[0].type;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load score counting types';
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		attempted = true;
		error = '';
		submitting = true;
		try {
			const created = await createScoringStructure({
				name: name.trim(),
				win_condition_counting_type: countingType,
				win_condition: {
					win_threshold: parseInt(winThreshold, 10),
					must_win_by: parseInt(mustWinBy, 10),
					instant_win_threshold: parseInt(instantWinThreshold || '0', 10)
				}
			});
			toast.success('Scoring structure created');
			await goto(`/scoring-structures/${created.id}`);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create scoring structure';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | New scoring structure</title>
</svelte:head>

<PageHeader
	title="New scoring structure"
	icon={GaugeIcon}
	backHref="/scoring-structures"
	backLabel="Back to scoring structures"
/>

<Card class="mt-6 max-w-md">
	<CardHeader>
		<CardTitle>Scoring structure details</CardTitle>
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
				<Label for="countingType">Score counting type</Label>
				<NativeSelect id="countingType" bind:value={countingType} class="w-full">
					{#each countingTypes as ct}
						<NativeSelectOption value={ct.type}>{ct.name}</NativeSelectOption>
					{/each}
				</NativeSelect>
			</div>
			<div class="flex flex-col gap-2">
				<Label for="winThreshold">Win threshold</Label>
				<Input
					id="winThreshold"
					type="number"
					min="1"
					bind:value={winThreshold}
					required
					aria-invalid={attempted && !winThreshold}
				/>
			</div>
			<div class="flex flex-col gap-2">
				<Label for="mustWinBy">Must win by</Label>
				<Input
					id="mustWinBy"
					type="number"
					min="1"
					bind:value={mustWinBy}
					required
					aria-invalid={attempted && !mustWinBy}
				/>
			</div>
			<div class="flex flex-col gap-2">
				<Label for="instantWinThreshold">Instant win threshold</Label>
				<Input
					id="instantWinThreshold"
					type="number"
					min="0"
					bind:value={instantWinThreshold}
					placeholder="0 (disabled)"
				/>
			</div>
			<Button type="submit" disabled={submitting} class="w-fit">
				{submitting ? 'Creating...' : 'Create scoring structure'}
			</Button>
		</form>
	</CardContent>
</Card>
