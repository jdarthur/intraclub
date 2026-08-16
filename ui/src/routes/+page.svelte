<script lang="ts">
	import { isLoggedIn, sessionExpired, SESSION_EXPIRED_MESSAGE } from '$lib/auth';
	import LoginForm from '$lib/components/LoginForm.svelte';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card/index.js';
	import DraftingCompassIcon from '@lucide/svelte/icons/drafting-compass';
	import CalendarDaysIcon from '@lucide/svelte/icons/calendar-days';
	import TrophyIcon from '@lucide/svelte/icons/trophy';

	let loggedIn = $state(false);

	// Keep the logged-in state in sync with the session: the subscription fires
	// immediately with the current store value and again when a session expires,
	// at which point the token has been cleared so isLoggedIn() is false.
	$effect(() => {
		const unsub = sessionExpired.subscribe(() => {
			loggedIn = isLoggedIn();
		});
		return unsub;
	});
</script>

<svelte:head>
	<title>IntraClub</title>
</svelte:head>

{#if loggedIn}
	<h1 class="text-2xl font-semibold tracking-tight">Welcome to IntraClub</h1>
	<p class="mt-2 text-muted-foreground">Head to <a href="/facilities" class="text-primary underline-offset-4 hover:underline">Facilities</a>, <a href="/formats" class="text-primary underline-offset-4 hover:underline">Formats</a>, <a href="/ratings" class="text-primary underline-offset-4 hover:underline">Ratings</a>, <a href="/rulesets" class="text-primary underline-offset-4 hover:underline">Rulesets</a>, <a href="/scoring-structures" class="text-primary underline-offset-4 hover:underline">Scoring Structures</a>, or <a href="/playoff-structures" class="text-primary underline-offset-4 hover:underline">Playoff Structures</a> to get started.</p>
{:else}
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
					<DraftingCompassIcon class="size-6 text-primary" aria-hidden />
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
					<CalendarDaysIcon class="size-6 text-primary" aria-hidden />
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
					<TrophyIcon class="size-6 text-primary" aria-hidden />
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
{/if}
