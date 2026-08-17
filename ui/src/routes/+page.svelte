<script lang="ts">
	import type { Component } from 'svelte';
	import { Async } from '$lib/async.svelte';
	import { getCurrentUserId, isLoggedIn, sessionExpired, SESSION_EXPIRED_MESSAGE } from '$lib/auth';
	import { identity } from '$lib/identity.svelte';
	import LoginForm from '$lib/components/LoginForm.svelte';
	import { listSeasons, listSeasonTeams } from '$lib/season';
	import type { Season, SeasonTeam } from '$lib/season';
	import { listWeeksForSeason } from '$lib/week';
	import type { Week } from '$lib/week';
	import { listSeasonCommissioners } from '$lib/seasonCommissioner';
	import type { SeasonCommissioner } from '$lib/seasonCommissioner';
	import { listDraftCaptains } from '$lib/draftCaptain';
	import type { DraftCaptain } from '$lib/draftCaptain';
	import { getStandings, getWeekMatches } from '$lib/match';
	import type { StandingsEntry, WeekMatchDetail } from '$lib/match';
	import { getScheduleForSeason } from '$lib/schedule';
	import type { ScheduleDetail } from '$lib/schedule';
	import {
		AVAILABILITY_AVAILABLE,
		AVAILABILITY_UNSET,
		getAvailabilityForUser,
		setAvailability
	} from '$lib/availability';
	import type { Availability } from '$lib/availability';
	import { getLineupDetail } from '$lib/lineup';
	import type { LineupDetail } from '$lib/lineup';
	import { getProposalDetail, listProposals } from '$lib/commissionerProposal';
	import type { ProposalDetail } from '$lib/commissionerProposal';
	import { AsyncSection, EmptyState, StatCard } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card/index.js';
	import { Progress } from '$lib/components/ui/progress/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import Standings from '$lib/components/Standings.svelte';
	import DraftingCompassIcon from '@lucide/svelte/icons/drafting-compass';
	import CalendarDaysIcon from '@lucide/svelte/icons/calendar-days';
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import ClipboardListIcon from '@lucide/svelte/icons/clipboard-list';
	import FlagIcon from '@lucide/svelte/icons/flag';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import SwordsIcon from '@lucide/svelte/icons/swords';
	import TrophyIcon from '@lucide/svelte/icons/trophy';
	import UsersIcon from '@lucide/svelte/icons/users';
	import VoteIcon from '@lucide/svelte/icons/vote';

	let loggedIn = $state(isLoggedIn());
	let userId = $state<string | null>(getCurrentUserId());

	// Keep the signed-in state in sync with the session: the subscription fires
	// immediately with the current store value and again when a session expires
	// or a fresh login lands, at which point the token has been updated.
	$effect(() => {
		const unsub = sessionExpired.subscribe(() => {
			loggedIn = isLoggedIn();
			userId = getCurrentUserId();
		});
		return unsub;
	});

	// ---- wave 0 / wave 1 / wave 2 state ------------------------------------
	// The dashboard loads in three finite waves (no polling, no auto-refresh):
	//   1. the list calls every card hangs off (parallel, allSettled)
	//   2. one weeks call per season (bounded — there are only a handful)
	//   3. one independent Async per card (see below)
	// Everything is gated on `userId`: with a token that has no decodable `sub`
	// claim (or no token at all) zero fetches are issued, so a 401 can never
	// tear down a session mid-render.
	let seasons = $state<Season[]>([]);
	let seasonTeams = $state<SeasonTeam[]>([]);
	let seasonCommissioners = $state<SeasonCommissioner[]>([]);
	let draftCaptains = $state<DraftCaptain[]>([]);
	let weeksBySeason = $state<Record<string, Week[]>>({});
	let wave1Done = $state(false);
	let wave2Done = $state(false);
	let waveError = $state('');

	let loadStarted = false;
	$effect(() => {
		if (userId === null || loadStarted) return;
		loadStarted = true;
		void loadDashboard();
	});

	async function loadDashboard(): Promise<void> {
		// Wave 1 — the three calls the ticket prescribes (season / team /
		// season_team) plus two public join tables: season_commissioner resolves
		// commissioner-ness without waiting on the schedule call, and
		// draft_captain names every team ("Team N" in draft order) for the
		// standings / matchups resolvers. All are readable by everyone, so this
		// stays a single parallel wave.
		const [seasonsRes, teamsRes, seasonTeamsRes, commissionersRes, captainsRes] =
			await Promise.allSettled([
				listSeasons(),
				identity.loadTeams(),
				listSeasonTeams(),
				listSeasonCommissioners(),
				listDraftCaptains()
			]);
		wave1Done = true;
		if (seasonsRes.status === 'fulfilled') seasons = seasonsRes.value;
		if (seasonTeamsRes.status === 'fulfilled') seasonTeams = seasonTeamsRes.value;
		if (commissionersRes.status === 'fulfilled') seasonCommissioners = commissionersRes.value;
		if (captainsRes.status === 'fulfilled') draftCaptains = captainsRes.value;
		waveError = [seasonsRes, seasonTeamsRes, commissionersRes, captainsRes]
			.filter((r): r is PromiseRejectedResult => r.status === 'rejected')
			.map((r) => (r.reason instanceof Error ? r.reason.message : String(r.reason)))
			.join('; ');

		// Wave 2 — one weeks call per season. With no seasons this issues zero
		// calls, which is the default path in dev and CI.
		const weekResults = await Promise.allSettled(seasons.map((s) => listWeeksForSeason(s.id)));
		const bySeason: Record<string, Week[]> = {};
		weekResults.forEach((r, i) => {
			if (r.status === 'fulfilled') bySeason[seasons[i].id] = r.value;
		});
		weeksBySeason = bySeason;
		wave2Done = true;
		const weekErrors = weekResults
			.filter((r): r is PromiseRejectedResult => r.status === 'rejected')
			.map((r) => (r.reason instanceof Error ? r.reason.message : String(r.reason)));
		waveError = [waveError, ...weekErrors].filter(Boolean).join('; ');
	}

	// ---- season inference ---------------------------------------------------
	// There is no "current season" concept in the backend: a Season has no
	// start date, end date or active flag, and GET /api/season returns every
	// season. So the current season is inferred from its weeks (a genuine
	// modelling gap, documented here):
	//   candidates    = seasons having >= 1 week
	//   currentSeason = candidates.find(minWeekDate <= now <= maxWeekDate)
	//                ?? candidates.sortBy(maxWeekDate, desc)[0]
	//   currentWeek   = weeks.find(w => !w.closed) ?? weeks.at(-1)
	//   weekNumber    = weeks.indexOf(currentWeek) + 1
	const currentSeason = $derived.by(() => {
		const candidates = seasons.filter((s) => (weeksBySeason[s.id]?.length ?? 0) > 0);
		if (candidates.length === 0) return null;
		const now = Date.now();
		const spanning = candidates.find((s) => {
			const times = (weeksBySeason[s.id] ?? []).map((w) => new Date(w.date).getTime());
			return now >= Math.min(...times) && now <= Math.max(...times);
		});
		if (spanning) return spanning;
		return [...candidates].sort((a, b) => {
			const maxA = Math.max(
				...(weeksBySeason[a.id] ?? []).map((w) => new Date(w.date).getTime())
			);
			const maxB = Math.max(
				...(weeksBySeason[b.id] ?? []).map((w) => new Date(w.date).getTime())
			);
			return maxB - maxA;
		})[0];
	});
	const seasonWeeks = $derived(currentSeason ? weeksBySeason[currentSeason.id] ?? [] : []);
	const currentWeek = $derived(seasonWeeks.find((w) => !w.closed) ?? seasonWeeks.at(-1) ?? null);
	const weekNumber = $derived(currentWeek ? seasonWeeks.indexOf(currentWeek) + 1 : 0);
	const totalWeeks = $derived(seasonWeeks.length);

	const isCommissioner = $derived.by(() => {
		if (userId === null || currentSeason === null) return false;
		return seasonCommissioners.some(
			(sc) => sc.season_id === currentSeason.id && sc.user_id === userId
		);
	});

	const myCaptainTeamIds = $derived.by(() => {
		if (userId === null) return [] as string[];
		return identity.teams
			.filter((r) =>
				r.assignments.some(
					(a) => a.user_id === userId && (a.role === 'captain' || a.role === 'co_captain')
				)
			)
			.map((r) => r.team.id);
	});

	const seasonTeamIds = $derived(
		currentSeason
			? seasonTeams.filter((st) => st.season_id === currentSeason.id).map((st) => st.team_id)
			: []
	);

	// Teams are named "Team N" in draft order (see model/draft.go Initialize);
	// the user's own rosters carry the real names via GET /api/team.
	function teamName(teamId: string): string {
		const mine = identity.teams.find((r) => r.team.id === teamId);
		if (mine) return mine.team.name;
		const captain = draftCaptains.find((c) => c.team_id === teamId);
		return captain ? `Team ${captain.draft_order + 1}` : teamId;
	}

	// ---- wave 3 — one independent Async per card ----------------------------
	const standingsAsync = new Async<StandingsEntry[]>();
	const scheduleAsync = new Async<ScheduleDetail>();
	const matchesAsync = new Async<WeekMatchDetail>();
	const availabilityAsync = new Async<Availability[]>();
	const lineupAsync = new Async<LineupDetail[]>();
	const votesAsync = new Async<ProposalDetail[]>();

	let wave3Key = '';
	$effect(() => {
		const season = currentSeason;
		const week = currentWeek;
		if (season === null || week === null) return;
		const key = `${season.id}:${week.id}`;
		if (wave3Key === key) return;
		wave3Key = key;

		standingsAsync.run(() => getStandings(season.id));
		scheduleAsync.run(() =>
			getScheduleForSeason(season.id).then((detail) => {
				// Write-through cache: keep the shared identity store's
				// commissioner data in sync (see $lib/identity.svelte.ts).
				identity.noteCommissioners(season.id, detail.commissioners);
				return detail;
			})
		);
		matchesAsync.run(() => getWeekMatches(week.id));
		availabilityAsync.run(() => getAvailabilityForUser(season.draft_id));
		lineupAsync.run(() => loadLineupDetails(season, week));
		votesAsync.run(loadPendingVotes);
	});

	// Only fetch lineup detail for teams the user can act on: their own captain
	// teams, or every team in the season when they are a commissioner (to spot
	// lineups that aren't official yet).
	async function loadLineupDetails(season: Season, week: Week): Promise<LineupDetail[]> {
		const teamIds = isCommissioner
			? seasonTeamIds
			: myCaptainTeamIds.filter((tid) => seasonTeamIds.includes(tid));
		const results = await Promise.allSettled(teamIds.map((tid) => getLineupDetail(tid, week.id)));
		return results
			.filter((r): r is PromiseFulfilledResult<LineupDetail> => r.status === 'fulfilled')
			.map((r) => r.value);
	}

	// A proposal needs the user's vote iff they are an eligible voter
	// (commissioner or captain of the season), haven't voted yet, and the
	// proposal hasn't been decided.
	async function loadPendingVotes(): Promise<ProposalDetail[]> {
		const proposals = await listProposals();
		const results = await Promise.allSettled(proposals.map((p) => getProposalDetail(p.id)));
		return results
			.filter((r): r is PromiseFulfilledResult<ProposalDetail> => r.status === 'fulfilled')
			.map((r) => r.value)
			.filter(
				(d) =>
					userId !== null &&
					d.voters.includes(userId) &&
					d.my_vote === undefined &&
					!d.accepted &&
					!d.rejected
			);
	}

	// ---- "needs you" action list -------------------------------------------
	interface NeedsYouItem {
		id: string;
		icon: Component;
		title: string;
		description: string;
		href?: string;
		actionLabel?: string;
		onAction?: () => void;
		progressValue?: number;
		progressMax?: number;
		progressLabel?: string;
	}

	function weekLabel(week: Week): string {
		const d = new Date(week.date);
		return Number.isNaN(d.getTime()) ? week.date : d.toLocaleDateString();
	}

	function availabilityItems(): NeedsYouItem[] {
		if (!currentSeason) return [];
		const items: NeedsYouItem[] = [];
		const fromIndex = currentWeek ? seasonWeeks.indexOf(currentWeek) : 0;
		for (let i = fromIndex; i < seasonWeeks.length; i++) {
			const week = seasonWeeks[i];
			const entry = availabilityAsync.data?.find((a) => a.week_id === week.id);
			if (entry && entry.available !== AVAILABILITY_UNSET) continue;
			items.push({
				id: `availability-${week.id}`,
				icon: CalendarDaysIcon,
				title: `Set your availability for week ${i + 1}`,
				description: `No availability set for ${weekLabel(week)}.`,
				actionLabel: 'Mark me available',
				onAction: () => {
					void markAvailable(week.id);
				}
			});
		}
		return items;
	}

	async function markAvailable(weekId: string): Promise<void> {
		if (!currentSeason) return;
		try {
			await setAvailability(weekId, AVAILABILITY_AVAILABLE);
			await availabilityAsync.run(() => getAvailabilityForUser(currentSeason.draft_id));
		} catch {
			// leave the item in place; the next reload reflects the saved value
		}
	}

	function lineupItems(): NeedsYouItem[] {
		const details = (lineupAsync.data ?? []).filter((d) => myCaptainTeamIds.includes(d.team_id));
		const items: NeedsYouItem[] = [];
		for (const detail of details) {
			const filled = detail.pairings.length;
			const total = detail.lines.length;
			if (detail.lineup && detail.lineup.confirmed && filled >= total) continue;
			let description: string;
			if (!detail.lineup) description = 'No lineup set yet.';
			else if (!detail.lineup.confirmed) description = 'Lineup not confirmed yet.';
			else description = `${filled} of ${total} lines filled.`;
			items.push({
				id: `lineup-${detail.team_id}`,
				icon: ClipboardListIcon,
				title: `Set ${teamName(detail.team_id)}'s lineup for week ${weekNumber}`,
				description,
				href: `/seasons/${currentSeason?.id}`,
				actionLabel: 'Set lineup',
				progressValue: filled,
				progressMax: total,
				progressLabel: `${filled} of ${total} lines`
			});
		}
		return items;
	}

	function voteItems(): NeedsYouItem[] {
		const pending = votesAsync.data ?? [];
		if (pending.length === 0) return [];
		return [
			{
				id: 'votes',
				icon: VoteIcon,
				title: `${pending.length} proposal${pending.length === 1 ? '' : 's'} awaiting your vote`,
				description: "Your vote hasn't been cast yet.",
				href: `/seasons/${currentSeason?.id}/proposals`,
				actionLabel: 'Vote'
			}
		];
	}

	function commissionerItems(): NeedsYouItem[] {
		if (!isCommissioner || !currentSeason || !currentWeek) return [];
		const items: NeedsYouItem[] = [];
		const week = currentWeek;
		const seasonHref = `/seasons/${currentSeason.id}`;

		const weeklyMatchup = scheduleAsync.data?.weekly_matchups.find(
			(wm) => wm.week_id === week.id
		);
		if (!weeklyMatchup || weeklyMatchup.matchups.length === 0) {
			items.push({
				id: 'matchups',
				icon: SwordsIcon,
				title: `No matchups assigned for week ${weekNumber}`,
				description: 'Assign weekly matchups so the week can be scheduled.',
				href: seasonHref,
				actionLabel: 'Open season'
			});
		}

		const matches = matchesAsync.data;
		if (matches && matches.team_matches.length === 0) {
			items.push({
				id: 'matches',
				icon: TrophyIcon,
				title: `No matches generated for week ${weekNumber}`,
				description: "Generate this week's matches from the official lineups.",
				href: seasonHref,
				actionLabel: 'Open season'
			});
		}

		const notOfficial = (lineupAsync.data ?? []).filter((d) => d.lineup && !d.lineup.official);
		if (notOfficial.length > 0) {
			items.push({
				id: 'lineups',
				icon: ShieldCheckIcon,
				title: `${notOfficial.length} lineup${notOfficial.length === 1 ? '' : 's'} not yet official`,
				description: 'Mark confirmed lineups as official before generating matches.',
				href: seasonHref,
				actionLabel: 'Open season'
			});
		}

		if (
			matches &&
			matches.team_matches.length > 0 &&
			matches.team_matches.every((tm) => tm.complete) &&
			!week.closed
		) {
			items.push({
				id: 'close-week',
				icon: FlagIcon,
				title: `Week ${weekNumber} is complete — close it`,
				description: 'All team matches are finished; closing the week finalises the standings.',
				href: seasonHref,
				actionLabel: 'Open season'
			});
		}
		return items;
	}

	const needsYouItems = $derived([
		...availabilityItems(),
		...lineupItems(),
		...voteItems(),
		...commissionerItems()
	]);

	// Before wave 3 has fired the Asyncs are idle — treat that as loading so the
	// "Nothing needs you" positive state never flashes prematurely.
	const wave3Started = $derived(
		standingsAsync.status !== 'idle' ||
			scheduleAsync.status !== 'idle' ||
			matchesAsync.status !== 'idle' ||
			availabilityAsync.status !== 'idle' ||
			lineupAsync.status !== 'idle' ||
			votesAsync.status !== 'idle'
	);
	const needsYouLoading = $derived(
		!wave3Started ||
			availabilityAsync.status === 'loading' ||
			lineupAsync.status === 'loading' ||
			votesAsync.status === 'loading' ||
			scheduleAsync.status === 'loading' ||
			matchesAsync.status === 'loading'
	);
	const needsYouAllErrored = $derived(
		availabilityAsync.status === 'error' &&
			lineupAsync.status === 'error' &&
			votesAsync.status === 'error' &&
			scheduleAsync.status === 'error' &&
			matchesAsync.status === 'error'
	);

	const pendingVotes = $derived(votesAsync.data?.length ?? 0);

	const currentMatchups = $derived(
		currentWeek
			? (scheduleAsync.data?.weekly_matchups.find((wm) => wm.week_id === currentWeek.id) ??
				null)
			: null
	);
