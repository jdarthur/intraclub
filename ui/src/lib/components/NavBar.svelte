<script lang="ts">
	import { isLoggedIn, clearToken, sessionExpired } from '$lib/auth';
	import { identity } from '$lib/identity.svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Popover,
		PopoverContent,
		PopoverTrigger
	} from '$lib/components/ui/popover/index.js';
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import logo from '$lib/assets/favicon.svg';

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
		'flex h-9 items-center rounded-md px-3 text-base font-medium text-primary-foreground transition-colors hover:bg-foreground/10';
	const navItemActiveClass = 'bg-foreground/15';

	function isActive(href: string) {
		const p = page.url.pathname;
		return p === href || p.startsWith(href + '/');
	}

	const settingsActive = $derived(settingsLinks.some((link) => isActive(link.href)));

	// Reads of `identity` state and the `sessionExpired` store are reactive, so
	// this flips to true once the layout resolves the signed-in user (also after
	// a client-side login redirect) and back to false on logout/expiry, when
	// identity.reset() drops the cached user.
	const showAccount = $derived(
		identity.userId !== null ||
			identity.user !== null ||
			(isLoggedIn() && !$sessionExpired)
	);

	function logout() {
		clearToken();
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
			{#if showAccount}
				<Popover>
					<PopoverTrigger
						class="flex h-9 items-center gap-2 rounded-md px-3 text-base font-medium text-primary-foreground transition-colors hover:bg-foreground/15 aria-expanded:bg-foreground/15"
						aria-label="Account"
					>
						{#if identity.user}
							<span
								class="flex size-6 shrink-0 items-center justify-center rounded-full bg-foreground/15 text-xs font-semibold"
								>{identity.initials}</span
							>
							<span class="max-w-40 truncate">{identity.displayName}</span>
						{:else}
							<span class="text-sm text-primary-foreground/80">Account</span>
						{/if}
						<ChevronDownIcon class="size-4 shrink-0 opacity-70" aria-hidden />
					</PopoverTrigger>
					<PopoverContent class="flex min-w-40 flex-col gap-0.5 p-1.5" align="end">
						<div class="px-2.5 py-1.5 text-sm text-muted-foreground">
							{identity.displayName || 'Signed in'}
						</div>
						<Button
							variant="ghost"
							class="justify-start rounded-md px-2.5 py-1.5 text-sm"
							onclick={logout}
							>Log out</Button
						>
					</PopoverContent>
				</Popover>
			{:else}
				<a
					href="/login"
					class="flex h-9 items-center rounded-md px-3 text-base font-medium text-primary-foreground transition-colors hover:bg-foreground/10"
					>Log in</a
				>
			{/if}
		</div>
	</div>
</header>
