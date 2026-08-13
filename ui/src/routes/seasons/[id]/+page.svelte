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
	import { getCurrentUserId } from '$lib/auth';
	import { listWeeksForSeason } from '$lib/week';
	import type { Week } from '$lib/week';
	import { listTeams } from '$lib/team';
	import type { TeamRoster } from '$lib/team';
	import {
		AVAILABILITY_OPTIONS,
		getAvailabilityForUser,
		getAvailabilityForTeam,
		setAvailability,
		availabilityLabel
	} from '$lib/availability';
	import type { AvailabilityOption } from '$lib/availability';
	import {
		createSchedule,
		getScheduleForSeason,
		assignWeeklyMatchup
	} from '$lib/schedule';
	import type { ScheduleDetail, WeeklyMatchup } from '$lib/schedule';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import {
		Card,
		CardContent,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		NativeSelect,
		NativeSelectOption
	} from '$lib/components/ui/native-select/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';

	const id = () => page.params.id as string;

	let season = $state<Season | null>(null);
	let seasonTeams = $state<SeasonTeam[]>([]);
	let teamRatings = $state<TeamRating[]>([]);
	let draftCaptains = $state<DraftCaptain[]>([]);
	let users = $state<User[]>([]);
	let ratingsById = $state<Record<string, Rating>>({});
	let facilities = $state<Facility[]>([]);
	let weeks = $state<Week[]>([]);
	let scheduleDetail = $state<ScheduleDetail | null>(null);

	// player availability state
	let rosters = $state<TeamRoster[]>([]);
	let myAvailability = $state<Record<string, AvailabilityOption>>({});
	let teamAvailability = $state<Record<string, Record<string, AvailabilityOption>>>({});
	let availabilityError = $state('');
	let savingWeek = $state<string | null>(null);

	let loading = $state(true);
	let loadError = $state('');

	// schedule builder state
	let creatingSchedule = $state(false);
	let scheduleError = $state('');
	let editingWeekId = $state<string | null>(null);
	let draftRows = $state<Record<string, EditRow[]>>({});

	interface EditRow {
		home: string;
		away: string;
		bye: boolean;
	}

	const isCommissioner = $derived(
		scheduleDetail !== null && scheduleDetail.commissioners.includes(getCurrentUserId() ?? '')
	);

	async function load() {
		loading = true;
		loadError = '';
		try {
			const [
				seasonData,
				seasonTeamList,
				teamRatingList,
				draftCaptainList,
				userList,
				ratingList,
				facilityList,
				weekList,
				scheduleData,
				rosterList
			] = await Promise.all([
				getSeason(id()),
				listSeasonTeams(),
				listTeamRatings(),
				listDraftCaptains(),
				listUsers(),
				listRatings(),
				listFacilities(),
				listWeeksForSeason(id()),
				getScheduleForSeason(id()),
				listTeams()
			]);
			season = seasonData;
			seasonTeams = seasonTeamList.filter((st) => st.season_id === seasonData.id);
			teamRatings = teamRatingList;
			draftCaptains = draftCaptainList;
			users = userList;
			ratingsById = Object.fromEntries(ratingList.map((r) => [r.id, r]));
			facilities = facilityList;
			weeks = weekList;
			scheduleDetail = scheduleData;
			rosters = rosterList;
			await loadAvailability();
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

	// --- player availability derived state --------------------------------

	const currentUserId = $derived(getCurrentUserId() ?? '');

	// The teams in this season that the current user is on (from the
	// reconstructed roster), used to gate the "set my availability" inputs.
	const myTeams = $derived(
		teams.filter((t) => t.members.some((m) => m.user_id === currentUserId))
	);

	// The teams in this season where the current user is a captain or
	// co-captain (from the role-bearing TeamRoster list), which gate the team
	// availability view.
	const myCaptainTeamIds = $derived(
		rosters
			.filter((r) =>
				r.assignments.some(
					(a) =>
						a.user_id === currentUserId &&
						(a.role === 'captain' || a.role === 'co_captain')
				)
			)
			.map((r) => r.team.id)
			.filter((tid) => teams.some((t) => t.teamId === tid))
	);

	const captainTeam = $derived(teams.find((t) => myCaptainTeamIds.includes(t.teamId)));

	function sortedMembers(members: TeamRating[]) {
		return [...members].sort((a, b) => userName(a.user_id).localeCompare(userName(b.user_id)));
	}

	function teamName(teamId: string): string {
		return teams.find((t) => t.teamId === teamId)?.name ?? teamId;
	}

	function weeklyMatchupFor(weekId: string): WeeklyMatchup | undefined {
		return scheduleDetail?.weekly_matchups.find((wm) => wm.week_id === weekId);
	}

	function weekLabel(week: Week): string {
		const d = new Date(week.date);
		const date = Number.isNaN(d.getTime()) ? week.date : d.toLocaleDateString();
		return week.note ? `${date} — ${week.note}` : date;
	}

	// --- player availability actions --------------------------------------

	// loadAvailability fetches the current user's availability for the season's
	// weeks, and (when they are a captain/co-captain) their team's availability
	// for the team view.
	async function loadAvailability() {
		if (!season || !currentUserId) return;
		const draftId = season.draft_id;
		try {
			const mine = await getAvailabilityForUser(draftId);
			myAvailability = Object.fromEntries(mine.map((a) => [a.week_id, a.available]));
			availabilityError = '';
		} catch (e) {
			availabilityError = e instanceof Error ? e.message : 'Failed to load availability';
			return;
		}

		if (myCaptainTeamIds.length === 0) return;
		try {
			const combined: Record<string, Record<string, AvailabilityOption>> = {};
			for (const teamId of myCaptainTeamIds) {
				const entries = await getAvailabilityForTeam(teamId, draftId);
				for (const entry of entries) {
					combined[entry.user_id] = Object.fromEntries(
						entry.availabilities.map((a) => [a.week_id, a.available])
					);
				}
			}
			teamAvailability = combined;
		} catch (e) {
			availabilityError = e instanceof Error ? e.message : 'Failed to load team availability';
		}
	}

	// setWeekAvailability saves the current user's availability for a week and
	// refreshes the availability data (including the team view, if shown) from
	// the API so the UI reflects the saved value.
	async function setWeekAvailability(weekId: string, value: string) {
		const option = Number(value) as AvailabilityOption;
		savingWeek = weekId;
		availabilityError = '';
		try {
			await setAvailability(weekId, option);
			await loadAvailability();
		} catch (e) {
			availabilityError = e instanceof Error ? e.message : 'Failed to save availability';
		} finally {
			savingWeek = null;
		}
	}

	// --- schedule create / assign actions -----------------------------------

	async function loadSchedule() {
		scheduleDetail = await getScheduleForSeason(id());
	}

	async function handleCreateSchedule() {
		creatingSchedule = true;
		scheduleError = '';
		try {
			await createSchedule(id());
			await loadSchedule();
		} catch (e) {
			scheduleError = e instanceof Error ? e.message : 'Failed to create schedule';
		} finally {
			creatingSchedule = false;
		}
	}

	function rowsFromMatchup(wm: WeeklyMatchup | undefined): EditRow[] {
		if (wm && wm.matchups.length > 0) {
			return wm.matchups.map((m) => ({
				home: m.home_team_id,
				away: m.bye ? '' : m.away_team_id,
				bye: m.bye
			}));
		}
		return [{ home: '', away: '', bye: false }];
	}

	function startEdit(weekId: string) {
		draftRows[weekId] = rowsFromMatchup(weeklyMatchupFor(weekId));
		editingWeekId = weekId;
		scheduleError = '';
	}

	function addRow(weekId: string) {
		draftRows[weekId].push({ home: '', away: '', bye: false });
	}

	function removeRow(weekId: string, index: number) {
		draftRows[weekId].splice(index, 1);
	}

	function setHome(weekId: string, index: number, value: string) {
		draftRows[weekId][index].home = value;
	}

	function setAway(weekId: string, index: number, value: string) {
		draftRows[weekId][index].away = value;
	}

	function toggleBye(weekId: string, index: number, checked: boolean) {
		const row = draftRows[weekId][index];
		row.bye = checked;
		if (checked) row.away = '';
	}

	function cancelEdit() {
		editingWeekId = null;
		scheduleError = '';
	}

	async function saveWeek(weekId: string) {
		if (!scheduleDetail?.schedule) return;
		scheduleError = '';
		const rows = draftRows[weekId] ?? [];
		const matchups = rows
			.filter((r) => r.home)
			.map((r) => ({
				home_team_id: r.home,
				away_team_id: r.bye ? '' : r.away,
				bye: r.bye
			}));
		try {
			await assignWeeklyMatchup(scheduleDetail.schedule.id, weekId, matchups);
			editingWeekId = null;
			await loadSchedule();
		} catch (e) {
			scheduleError = e instanceof Error ? e.message : 'Failed to save weekly matchup';
		}
	}
</script>

<svelte:head>
	<title>Intraclub | {season ? season.name : 'Season'}</title>
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

	<!-- Schedule -->
	<Card class="mt-6">
		<CardHeader>
			<CardTitle class="text-base">Season schedule</CardTitle>
		</CardHeader>
		<CardContent>
			{#if scheduleError}
				<p class="mb-4 text-sm font-medium text-destructive">{scheduleError}</p>
			{/if}

			{#if scheduleDetail && !scheduleDetail.schedule}
				{#if isCommissioner}
					<p class="text-sm text-muted-foreground">
						No schedule yet. Create one to assign weekly matchups to each week.
					</p>
					<Button class="mt-4" onclick={handleCreateSchedule} disabled={creatingSchedule}>
						{creatingSchedule ? 'Creating…' : 'Create schedule'}
					</Button>
				{:else}
					<p class="text-sm text-muted-foreground">
						No schedule has been created for this season yet.
					</p>
				{/if}
			{:else if scheduleDetail}
				{#if weeks.length === 0}
					<p class="text-sm text-muted-foreground">
						No weeks have been added to this season yet.
					</p>
				{:else}
					<ul class="flex flex-col gap-4">
						{#each weeks as week (week.id)}
							{#if editingWeekId === week.id}
								<li class="rounded-lg border p-4">
									<div class="mb-3 flex items-center justify-between">
										<span class="text-sm font-medium">{weekLabel(week)}</span>
										<Button
											type="button"
											variant="ghost"
											size="sm"
											onclick={cancelEdit}
										>
											Cancel
										</Button>
									</div>
									{#if teams.length === 0}
										<p class="text-sm text-muted-foreground">No teams to schedule.</p>
									{:else}
										{#each draftRows[week.id] ?? [] as row, i (i)}
											<div class="mb-3 flex flex-wrap items-end gap-3">
												<div class="flex flex-col gap-1">
													<Label for={`${week.id}-home-${i}`} class="text-xs">Home</Label>
													<NativeSelect
														id={`${week.id}-home-${i}`}
														value={row.home}
														oninput={(e) =>
															setHome(week.id, i, (e.currentTarget as HTMLSelectElement).value)
														}
													>
														<NativeSelectOption value="" disabled>Select team…</NativeSelectOption>
														{#each teams as t}
															<NativeSelectOption value={t.teamId}>{t.name}</NativeSelectOption>
														{/each}
													</NativeSelect>
												</div>
												<div class="flex flex-col gap-1">
													<Label for={`${week.id}-away-${i}`} class="text-xs">Away</Label>
													<NativeSelect
														id={`${week.id}-away-${i}`}
														value={row.away}
														disabled={row.bye}
														oninput={(e) =>
															setAway(week.id, i, (e.currentTarget as HTMLSelectElement).value)
														}
													>
														<NativeSelectOption value="" disabled>Select team…</NativeSelectOption>
														{#each teams as t}
															<NativeSelectOption value={t.teamId}>{t.name}</NativeSelectOption>
														{/each}
													</NativeSelect>
												</div>
												<div class="flex items-center gap-2 pb-1">
													<Checkbox
														checked={row.bye}
														onCheckedChange={(checked) => toggleBye(week.id, i, checked)}
													/>
													<Label class="text-sm">Bye</Label>
												</div>
												<Button
													type="button"
													variant="ghost"
													size="sm"
													onclick={() => removeRow(week.id, i)}
												>
													Remove
												</Button>
											</div>
										{/each}
										<div class="mt-2 flex items-center gap-3">
											<Button
												type="button"
												variant="outline"
												size="sm"
												onclick={() => addRow(week.id)}
											>
												Add matchup
											</Button>
											<Button type="button" size="sm" onclick={() => saveWeek(week.id)}>
												Save
											</Button>
										</div>
									{/if}
								</li>
							{:else}
								<li class="rounded-lg border p-4">
									<div class="mb-2 flex items-center justify-between">
										<span class="text-sm font-medium">{weekLabel(week)}</span>
										{#if isCommissioner}
											<Button
												type="button"
												variant="outline"
												size="sm"
												onclick={() => startEdit(week.id)}
											>
												Edit
											</Button>
										{/if}
									</div>
									{#if (weeklyMatchupFor(week.id)?.matchups.length ?? 0) > 0}
										<ul class="flex flex-col gap-1 text-sm">
											{#each weeklyMatchupFor(week.id)!.matchups as m}
												<li>
													{#if m.bye}
														{teamName(m.home_team_id)} — bye
													{:else}
														{teamName(m.home_team_id)} vs {teamName(m.away_team_id)}
													{/if}
												</li>
											{/each}
										</ul>
									{:else}
										<p class="text-sm text-muted-foreground">No matchups assigned.</p>
									{/if}
								</li>
							{/if}
						{/each}
					</ul>
				{/if}
			{/if}
		</CardContent>
	</Card>

	<!-- Player availability -->
	<Card class="mt-6">
		<CardHeader>
			<CardTitle class="text-base">Player availability</CardTitle>
		</CardHeader>
		<CardContent>
			{#if availabilityError}
				<p class="mb-4 text-sm font-medium text-destructive">{availabilityError}</p>
			{/if}

			{#if weeks.length === 0}
				<p class="text-sm text-muted-foreground">
					No weeks have been added to this season yet.
				</p>
			{:else if myTeams.length === 0}
				<p class="text-sm text-muted-foreground">
					You are not on a team in this season, so there is nothing to set.
				</p>
			{:else}
				<p class="mb-4 text-sm text-muted-foreground">
					Set your availability for each week. Captains and co-captains can also view their
					team's availability below.
				</p>
				<ul class="flex flex-col gap-3">
					{#each weeks as week (week.id)}
						<li class="flex items-center justify-between gap-4">
							<Label for={`availability-${week.id}`} class="text-sm font-medium">
								{weekLabel(week)}
							</Label>
							<div class="flex items-center gap-3">
								{#if savingWeek === week.id}
									<span class="text-xs text-muted-foreground">Saving…</span>
								{/if}
								<NativeSelect
									id={`availability-${week.id}`}
									value={String(myAvailability[week.id] ?? 0)}
									oninput={(e) =>
										setWeekAvailability(
											week.id,
											(e.currentTarget as HTMLSelectElement).value
										)
									}
								>
									{#each AVAILABILITY_OPTIONS as opt}
										<NativeSelectOption value={String(opt.value)}>
											{opt.label}
										</NativeSelectOption>
									{/each}
								</NativeSelect>
							</div>
						</li>
					{/each}
				</ul>
			{/if}

			{#if captainTeam}
				<div class="mt-6">
					<h3 class="mb-2 text-sm font-semibold">
						{captainTeam.name} — team availability
					</h3>
					<div class="overflow-x-auto">
						<table class="w-full text-sm">
							<thead>
								<tr class="border-b text-left text-muted-foreground">
									<th class="py-1 pr-4 font-medium">Player</th>
									{#each weeks as week}
										<th class="py-1 pr-4 font-medium">{weekLabel(week)}</th>
									{/each}
								</tr>
							</thead>
							<tbody>
								{#each sortedMembers(captainTeam.members) as member}
									<tr class="border-b">
										<td class="py-1 pr-4">
											{userName(member.user_id)}
											{#if member.user_id === captainTeam.captainId}
												<span class="text-xs text-muted-foreground">(captain)</span>
											{/if}
										</td>
										{#each weeks as week}
											<td class="py-1 pr-4">
												{availabilityLabel(teamAvailability[member.user_id]?.[week.id] ?? 0)}
											</td>
										{/each}
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}
		</CardContent>
	</Card>

	{#if teams.length === 0}
		<p class="mt-6 text-sm text-muted-foreground">No teams have been added to this season yet.</p>
	{:else}
		<div class="mt-6 grid gap-6 lg:grid-cols-2">
			{#each teams as team}
				<Card>
					<CardHeader>
						<CardTitle class="text-base">
							<a
								href={`/teams/${team.teamId}`}
								class="text-primary underline-offset-4 hover:underline"
							>
								{team.name}
							</a>
						</CardTitle>
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
