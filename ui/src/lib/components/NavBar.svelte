<script lang="ts">
	import { isLoggedIn, clearToken } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import {
		Popover,
		PopoverContent,
		PopoverTrigger
	} from '$lib/components/ui/popover/index.js';
	import logo from '$lib/assets/favicon.svg';

	let loggedIn = $state(isLoggedIn());

	const settingsLinks = [
		{ href: '/seasons', label: 'Seasons' },
		{ href: '/organizations', label: 'Organizations' },
		{ href: '/facilities', label: 'Facilities' },
		{ href: '/formats', label: 'Formats' },
		{ href: '/ratings', label: 'Ratings' },
		{ href: '/rulesets', label: 'Rulesets' },
		{ href: '/scoring-structures', label: 'Scoring Structures' },
		{ href: '/playoff-structures', label: 'Playoff Structures' }
	];

	function logout() {
		clearToken();
		loggedIn = false;
		goto('/');
	}
</script>

<header class="sticky top-0 z-40 w-full border-b bg-background">
	<div class="mx-auto flex h-14 max-w-6xl items-center gap-6 px-6">
		<a href="/" class="flex items-center gap-2 text-base font-semibold tracking-tight">
			<img src={logo} alt="" class="size-6" />
			IntraClub
		</a>
		<nav class="flex items-center gap-1 text-sm">
			<Popover>
				<PopoverTrigger class={buttonVariants({ variant: 'ghost', size: 'sm' })}>Settings</PopoverTrigger>
				<PopoverContent class="flex flex-col gap-0.5 p-1.5" align="start">
					{#each settingsLinks as link}
						<a
							href={link.href}
							class="rounded-md px-2.5 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
							>{link.label}</a
						>
					{/each}
				</PopoverContent>
			</Popover>
			<a href="/drafts" class="rounded-md px-2 py-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">Drafts</a>
			<a href="/teams" class="rounded-md px-2 py-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">Teams</a>
			<a href="/photos" class="rounded-md px-2 py-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">Photos</a>
			<a href="/users" class="rounded-md px-2 py-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground">Users</a>
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
