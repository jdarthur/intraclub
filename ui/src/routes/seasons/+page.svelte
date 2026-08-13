<script lang="ts">
	import { onMount } from 'svelte';
	import { listSeasons } from '$lib/season';
	import type { Season } from '$lib/season';
	import { listFacilities } from '$lib/facility';
	import { listDrafts } from '$lib/draft';
	import { listPlayoffStructures } from '$lib/playoffStructure';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let seasons = $state<Season[]>([]);
	let facilities = $state<Record<string, string>>({});
	let drafts = $state<Record<string, string>>({});
	let playoffs = $state<Record<string, string>>({});
	let loading = $state(true);
	let error = $state('');

	function playoffLabel(p: { byes: number; number_of_teams: number }): string {
		if (p.number_of_teams === 0 && p.byes === 0) return '';
		return `${p.number_of_teams} teams${p.byes > 0 ? ` / ${p.byes} bye${p.byes > 1 ? 's' : ''}` : ''}`;
	}

	onMount(async () => {
		try {
			const [seasonList, facilityList, draftList, playoffList] = await Promise.all([
				listSeasons(),
				listFacilities(),
				listDrafts(),
				listPlayoffStructures()
			]);
			seasons = seasonList;
			facilities = Object.fromEntries(facilityList.map((f) => [f.id, f.name]));
			drafts = Object.fromEntries(draftList.map((d) => [d.id, d.name]));
			playoffs = Object.fromEntries(
				playoffList.map((p) => [p.id, playoffLabel(p)])
			);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load seasons';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Intraclub | Seasons</title>
</svelte:head>

<div class="flex items-center justify-between gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Seasons</h1>
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{:else if seasons.length === 0}
	<p class="text-muted-foreground">No seasons yet. Complete a draft to create one.</p>
{:else}
	<div class="mt-4 overflow-hidden rounded-lg border">
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead>Name</TableHead>
					<TableHead>Facility</TableHead>
					<TableHead>Start time</TableHead>
					<TableHead>Draft</TableHead>
					<TableHead>Playoff structure</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each seasons as season}
					<TableRow>
						<TableCell>
							<a
								href={`/seasons/${season.id}`}
								class="font-medium text-primary underline-offset-4 hover:underline"
							>
								{season.name}
							</a>
						</TableCell>
						<TableCell>{facilities[season.facility] ?? season.facility}</TableCell>
						<TableCell>{season.start_time}</TableCell>
						<TableCell>{drafts[season.draft_id] ?? season.draft_id}</TableCell>
						<TableCell>
							{playoffs[season.playoff_structure] ?? season.playoff_structure}
						</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
