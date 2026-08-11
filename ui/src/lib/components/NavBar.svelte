<script lang="ts">
	import { isLoggedIn, clearToken } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';

	let loggedIn = $state(isLoggedIn());

	function logout() {
		clearToken();
		loggedIn = false;
		goto('/');
	}
</script>

<header class="sticky top-0 z-40 w-full border-b bg-background">
	<div class="mx-auto flex h-14 max-w-6xl items-center gap-6 px-6">
		<a href="/" class="text-base font-semibold tracking-tight">IntraClub</a>
		<nav class="flex items-center gap-1 text-sm">
			<a href="/facilities" class="rounded-md px-2 py-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">Facilities</a>
			<a href="/formats" class="rounded-md px-2 py-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">Formats</a>
			<a href="/ratings" class="rounded-md px-2 py-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">Ratings</a>
			<a href="/rulesets" class="rounded-md px-2 py-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">Rulesets</a>
			<a href="/scoring-structures" class="rounded-md px-2 py-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">Scoring Structures</a>
			<a href="/playoff-structures" class="rounded-md px-2 py-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">Playoff Structures</a>
			<a href="/photos" class="rounded-md px-2 py-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">Photos</a>
		</nav>
		<div class="ml-auto flex items-center">
			{#if loggedIn}
				<Button variant="ghost" onclick={logout}>Log out</Button>
			{:else}
				<a href="/login" class="text-sm text-muted-foreground transition-colors hover:text-foreground">Log in</a>
			{/if}
		</div>
	</div>
</header>
