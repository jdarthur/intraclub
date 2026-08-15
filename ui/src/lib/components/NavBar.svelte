<script lang="ts">
	import { isLoggedIn, clearToken } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button/index.js';
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

	// Shared look for every nav item (Settings trigger + top-level links): same
	// height, size and weight so nothing reads smaller or lighter than the rest.
	const navItemClass =
		'flex h-9 items-center rounded-md px-3 text-base font-semibold text-primary-foreground transition-colors hover:bg-foreground/10';
	const navItemActiveClass = 'bg-foreground/15';

	function isActive(href: string) {
		const p = page.url.pathname;
		return p === href || p.startsWith(href + '/');
	}

	const settingsActive = $derived(settingsLinks.some((link) => isActive(link.href)));

	function logout() {
		clearToken();
		loggedIn = false;
		goto('/');
	}
</script>

<header
	class="sticky top-0 z-40 w-full border-b border-primary-foreground/20 bg-primary text-primary-foreground"
>
	<div class="mx-auto flex h-14 max-w-6xl items-center gap-6 px-6">
		<a href="/" class="flex items-center gap-2 text-base font-semibold tracking-tight">
			<img src={logo} alt="" class="size-6" />
			IntraClub
		</a>
		<nav class="flex items-center gap-1">
			<Popover>
				<PopoverTrigger
					class={`${navItemClass} ${settingsActive ? navItemActiveClass : ''} aria-expanded:bg-foreground/15 aria-expanded:text-primary-foreground`}
					aria-current={settingsActive ? 'true' : undefined}
					>Settings</PopoverTrigger
				>
				<PopoverContent class="flex flex-col gap-0.5 p-1.5" align="start">
					{#each settingsLinks as link}
						<a
							href={link.href}
							class="rounded-md px-2.5 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground {isActive(link.href) ? 'bg-muted font-semibold text-foreground' : ''}"
							aria-current={isActive(link.href) ? 'page' : undefined}
							>{link.label}</a
						>
					{/each}
				</PopoverContent>
			</Popover>
			<a
				href="/drafts"
				class={`${navItemClass} ${isActive('/drafts') ? navItemActiveClass : ''}`}
				aria-current={isActive('/drafts') ? 'page' : undefined}
				>Drafts</a
			>
			<a
				href="/teams"
				class={`${navItemClass} ${isActive('/teams') ? navItemActiveClass : ''}`}
				aria-current={isActive('/teams') ? 'page' : undefined}
				>Teams</a
			>
			<a
				href="/photos"
				class={`${navItemClass} ${isActive('/photos') ? navItemActiveClass : ''}`}
				aria-current={isActive('/photos') ? 'page' : undefined}
				>Photos</a
			>
			<a
				href="/users"
				class={`${navItemClass} ${isActive('/users') ? navItemActiveClass : ''}`}
				aria-current={isActive('/users') ? 'page' : undefined}
				>Users</a
			>
		</nav>
		<div class="ml-auto flex items-center">
			{#if loggedIn}
				<Button
					variant="ghost"
					class="h-9 rounded-md px-3 text-base font-semibold hover:bg-foreground/15 hover:text-primary-foreground"
					onclick={logout}
					>Log out</Button
				>
			{:else}
				<a
					href="/login"
					class="flex h-9 items-center rounded-md px-3 text-base font-semibold text-primary-foreground transition-colors hover:bg-foreground/10"
					>Log in</a
				>
			{/if}
		</div>
	</div>
</header>
