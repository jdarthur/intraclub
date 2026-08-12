<script lang="ts">
	import { onMount } from 'svelte';
	import { listScoringStructures, getScoreCountingTypes } from '$lib/scoringStructure';
	import type { ScoringStructure, ScoreCountingType } from '$lib/scoringStructure';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let structures = $state<ScoringStructure[]>([]);
	let countingTypes = $state<ScoreCountingType[]>([]);
	let loading = $state(true);
	let error = $state('');

	function countingTypeName(type: number): string {
		return countingTypes.find((t) => t.type === type)?.name ?? String(type);
	}

	onMount(async () => {
		try {
			[structures, countingTypes] = await Promise.all([
				listScoringStructures(),
				getScoreCountingTypes()
			]);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load scoring structures';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Intraclub | Scoring Structures</title>
</svelte:head>

<div class="flex items-center justify-between gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Scoring Structures</h1>
	<Button href="/scoring-structures/new">New scoring structure</Button>
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{:else if structures.length === 0}
	<p class="text-muted-foreground">No scoring structures yet.</p>
{:else}
	<div class="mt-4 overflow-hidden rounded-lg border">
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead>Name</TableHead>
					<TableHead>Counting type</TableHead>
					<TableHead>Win threshold</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each structures as structure}
					<TableRow>
						<TableCell>
							<a
								href={`/scoring-structures/${structure.id}`}
								class="font-medium text-primary underline-offset-4 hover:underline"
							>
								{structure.name}
							</a>
						</TableCell>
						<TableCell>{countingTypeName(structure.win_condition_counting_type)}</TableCell>
						<TableCell>{structure.win_condition.win_threshold}</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
