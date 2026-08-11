<script lang="ts">
	import { isLoggedIn, sessionExpired, SESSION_EXPIRED_MESSAGE } from '$lib/auth';
	import LoginForm from '$lib/components/LoginForm.svelte';

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
	<h1 class="text-2xl font-semibold tracking-tight">IntraClub</h1>
	{#if $sessionExpired}
		<p class="mt-2 text-sm font-semibold text-destructive" role="status">{SESSION_EXPIRED_MESSAGE}</p>
	{/if}
	<p class="mt-2 text-muted-foreground">Welcome! Log in to manage your club.</p>
	<div class="mt-4">
		<LoginForm />
	</div>
{/if}
