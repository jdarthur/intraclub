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
	<h1>Welcome to IntraClub</h1>
	<p>Head to <a href="/facilities">Facilities</a>, <a href="/formats">Formats</a>, <a href="/ratings">Ratings</a>, <a href="/rulesets">Rulesets</a>, <a href="/scoring-structures">Scoring Structures</a>, or <a href="/playoff-structures">Playoff Structures</a> to get started.</p>
{:else}
	<h1>IntraClub</h1>
	{#if $sessionExpired}
		<p class="session-expired" role="status">{SESSION_EXPIRED_MESSAGE}</p>
	{/if}
	<p>Welcome! Log in to manage your club.</p>
	<LoginForm />
{/if}

<style>
	.session-expired {
		color: #c00;
		font-weight: 600;
	}
</style>
