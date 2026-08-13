<script lang="ts">
	import { onMount } from 'svelte';
	import { listTeams } from '$lib/team';
	import type { TeamRoster } from '$lib/team';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let rosters = $state<TeamRoster[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			rosters = await listTeams();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load teams';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Intraclub | Teams</title>
</svelte:head>

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Teams</h1>
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{:else if rosters.length === 0}
	<p class="text-muted-foreground">
		You aren't on any teams. Teams are created when a draft is finalized and are only visible to their members.
	</p>
{:else}
	<div class="mt-4 overflow-hidden rounded-lg border">
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead>Team</TableHead>
					<TableHead>Members</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each rosters as roster}
					<TableRow>
						<TableCell>
							<a
								href={`/teams/${roster.team.id}`}
								class="font-medium text-primary underline-offset-4 hover:underline"
							>
								{roster.team.name}
							</a>
						</TableCell>
						<TableCell class="text-muted-foreground">
							{roster.assignments.length}
						</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
