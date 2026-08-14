<script lang="ts">
	import type { StandingsEntry } from '$lib/match';
	import {
		Card,
		CardContent,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card/index.js';

	let { standings, teamName }: {
		/** Weekly standings entries (fetched from GET /match/standings?season_id=). */
		standings: StandingsEntry[];
		/** Resolves a team id to its display name. */
		teamName: (teamId: string) => string;
	} = $props();
</script>

<Card class="mt-6">
	<CardHeader>
		<CardTitle class="text-base">Standings</CardTitle>
	</CardHeader>
	<CardContent>
		{#if standings.length === 0}
			<p class="text-sm text-muted-foreground">
				No completed matches yet. Standings update as team matches are completed.
			</p>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b text-left text-muted-foreground">
							<th class="py-1 pr-4 font-medium">Team</th>
							<th class="py-1 pr-4 font-medium">Wins</th>
							<th class="py-1 pr-4 font-medium">Losses</th>
							<th class="py-1 pr-4 font-medium">Ties</th>
						</tr>
					</thead>
					<tbody>
						{#each standings as entry}
							<tr class="border-b">
								<td class="py-1 pr-4 font-medium">{teamName(entry.team_id)}</td>
								<td class="py-1 pr-4">{entry.wins}</td>
								<td class="py-1 pr-4">{entry.losses}</td>
								<td class="py-1 pr-4">{entry.ties}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	</CardContent>
</Card>
