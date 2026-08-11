<script lang="ts">
	import { isLoggedIn } from '$lib/auth';
	import LoginForm from '$lib/components/LoginForm.svelte';

	let loggedIn = $state(isLoggedIn());
</script>

<svelte:head>
	<title>IntraClub</title>
</svelte:head>

{#if loggedIn}
	<h1>Welcome to IntraClub</h1>
	<p>Head to <a href="/facilities">Facilities</a>, <a href="/formats">Formats</a>, <a href="/ratings">Ratings</a>, <a href="/rulesets">Rulesets</a>, <a href="/scoring-structures">Scoring Structures</a>, or <a href="/playoff-structures">Playoff Structures</a> to get started.</p>
{:else}
	<h1>IntraClub</h1>
	<p>Welcome! Log in to manage your club.</p>
	<LoginForm />
{/if}