</script>

<svelte:head>
	<title>IntraClub</title>
</svelte:head>

{#if !loggedIn}
	<section class="mx-auto max-w-2xl pt-4 text-center">
		<h1 class="text-4xl font-bold tracking-tight sm:text-5xl">IntraClub</h1>
		<p class="mt-4 text-lg text-muted-foreground">
			Run your intra-club season from draft to trophy — snake-order drafts,
			weekly matches, per-line scoring, and playoff standings, all under one
			roof.
		</p>
	</section>

	{#if $sessionExpired}
		<p class="mt-6 text-center text-sm font-semibold text-destructive" role="status">
			{SESSION_EXPIRED_MESSAGE}
		</p>
	{/if}

	<div class="mx-auto mt-8 max-w-sm">
		<Card>
			<CardHeader>
				<CardTitle level={2} class="text-center">Welcome to your club</CardTitle>
				<CardDescription class="text-center">
					Enter your email and we'll send you a login link.
				</CardDescription>
			</CardHeader>
			<CardContent>
				<LoginForm />
			</CardContent>
		</Card>
	</div>

	<section class="mt-16">
		<h2 class="text-center text-2xl font-semibold tracking-tight">How it works</h2>
		<div class="mt-8 grid gap-6 sm:grid-cols-3">
			<Card>
				<CardHeader>
					<DraftingCompassIcon class="size-6 text-primary dark:text-brand" aria-hidden />
					<CardTitle>Draft</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="text-muted-foreground">
						Grade players before the draft, then pick in snake order to build
						balanced teams.
					</p>
				</CardContent>
			</Card>
			<Card>
				<CardHeader>
					<CalendarDaysIcon class="size-6 text-primary dark:text-brand" aria-hidden />
					<CardTitle>Season</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="text-muted-foreground">
						Players share weekly availability; captains set lineups against
						your club's format.
					</p>
				</CardContent>
			</Card>
			<Card>
				<CardHeader>
					<TrophyIcon class="size-6 text-primary dark:text-brand" aria-hidden />
					<CardTitle>Score</CardTitle>
				</CardHeader>
				<CardContent>
					<p class="text-muted-foreground">
						Per-line scoring, composite structures and tie-breakers decide
						matches, standings and the playoffs.
					</p>
				</CardContent>
			</Card>
		</div>
		<p class="mt-8 text-center text-sm text-muted-foreground">
			Club rules live in an amendable ruleset — commissioners propose changes
			and captains vote.
		</p>
	</section>
{:else if userId === null}
	<!-- A token with no decodable `sub` claim (e.g. an exp-only token): render a
	     neutral signed-in state and issue zero fetches — any request would 401
	     and tear down the session. -->
	<h1 class="text-2xl font-semibold tracking-tight">Welcome back</h1>
	<p class="mt-2 text-muted-foreground">Restoring your session…</p>
{:else}
	<h1 class="text-2xl font-semibold tracking-tight">
		Welcome back{identity.displayName ? `, ${identity.displayName}` : ''}
	</h1>

	{#if waveError}
		<p class="mt-2 text-sm font-medium text-destructive" role="status">{waveError}</p>
	{/if}

	{#if !wave1Done}
		<div class="mt-6 grid gap-4 sm:grid-cols-3">
			<Skeleton class="h-24" />
			<Skeleton class="h-24" />
			<Skeleton class="h-24" />
		</div>
		<Skeleton class="mt-6 h-40" />
	{:else}
		<!-- Your teams — renders off wave 1 alone, never waits on season
		     inference. -->
		<Card class="mt-6">
			<CardHeader>
				<CardTitle class="text-base">Your teams</CardTitle>
			</CardHeader>
			<CardContent>
				{#if identity.teams.length === 0}
					<p class="text-sm text-muted-foreground">You're not on a team yet.</p>
				{:else}
					<ul class="flex flex-wrap gap-x-4 gap-y-2">
						{#each identity.teams as roster}
							<li>
								<a
									href={`/teams/${roster.team.id}`}
									class="text-sm font-medium text-primary underline-offset-4 hover:underline dark:text-brand"
								>
									{roster.team.name}
								</a>
							</li>
						{/each}
					</ul>
				{/if}
			</CardContent>
		</Card>

		{#if !wave2Done}
			<div class="mt-6 grid gap-4 sm:grid-cols-3">
				<Skeleton class="h-24" />
				<Skeleton class="h-24" />
				<Skeleton class="h-24" />
			</div>
			<Skeleton class="mt-6 h-40" />
			<Skeleton class="mt-6 h-40" />
		{:else if !currentSeason}
			<!-- No season: collapse the whole season region into one empty state
			     rather than rendering "Week 0 of 0" tiles. -->
			<div class="mt-6">
				<EmptyState
					title="No season yet"
					description="Once a draft is finalised into a season, your current week, standings and matchups appear here."
					actionHref="/drafts"
					actionLabel="Start a draft"
				/>
			</div>
		{:else}
			<!-- Stat tiles -->
			<div class="mt-6 grid gap-4 sm:grid-cols-3">
				<StatCard
					label="Current week"
					value={weekNumber}
					hint={`of ${totalWeeks} weeks`}
					icon={CalendarDaysIcon}
				/>
				<StatCard
					label="Your teams"
					value={identity.teams.length}
					hint={identity.teams.length === 1 ? 'team' : 'teams'}
					icon={UsersIcon}
				/>
				<StatCard
					label="Pending votes"
					value={pendingVotes}
					hint="awaiting your vote"
					icon={VoteIcon}
					loading={!wave3Started || votesAsync.status === 'loading'}
				/>
			</div>

			<!-- Needs you -->
			<Card class="mt-6">
				<CardHeader>
					<CardTitle class="text-base">Needs you</CardTitle>
				</CardHeader>
				<CardContent>
					{#if needsYouLoading}
						<Skeleton class="h-8 w-full" />
						<Skeleton class="mt-3 h-8 w-full" />
					{:else if needsYouAllErrored}
						<Alert variant="destructive">
							<AlertDescription>Couldn't load your action items.</AlertDescription>
						</Alert>
					{:else if needsYouItems.length === 0}
						<div class="flex items-center gap-2 text-sm text-muted-foreground">
							<CircleCheckIcon class="size-4 text-success" aria-hidden />
							Nothing needs you right now
						</div>
					{:else}
						<ul class="flex flex-col gap-4">
							{#each needsYouItems as item}
								<li>
									<div class="flex items-start gap-3">
										<item.icon class="mt-0.5 size-4 shrink-0 text-muted-foreground" aria-hidden />
										<div class="min-w-0 flex-1">
											<p class="text-sm font-medium">{item.title}</p>
											<p class="mt-0.5 text-sm text-muted-foreground">{item.description}</p>
											{#if item.progressValue !== undefined && item.progressMax !== undefined}
												<div class="mt-2 flex items-center gap-2">
													<Progress
														value={item.progressValue}
														max={item.progressMax}
														class="h-1.5 w-40"
													/>
													<span class="text-xs text-muted-foreground">
														{item.progressLabel}
													</span>
												</div>
											{/if}
										</div>
										{#if item.href && item.actionLabel}
											<Button href={item.href} variant="outline" size="sm" class="shrink-0">
												{item.actionLabel}
											</Button>
										{:else if item.actionLabel}
											<Button onclick={item.onAction} variant="outline" size="sm" class="shrink-0">
												{item.actionLabel}
											</Button>
										{/if}
									</div>
								</li>
							{/each}
						</ul>
					{/if}
				</CardContent>
			</Card>

			<!-- Standings -->
			<AsyncSection state={standingsAsync}>
				{#snippet children(standings)}
					<Standings {standings} {teamName} />
				{/snippet}
			</AsyncSection>

			<!-- This week's matchups + scores -->
			<div class="mt-6 grid gap-6 lg:grid-cols-2">
				<Card>
					<CardHeader>
						<CardTitle class="text-base">This week's matchups</CardTitle>
					</CardHeader>
					<CardContent>
						<AsyncSection state={scheduleAsync}>
							{#snippet children(detail)}
								{#if !currentMatchups || currentMatchups.matchups.length === 0}
									<p class="text-sm text-muted-foreground">No matchups assigned yet.</p>
								{:else}
									<ul class="flex flex-col gap-1 text-sm">
										{#each currentMatchups.matchups as m}
											<li>
												{#if m.bye}
													{teamName(m.home_team_id)} — bye
												{:else}
													{teamName(m.home_team_id)} vs {teamName(m.away_team_id)}
												{/if}
											</li>
										{/each}
									</ul>
								{/if}
							{/snippet}
						</AsyncSection>
					</CardContent>
				</Card>

				<Card>
					<CardHeader>
						<CardTitle class="text-base">This week's scores</CardTitle>
					</CardHeader>
					<CardContent>
						<AsyncSection state={matchesAsync}>
							{#snippet children(detail)}
								{#if detail.team_matches.length === 0}
									<p class="text-sm text-muted-foreground">No matches generated yet.</p>
								{:else}
									<ul class="flex flex-col gap-1 text-sm">
										{#each detail.team_matches as tm}
											<li>
												{teamName(tm.home_team_id)} vs {teamName(tm.away_team_id)}
												{#if tm.complete}
													— {tm.home_wins}-{tm.away_wins}
												{:else}
													— in progress
												{/if}
											</li>
										{/each}
									</ul>
								{/if}
							{/snippet}
						</AsyncSection>
					</CardContent>
				</Card>
			</div>
		{/if}
	{/if}
{/if}
