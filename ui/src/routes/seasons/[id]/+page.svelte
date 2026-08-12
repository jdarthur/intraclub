<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getSeason, listSeasonTeams, listTeamRatings } from '$lib/season';
	import type { Season, SeasonTeam, TeamRating } from '$lib/season';
	import { listDraftCaptains } from '$lib/draftCaptain';
	import type { DraftCaptain } from '$lib/draftCaptain';
	import { listUsers, fullName } from '$lib/user';
	import type { User } from '$lib/user';
	import { listRatings } from '$lib/rating';
	import type { Rating } from '$lib/rating';
	import { listFacilities } from '$lib/facility';
	import type { Facility } from '$lib/facility';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import {
		Card,
		CardContent,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card/index.js';

	const id = () => page.params.id as string;

	let season = $state<Season | null>(null);
	let seasonTeams = $state<SeasonTeam[]>([]);
	let teamRatings = $state<TeamRating[]>([]);
	let draftCaptains = $state<DraftCaptain[]>([]);
	let users = $state<User[]>([]);
	let ratingsById = $state<Record<string, Rating>>({});
	let facilities = $state<Facility[]>([]);

	let loading = $state(true);
	let loadError = $state('');

	async function load() {
		loading = true;
		loadError = '';
		try {
			const [seasonData, seasonTeamList, teamRatingList, draftCaptainList, userList, ratingList, facilityList] =
				await Promise.all([
					getSeason(id()),
					listSeasonTeams(),
					listTeamRatings(),
					listDraftCaptains(),
					listUsers(),
					listRatings(),
					listFacilities()
				]);
			season = seasonData;
			seasonTeams = seasonTeamList.filter((st) => st.season_id === seasonData.id);
			teamRatings = teamRatingList;
			draftCaptains = draftCaptainList;
			users = userList;
			ratingsById = Object.fromEntries(ratingList.map((r) => [r.id, r]));
			facilities = facilityList;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load season';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function userName(userId: string): string {
		const u = users.find((x) => x.id === userId);
		return u ? fullName(u) : userId;
	}

	function ratingName(ratingId: string): string {
		return ratingsById[ratingId]?.name ?? ratingId;
	}

	function facilityName(facilityId: string): string {
		return facilities.find((f) => f.id === facilityId)?.name ?? facilityId;
	}

	// The season's teams are reconstructed from the public season_team /
	// draft_captain join rows (Team/TeamAssignment records are restricted to
	// team members). Teams were named "Team N" in draft order. Captains are
	// pre-assigned at draft init and don't get a team_rating row, so they're
	// added from draft_captain.
	function teamInfo(seasonTeamId: string) {
		const captain = draftCaptains.find((c) => c.team_id === seasonTeamId);
		const order = captain ? captain.draft_order : 0;
		const members = teamRatings.filter((tr) => tr.team_id === seasonTeamId);
		if (captain && !members.some((m) => m.user_id === captain.captain_id)) {
			members.push({ id: '', team_id: seasonTeamId, user_id: captain.captain_id, rating_id: '' });
		}
		return { name: `Team ${order + 1}`, captainId: captain?.captain_id, members };
	}

	const teams = $derived(
		seasonTeams
			.map((st) => ({ teamId: st.team_id, ...teamInfo(st.team_id) }))
			.sort((a, b) => {
				const na = parseInt(a.name.replace(/[^\d]/g, ''), 10) || 0;
				const nb = parseInt(b.name.replace(/[^\d]/g, ''), 10) || 0;
				return na - nb;
			})
	);

	function sortedMembers(members: TeamRating[]) {
		return [...members].sort((a, b) => userName(a.user_id).localeCompare(userName(b.user_id)));
	}
</script>

<svelte:head>
	<title>{season ? season.name : 'Season'}</title>
</svelte:head>

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">{season?.name ?? 'Season'}</h1>
	<div class="ml-auto flex items-center gap-3">
		{#if season}
			<a href={`/drafts/${season.draft_id}`} class="text-sm text-muted-foreground hover:text-foreground">
				&larr; View draft
			</a>
		{/if}
	</div>
</div>

{#if loading}
	<p class="mt-4 text-muted-foreground">Loading...</p>
{:else if loadError}
	<p class="mt-4 text-sm font-medium text-destructive">{loadError}</p>
{:else if season}
	<Card class="mt-6">
		<CardHeader>
			<CardTitle class="text-base">Season details</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="flex flex-wrap items-center gap-4 text-sm">
				<Badge variant="secondary">Season</Badge>
				<span class="text-muted-foreground">Facility: {facilityName(season.facility)}</span>
				<span class="text-muted-foreground">Start time: {season.start_time}</span>
				<span class="text-muted-foreground">{teams.length} teams</span>
			</div>
		</CardContent>
	</Card>

	{#if teams.length === 0}
		<p class="mt-6 text-sm text-muted-foreground">No teams have been added to this season yet.</p>
	{:else}
		<div class="mt-6 grid gap-6 lg:grid-cols-2">
			{#each teams as team}
				<Card>
					<CardHeader>
						<CardTitle class="text-base">{team.name}</CardTitle>
					</CardHeader>
					<CardContent>
						{#if team.members.length === 0}
							<p class="text-sm text-muted-foreground">No players assigned.</p>
						{:else}
							<ul class="mt-2 space-y-2 text-sm">
								{#each sortedMembers(team.members) as member}
									<li class="flex items-center justify-between gap-3">
										<span class="font-medium">
											{userName(member.user_id)}
											{#if member.user_id === team.captainId}
												<span class="text-xs text-muted-foreground">(captain)</span>
											{/if}
										</span>
										{#if ratingName(member.rating_id)}
											<span class="text-xs text-muted-foreground">
												{ratingName(member.rating_id)}
											</span>
										{/if}
									</li>
								{/each}
							</ul>
						{/if}
					</CardContent>
				</Card>
			{/each}
		</div>
	{/if}
{/if}
