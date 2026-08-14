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
	import { getCurrentUserId, getRoles } from '$lib/auth';
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
	import {
		getLineupDetail,
		setLineup,
		confirmLineup,
		markOfficial
	} from '$lib/lineup';
	import type { LineupDetail, PairingInput } from '$lib/lineup';
	import {
		generateMatches,
		getWeekMatches,
		recordScore,
		completeMatch,
		getStandings
	} from '$lib/match';
	import type { WeekMatchDetail, StandingsEntry, IndividualMatch } from '$lib/match';
	import { listScoringStructures } from '$lib/scoringStructure';
	import type { ScoringStructure } from '$lib/scoringStructure';
	import {
		listSeasonLateAdditions,
		addSeasonLateAddition,
		removeSeasonLateAddition
	} from '$lib/seasonLateAddition';
	import type { SeasonLateAddition } from '$lib/seasonLateAddition';
	import {
		listSeasonCommissioners,
		addSeasonCommissioner,
		removeSeasonCommissioner
	} from '$lib/seasonCommissioner';
	import type { SeasonCommissioner } from '$lib/seasonCommissioner';
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
	import { Input } from '$lib/components/ui/input/index.js';
	import Standings from '$lib/components/Standings.svelte';

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

	// weekly lineup state
	let lineupDetails = $state<Record<string, Record<string, LineupDetail>>>({});
	let lineupDrafts = $state<Record<string, PairingInput[]>>({});
	let editingLineupWeek = $state<string | null>(null);
	let lineupError = $state('');
	let savingLineup = $state(false);

	// weekly match scoring state
	let matchDetails = $state<Record<string, WeekMatchDetail>>({});
	let standings = $state<StandingsEntry[]>([]);
	let matchError = $state('');
	let scoringStructures = $state<ScoringStructure[]>([]);
	let selectedScoringStructure = $state('');
	let generatingWeek = $state<string | null>(null);
	let scoreDrafts = $state<Record<string, { main: string; secondary: string }>>({});
	let savingScore = $state<string | null>(null);
	let completingMatch = $state<string | null>(null);

	// late additions state (sysadmin-only writes)
	let lateAdditions = $state<SeasonLateAddition[]>([]);
	let lateAdditionUserId = $state('');
	let lateAdditionError = $state('');
	let savingLateAddition = $state(false);
	let isSysAdmin = $state(false);

	// co-commissioners state (sysadmin-only writes)
	let seasonCommissioners = $state<SeasonCommissioner[]>([]);
	let commissionerUserId = $state('');
	let commissionerError = $state('');
	let savingCommissioner = $state(false);

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
				rosterList,
				scoringStructureList,
				lateAdditionList,
				commissionerList
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
				listTeams(),
				listScoringStructures(),
				listSeasonLateAdditions(),
				listSeasonCommissioners()
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
			scoringStructures = scoringStructureList;
			lateAdditions = lateAdditionList.filter((l) => l.season_id === seasonData.id);
			seasonCommissioners = commissionerList.filter((c) => c.season_id === seasonData.id);
			isSysAdmin = (await getRoles()).includes('System Administrator');
			selectedScoringStructure =
				scoringStructureList.find((s) => s.name === 'Tennis standard set')?.id ??
				scoringStructureList[0]?.id ??
				'';
			await loadAvailability();
			await loadLineups();
			await loadMatches();
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

	// --- weekly lineup actions -----------------------------------------------

	// loadLineups fetches the LineupDetail for every team x week combination the
	// current user can act on: their own team (if captain/co-captain) and all
	// teams (if commissioner, so they can mark lineups official).
	async function loadLineups() {
		const teamsToLoad = isCommissioner ? teams : captainTeam ? [captainTeam] : [];
		const details: Record<string, Record<string, LineupDetail>> = {};
		for (const team of teamsToLoad) {
			details[team.teamId] = {};
			for (const week of weeks) {
				try {
					details[team.teamId][week.id] = await getLineupDetail(team.teamId, week.id);
				} catch {
					// ignore individual failures; the row just stays empty
				}
			}
		}
		lineupDetails = details;
	}

	// members eligible to play a slot in a line, filtered by the required rating.
	function lineupPlayerOptions(ratingId: string) {
		return sortedMembers(captainTeam!.members.filter((m) => m.rating_id === ratingId));
	}

	function setLineupPlayer(weekId: string, idx: number, slot: 'player1' | 'player2', value: string) {
		lineupDrafts[weekId][idx][slot] = value;
	}

	// startEditLineup seeds the draft rows from the existing pairings (if any),
	// one row per format line.
	function startEditLineup(weekId: string) {
		const detail = lineupDetails[captainTeam!.teamId]?.[weekId];
		const pairings = detail?.pairings ?? [];
		lineupDrafts[weekId] =
			detail?.lines.map((line, idx) => {
				const p = pairings.find((x) => x.format_line_index === idx);
				return {
					player1: p?.player1 ?? '',
					player2: p?.player2 ?? '',
					format_line_index: idx
				};
			}) ?? [];
		editingLineupWeek = weekId;
		lineupError = '';
	}

	async function saveLineup(weekId: string) {
		if (!captainTeam) return;
		const pairings = (lineupDrafts[weekId] ?? []).filter((p) => p.player1 && p.player2);
		savingLineup = true;
		lineupError = '';
		try {
			await setLineup(captainTeam.teamId, weekId, pairings);
			editingLineupWeek = null;
			await loadLineups();
		} catch (e) {
			lineupError = e instanceof Error ? e.message : 'Failed to save lineup';
		} finally {
			savingLineup = false;
		}
	}

	async function confirmWeekLineup(weekId: string) {
		if (!captainTeam) return;
		const detail = lineupDetails[captainTeam.teamId]?.[weekId];
		if (!detail?.lineup) return;
		lineupError = '';
		try {
			await confirmLineup(detail.lineup.id);
			await loadLineups();
		} catch (e) {
			lineupError = e instanceof Error ? e.message : 'Failed to confirm lineup';
		}
	}

	async function markWeekOfficial(teamId: string, weekId: string) {
		const detail = lineupDetails[teamId]?.[weekId];
		if (!detail?.lineup) return;
		lineupError = '';
		try {
			await markOfficial(detail.lineup.id);
			await loadLineups();
		} catch (e) {
			lineupError = e instanceof Error ? e.message : 'Failed to mark lineup official';
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

	// --- weekly match scoring actions ----------------------------------------

	// loadMatches fetches the week score sheet for every week plus the season
	// standings. Individual week failures are tolerated so the rest of the page
	// still renders.
	async function loadMatches() {
		const details: Record<string, WeekMatchDetail> = {};
		for (const week of weeks) {
			try {
				details[week.id] = await getWeekMatches(week.id);
			} catch {
				// ignore individual failures; the week just stays empty
			}
		}
		matchDetails = details;
		try {
			standings = await getStandings(id());
		} catch {
			// standings are a nice-to-have; ignore load failures
		}
	}

	function setScoreDraft(matchId: string, field: 'main' | 'secondary', value: string) {
		const current = scoreDrafts[matchId] ?? { main: '', secondary: '' };
		scoreDrafts[matchId] = { ...current, [field]: value };
	}

	async function handleGenerateMatches(weekId: string) {
		if (!selectedScoringStructure) {
			matchError = 'Choose a scoring structure first.';
			return;
		}
		generatingWeek = weekId;
		matchError = '';
		try {
			matchDetails[weekId] = await generateMatches(weekId, selectedScoringStructure);
		} catch (e) {
			matchError = e instanceof Error ? e.message : 'Failed to generate matches';
		} finally {
			generatingWeek = null;
		}
	}

	async function saveScore(weekId: string, matchId: string) {
		const draft = scoreDrafts[matchId];
		const main = parseInt(draft?.main ?? '', 10) || 0;
		const secondary = parseInt(draft?.secondary ?? '', 10) || 0;
		savingScore = matchId;
		matchError = '';
		try {
			await recordScore(matchId, main, secondary);
			delete scoreDrafts[matchId];
			await loadMatches();
		} catch (e) {
			matchError = e instanceof Error ? e.message : 'Failed to save score';
		} finally {
			savingScore = null;
		}
	}

	async function handleCompleteMatch(weekId: string, matchId: string) {
		completingMatch = matchId;
		matchError = '';
		try {
			await completeMatch(matchId);
			delete scoreDrafts[matchId];
			await loadMatches();
		} catch (e) {
			matchError = e instanceof Error ? e.message : 'Failed to complete match';
		} finally {
			completingMatch = null;
		}
	}

	function matchStatusLabel(status: number): string {
		switch (status) {
			case 1:
				return 'In progress';
			case 2:
				return 'Won';
			case 3:
				return 'Lost';
			default:
				return 'Unstarted';
		}
	}

	function matchPlayerLabel(m: IndividualMatch): string {
		return `${userName(m.player1)} & ${userName(m.player2)}`;
	}

	// --- late additions (sysadmin only) ------------------------------------

	// Whether a user is already part of the season (on a drafted team or an
	// existing late addition), used to filter the add dropdown.
	function userInSeason(userId: string): boolean {
		if (lateAdditions.some((l) => l.user_id === userId)) return true;
		return teams.some((t) => t.members.some((m) => m.user_id === userId));
	}

	const lateAdditionOptions = $derived(
		[...users].filter((u) => !userInSeason(u.id)).sort((a, b) => fullName(a).localeCompare(fullName(b)))
	);

	const lateAdditionNames = $derived(
		[...lateAdditions].sort((a, b) => userName(a.user_id).localeCompare(userName(b.user_id)))
	);

	// addLateAddition creates a SeasonLateAddition for the selected user and
	// refreshes the list from the API.
	async function handleAddLateAddition() {
		const seasonId = season?.id;
		if (!seasonId || !lateAdditionUserId) return;
		savingLateAddition = true;
		lateAdditionError = '';
		try {
			await addSeasonLateAddition(seasonId, lateAdditionUserId);
			lateAdditionUserId = '';
			lateAdditions = (await listSeasonLateAdditions()).filter((l) => l.season_id === seasonId);
		} catch (e) {
			lateAdditionError = e instanceof Error ? e.message : 'Failed to add late addition';
		} finally {
			savingLateAddition = false;
		}
	}

	// removeLateAddition deletes the SeasonLateAddition record and refreshes.
	async function handleRemoveLateAddition(lateId: string) {
		const seasonId = season?.id;
		savingLateAddition = true;
		lateAdditionError = '';
		try {
			await removeSeasonLateAddition(lateId);
			if (seasonId) {
				lateAdditions = (await listSeasonLateAdditions()).filter((l) => l.season_id === seasonId);
			}
		} catch (e) {
			lateAdditionError = e instanceof Error ? e.message : 'Failed to remove late addition';
		} finally {
			savingLateAddition = false;
		}
	}

	// --- co-commissioners (sysadmin only) --------------------------------

	// Users already serving as commissioners of this season (from the schedule
	// detail), used to filter the add dropdown and to render the current list.
	const currentCommissionerIds = $derived(scheduleDetail?.commissioners ?? []);

	// Co-commissioners are typically users who help administer the season
	// rather than players in it, so the add dropdown excludes anyone already in
	// the season (drafted roster / late addition) as well as current
	// commissioners — mirroring the late-addition dropdown.
	const commissionerOptions = $derived(
		[...users]
			.filter((u) => !userInSeason(u.id) && !currentCommissionerIds.includes(u.id))
			.sort((a, b) => fullName(a).localeCompare(fullName(b)))
	);

	const commissionerNames = $derived(
		[...seasonCommissioners].sort((a, b) => userName(a.user_id).localeCompare(userName(b.user_id)))
	);

	// addSeasonCommissioner creates a SeasonCommissioner for the selected user
	// and refreshes both the join rows and the schedule detail (so the new
	// co-commissioner immediately counts as a commissioner of the season).
	async function handleAddSeasonCommissioner() {
		const seasonId = season?.id;
		if (!seasonId || !commissionerUserId) return;
		savingCommissioner = true;
		commissionerError = '';
		try {
			await addSeasonCommissioner(seasonId, commissionerUserId);
			commissionerUserId = '';
			seasonCommissioners = (await listSeasonCommissioners()).filter((c) => c.season_id === seasonId);
			scheduleDetail = await getScheduleForSeason(seasonId);
		} catch (e) {
			commissionerError = e instanceof Error ? e.message : 'Failed to add co-commissioner';
		} finally {
			savingCommissioner = false;
		}
	}

	// removeSeasonCommissioner deletes the SeasonCommissioner record and
	// refreshes both the join rows and the schedule detail.
	async function handleRemoveSeasonCommissioner(scId: string) {
		const seasonId = season?.id;
		savingCommissioner = true;
		commissionerError = '';
		try {
			await removeSeasonCommissioner(scId);
			if (seasonId) {
				seasonCommissioners = (await listSeasonCommissioners()).filter((c) => c.season_id === seasonId);
				scheduleDetail = await getScheduleForSeason(seasonId);
			}
		} catch (e) {
			commissionerError = e instanceof Error ? e.message : 'Failed to remove co-commissioner';
		} finally {
			savingCommissioner = false;
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
			<a
				href={`/seasons/${season.id}/proposals`}
				class="text-sm text-muted-foreground hover:text-foreground"
			>
				Commissioner proposals
			</a>
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

	<!-- Weekly lineups -->
	<Card class="mt-6">
		<CardHeader>
			<CardTitle class="text-base">Weekly lineups</CardTitle>
		</CardHeader>
		<CardContent>
			{#if lineupError}
				<p class="mb-4 text-sm font-medium text-destructive">{lineupError}</p>
			{/if}

			{#if weeks.length === 0}
				<p class="text-sm text-muted-foreground">
					No weeks have been added to this season yet.
				</p>
			{:else if isCommissioner}
				<p class="mb-4 text-sm text-muted-foreground">
					Mark confirmed lineups official. A lineup must be confirmed by the team captain
					before it can be marked official.
				</p>
				<ul class="flex flex-col gap-4">
					{#each weeks as week (week.id)}
						<li class="rounded-lg border p-4">
							<span class="text-sm font-medium">{weekLabel(week)}</span>
							<ul class="mt-2 flex flex-col gap-2 text-sm">
								{#each teams as team}
									{@const d = lineupDetails[team.teamId]?.[week.id]}
									<li class="flex items-center justify-between gap-3">
										<span>
											{team.name}
											{#if d?.lineup?.confirmed}
												<span class="text-xs text-muted-foreground"> — confirmed</span>
											{/if}
											{#if d?.lineup?.official}
												<span class="text-xs text-muted-foreground"> — official</span>
											{/if}
										</span>
										{#if d?.lineup?.confirmed && !d.lineup.official}
											<Button
												type="button"
												size="sm"
												onclick={() => markWeekOfficial(team.teamId, week.id)}
											>
												Mark official
											</Button>
										{/if}
									</li>
								{/each}
							</ul>
						</li>
					{/each}
				</ul>
			{:else if captainTeam}
				<p class="mb-4 text-sm text-muted-foreground">
					Build and confirm {captainTeam.name}'s weekly lineup from the team's rated players.
				</p>
				<ul class="flex flex-col gap-4">
					{#each weeks as week (week.id)}
						{@const detail = lineupDetails[captainTeam.teamId]?.[week.id]}
						{@const confirmed = detail?.lineup?.confirmed}
						{@const official = detail?.lineup?.official}
						<li class="rounded-lg border p-4">
							<div class="mb-2 flex items-center justify-between">
								<span class="text-sm font-medium">{weekLabel(week)}</span>
								{#if official}
									<Badge variant="secondary">Official</Badge>
								{:else if confirmed}
									<Badge variant="secondary">Confirmed</Badge>
								{/if}
							</div>

							{#if editingLineupWeek === week.id}
								{#each lineupDrafts[week.id] ?? [] as row, idx (idx)}
									{@const line = detail?.lines[idx]}
									<div class="mb-3 flex flex-wrap items-end gap-3">
										<div class="flex flex-col gap-1">
											<Label for={`${week.id}-lineup-p1-${idx}`} class="text-xs">Player 1</Label>
											<NativeSelect
												id={`${week.id}-lineup-p1-${idx}`}
												value={row.player1}
												oninput={(e) =>
													setLineupPlayer(
														week.id,
														idx,
														'player1',
														(e.currentTarget as HTMLSelectElement).value
													)
												}
											>
												<NativeSelectOption value="" disabled>Select…</NativeSelectOption>
												{#each lineupPlayerOptions(line?.player_1_rating ?? '') as m}
													<NativeSelectOption value={m.user_id}>
														{userName(m.user_id)}
													</NativeSelectOption>
												{/each}
											</NativeSelect>
										</div>
										<div class="flex flex-col gap-1">
											<Label for={`${week.id}-lineup-p2-${idx}`} class="text-xs">Player 2</Label>
											<NativeSelect
												id={`${week.id}-lineup-p2-${idx}`}
												value={row.player2}
												oninput={(e) =>
													setLineupPlayer(
														week.id,
														idx,
														'player2',
														(e.currentTarget as HTMLSelectElement).value
													)
												}
											>
												<NativeSelectOption value="" disabled>Select…</NativeSelectOption>
												{#each lineupPlayerOptions(line?.player_2_rating ?? '') as m}
													<NativeSelectOption value={m.user_id}>
														{userName(m.user_id)}
													</NativeSelectOption>
												{/each}
											</NativeSelect>
										</div>
										{#if line}
											<span class="pb-1 text-xs text-muted-foreground">
												{ratingName(line.player_1_rating)} / {ratingName(line.player_2_rating)}
											</span>
										{/if}
									</div>
								{/each}
								<div class="mt-2 flex items-center gap-3">
									<Button type="button" size="sm" onclick={() => saveLineup(week.id)} disabled={savingLineup}>
										{savingLineup ? 'Saving…' : 'Save'}
									</Button>
									<Button
										type="button"
										variant="ghost"
										size="sm"
										onclick={() => (editingLineupWeek = null)}
									>
										Cancel
									</Button>
								</div>
							{:else}
								<div class="flex items-center justify-between gap-3">
									<span class="text-sm text-muted-foreground">
										{#if (detail?.pairings.length ?? 0) > 0}
											{detail!.pairings.length} line(s) assigned
										{:else}
											No lineup set.
										{/if}
									</span>
									<div class="flex items-center gap-2">
										{#if !confirmed && !official}
											<Button
												type="button"
												variant="outline"
												size="sm"
												onclick={() => startEditLineup(week.id)}
											>
												Build
											</Button>
											{#if detail?.lineup}
												<Button type="button" size="sm" onclick={() => confirmWeekLineup(week.id)}>
													Confirm
												</Button>
											{/if}
										{/if}
									</div>
								</div>
							{/if}
						</li>
					{/each}
				</ul>
			{:else}
				<p class="text-sm text-muted-foreground">
					You are not a captain or co-captain on a team in this season, and you are not a
					commissioner.
				</p>
			{/if}
		</CardContent>
	</Card>

	<!-- Match scoring -->
	<Card class="mt-6">
		<CardHeader>
			<CardTitle class="text-base">Match scoring</CardTitle>
		</CardHeader>
		<CardContent>
			{#if matchError}
				<p class="mb-4 text-sm font-medium text-destructive">{matchError}</p>
			{/if}

			{#if weeks.length === 0}
				<p class="text-sm text-muted-foreground">
					No weeks have been added to this season yet.
				</p>
			{:else}
				<p class="mb-4 text-sm text-muted-foreground">
					The commissioner generates a week's matches from the scheduled matchup and both
					teams' official lineups, then records scores and completes each individual match.
				</p>
				<ul class="flex flex-col gap-4">
					{#each weeks as week (week.id)}
						{@const detail = matchDetails[week.id]}
						<li class="rounded-lg border p-4">
							<div class="mb-2 flex items-center justify-between">
								<span class="text-sm font-medium">{weekLabel(week)}</span>
								{#if detail?.closed}
									<Badge variant="secondary">Closed</Badge>
								{/if}
							</div>

							{#if !detail || detail.team_matches.length === 0}
								{#if isCommissioner}
									<div class="flex flex-wrap items-end gap-3">
										<div class="flex flex-col gap-1">
											<Label for={`${week.id}-scoring-structure`} class="text-xs">
												Scoring structure
											</Label>
											<NativeSelect
												id={`${week.id}-scoring-structure`}
												value={selectedScoringStructure}
												oninput={(e) =>
													(selectedScoringStructure = (
														e.currentTarget as HTMLSelectElement
													).value)
												}
											>
												{#each scoringStructures as s}
													<NativeSelectOption value={s.id}>{s.name}</NativeSelectOption>
												{/each}
											</NativeSelect>
										</div>
										<Button
											type="button"
											size="sm"
											disabled={generatingWeek === week.id || scoringStructures.length === 0}
											onclick={() => handleGenerateMatches(week.id)}
										>
											{generatingWeek === week.id ? 'Generating…' : 'Generate matches'}
										</Button>
									</div>
								{:else}
									<p class="text-sm text-muted-foreground">
										No matches generated for this week yet.
									</p>
								{/if}
							{:else}
								<ul class="flex flex-col gap-3">
									{#each detail.team_matches as tm (tm.id)}
										<li class="rounded-lg border p-3">
											<div class="mb-2 flex flex-wrap items-center justify-between gap-2">
												<span class="text-sm font-semibold">
													{teamName(tm.home_team_id)} vs {teamName(tm.away_team_id)}
												</span>
												{#if tm.complete}
													<Badge variant="secondary">
														{tm.home_wins}-{tm.away_wins} · {teamName(tm.winner_team_id)} win
													</Badge>
												{:else}
													<Badge variant="outline">{tm.home_wins}-{tm.away_wins}</Badge>
												{/if}
											</div>
											<ul class="flex flex-col gap-2">
												{#each tm.matches as m (m.id)}
													{@const decided = m.status === 2 || m.status === 3}
													<li class="flex flex-wrap items-center gap-3 text-sm">
														<span class="w-44 shrink-0">
															<span class="font-medium">{teamName(m.team_id)}</span>
															<span class="block text-xs text-muted-foreground">
																{matchPlayerLabel(m)}
															</span>
														</span>
														{#if isCommissioner && !decided}
															<div class="flex flex-wrap items-center gap-2">
																<Label for={`${m.id}-main`} class="text-xs">Score</Label>
																<Input
																	id={`${m.id}-main`}
																	type="number"
																	class="w-16"
																	value={scoreDrafts[m.id]?.main ?? String(m.main_value)}
																	oninput={(e) =>
																		setScoreDraft(
																			m.id,
																			'main',
																			(e.currentTarget as HTMLInputElement).value
																		)
																	}
																/>
																<Label for={`${m.id}-secondary`} class="text-xs">
																	Secondary
																</Label>
																<Input
																	id={`${m.id}-secondary`}
																	type="number"
																	class="w-16"
																	value={scoreDrafts[m.id]?.secondary ?? String(m.secondary_value)}
																	oninput={(e) =>
																		setScoreDraft(
																			m.id,
																			'secondary',
																			(e.currentTarget as HTMLInputElement).value
																		)
																	}
																/>
																<Button
																	type="button"
																	variant="outline"
																	size="sm"
																	disabled={savingScore === m.id}
																	onclick={() => saveScore(week.id, m.id)}
																>
																	{savingScore === m.id ? 'Saving…' : 'Save'}
																</Button>
																<Button
																	type="button"
																	size="sm"
																	disabled={completingMatch === m.id}
																	onclick={() => handleCompleteMatch(week.id, m.id)}
																>
																	{completingMatch === m.id ? 'Completing…' : 'Complete'}
																</Button>
															</div>
														{:else}
															<span class="text-muted-foreground">
																{m.main_value}
																{m.secondary_value ? `/${m.secondary_value}` : ''} ·{' '}
																{matchStatusLabel(m.status)}
															</span>
														{/if}
													</li>
												{/each}
											</ul>
										</li>
									{/each}
								</ul>
							{/if}
						</li>
					{/each}
				</ul>
			{/if}
		</CardContent>
	</Card>

	<!-- Standings -->
	<Standings {standings} {teamName} />


	<!-- Co-commissioners (sysadmin only) -->
	{#if isSysAdmin}
		<Card class="mt-6">
			<CardHeader>
				<CardTitle class="text-base">Co-commissioners</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="mb-4 text-sm text-muted-foreground">
					Additional users who can help administer this season.
				</p>

				{#if commissionerError}
					<p class="mb-4 text-sm font-medium text-destructive">{commissionerError}</p>
				{/if}

				<form
					class="flex flex-wrap items-end gap-3"
					onsubmit={(e) => {
						e.preventDefault();
						handleAddSeasonCommissioner();
					}}
				>
					<div class="flex flex-col gap-1">
						<Label for="co-commissioner-user" class="text-xs">User</Label>
						<NativeSelect
							id="co-commissioner-user"
							value={commissionerUserId}
							oninput={(e) =>
								(commissionerUserId = (
									e.currentTarget as HTMLSelectElement
								).value)}
						>
							<NativeSelectOption value="" disabled>Select a user…</NativeSelectOption>
							{#each commissionerOptions as u (u.id)}
								<NativeSelectOption value={u.id}>{fullName(u)}</NativeSelectOption>
							{/each}
						</NativeSelect>
					</div>
					<Button
						type="submit"
						disabled={savingCommissioner || !commissionerUserId}
					>
						{savingCommissioner ? 'Saving…' : 'Add co-commissioner'}
					</Button>
				</form>

				{#if commissionerNames.length === 0}
					<p class="mt-4 text-sm text-muted-foreground">No co-commissioners yet.</p>
				{:else}
					<ul class="mt-4 space-y-2 text-sm">
						{#each commissionerNames as sc (sc.id)}
							<li class="flex items-center justify-between gap-3">
								<span class="font-medium">{userName(sc.user_id)}</span>
								<Button
									type="button"
									variant="ghost"
									size="sm"
									disabled={savingCommissioner}
									onclick={() => handleRemoveSeasonCommissioner(sc.id)}
								>
									Remove
								</Button>
							</li>
						{/each}
					</ul>
				{/if}
			</CardContent>
		</Card>
	{/if}


	<!-- Late additions (sysadmin only) -->
	{#if isSysAdmin}
		<Card class="mt-6">
			<CardHeader>
				<CardTitle class="text-base">Late additions</CardTitle>
			</CardHeader>
			<CardContent>
				<p class="mb-4 text-sm text-muted-foreground">
					Players added to this season after the draft was completed.
				</p>

				{#if lateAdditionError}
					<p class="mb-4 text-sm font-medium text-destructive">{lateAdditionError}</p>
				{/if}

				<form
					class="flex flex-wrap items-end gap-3"
					onsubmit={(e) => {
						e.preventDefault();
						handleAddLateAddition();
					}}
				>
					<div class="flex flex-col gap-1">
						<Label for="late-addition-user" class="text-xs">Player</Label>
						<NativeSelect
							id="late-addition-user"
							value={lateAdditionUserId}
							oninput={(e) =>
								(lateAdditionUserId = (
									e.currentTarget as HTMLSelectElement
								).value)}
						>
							<NativeSelectOption value="" disabled>Select a player…</NativeSelectOption>
							{#each lateAdditionOptions as u (u.id)}
								<NativeSelectOption value={u.id}>{fullName(u)}</NativeSelectOption>
							{/each}
						</NativeSelect>
					</div>
					<Button
						type="submit"
						disabled={savingLateAddition || !lateAdditionUserId}
					>
						{savingLateAddition ? 'Saving…' : 'Add late player'}
					</Button>
				</form>

				{#if lateAdditionNames.length === 0}
					<p class="mt-4 text-sm text-muted-foreground">No late additions yet.</p>
				{:else}
					<ul class="mt-4 space-y-2 text-sm">
						{#each lateAdditionNames as la (la.id)}
							<li class="flex items-center justify-between gap-3">
								<span class="font-medium">{userName(la.user_id)}</span>
								<Button
									type="button"
									variant="ghost"
									size="sm"
									disabled={savingLateAddition}
									onclick={() => handleRemoveLateAddition(la.id)}
								>
									Remove
								</Button>
							</li>
						{/each}
					</ul>
				{/if}
			</CardContent>
		</Card>
	{/if}

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
