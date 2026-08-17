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
	import {
		DropdownMenu,
		DropdownMenuContent,
		DropdownMenuRadioGroup,
		DropdownMenuRadioItem,
		DropdownMenuTrigger
	} from '$lib/components/ui/dropdown-menu/index.js';
	import { mode, setMode, userPrefersMode } from 'mode-watcher';
	import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '$lib/components/ui/sheet/index.js';
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import MenuIcon from '@lucide/svelte/icons/menu';
	import SunIcon from '@lucide/svelte/icons/sun';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import MonitorIcon from '@lucide/svelte/icons/monitor';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ListOrderedIcon from '@lucide/svelte/icons/list-ordered';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import BuildingIcon from '@lucide/svelte/icons/building';
	import Building2Icon from '@lucide/svelte/icons/building-2';
	import ListTreeIcon from '@lucide/svelte/icons/list-tree';
	import StarIcon from '@lucide/svelte/icons/star';
	import ScrollTextIcon from '@lucide/svelte/icons/scroll-text';
	import GaugeIcon from '@lucide/svelte/icons/gauge';
	import TrophyIcon from '@lucide/svelte/icons/trophy';
	import UsersIcon from '@lucide/svelte/icons/users';
	import ImageIcon from '@lucide/svelte/icons/image';
	import logo from '$lib/assets/favicon.svg';

	// League-config pages live in the Settings dropdown. Each link carries the
	// same icon its page uses (see the PageHeader on each list page).
	const settingsLinks = [
		{ href: '/organizations', label: 'Organizations', icon: BuildingIcon },
		{ href: '/facilities', label: 'Facilities', icon: Building2Icon },
		{ href: '/formats', label: 'Formats', icon: ListTreeIcon },
		{ href: '/ratings', label: 'Ratings', icon: StarIcon },
		{ href: '/rulesets', label: 'Rulesets', icon: ScrollTextIcon },
		{ href: '/scoring-structures', label: 'Scoring Structures', icon: GaugeIcon },
		{ href: '/playoff-structures', label: 'Playoff Structures', icon: TrophyIcon }
	];

	// Users and Photos are also under Settings, separated from the config pages
	// by a divider so the account/asset group reads as its own tab.
	const accountLinks = [
		{ href: '/users', label: 'Users', icon: UsersIcon },
		{ href: '/photos', label: 'Photos', icon: ImageIcon }
	];

	// Seasons is a dropdown: the season dashboard plus the draft-creation flow.
	const seasonsLinks = [
		{ href: '/seasons', label: 'Seasons', icon: CalendarIcon },
		{ href: '/drafts/new', label: 'New Draft', icon: ListOrderedIcon }
	];

	// Shared look for every nav item (dropdown triggers + top-level links): same
	// height, size and weight so nothing reads smaller or lighter than the rest.
	const navItemClass =
		'flex h-9 items-center gap-2 rounded-md px-3 text-base font-medium text-primary-foreground transition-colors hover:bg-foreground/10';
	const navItemActiveClass = 'bg-foreground/15';

	// Mobile sheet links sit on the popover surface instead of the primary bar.
	const mobileNavItemClass =
		'flex h-9 items-center gap-2 rounded-md px-3 text-base font-medium transition-colors hover:bg-muted';
	const mobileNavItemActiveClass = 'bg-muted text-foreground';

	// Sheet state: bind:open keeps the trigger and content in sync so links can
	// close the sheet (bits-ui unmounts the content while closed).
	let mobileOpen = $state(false);

	function isActive(href: string) {
		const p = page.url.pathname;
		return p === href || p.startsWith(href + '/');
	}

	const settingsActive = $derived(
		[...settingsLinks, ...accountLinks].some((link) => isActive(link.href))
	);
	const seasonsActive = $derived(seasonsLinks.some((link) => isActive(link.href)));

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
		mobileOpen = false;
		clearToken();
		goto('/');
	}

	function closeMobile() {
		mobileOpen = false;
	}
</script>

<header
	class="sticky top-0 z-40 w-full border-b border-primary-foreground/20 bg-primary text-primary-foreground"
>
	<div class="mx-auto flex h-14 max-w-6xl items-center gap-3 px-4 md:gap-6 md:px-6">
		<a href="/" class="flex items-center gap-2 text-base font-semibold tracking-tight">
			<img src={logo} alt="" class="size-6" />
			IntraClub
		</a>
		<!-- Exactly one <nav>: it always carries the mobile sheet trigger and the
		     desktop links. The sheet's links live inside SheetContent (unmounted
		     while closed), so at most one set of nav links is in the a11y tree. -->
		<nav class="flex items-center gap-1">
			<Sheet bind:open={mobileOpen}>
				<SheetTrigger
					class="flex size-9 items-center justify-center rounded-md text-primary-foreground transition-colors hover:bg-foreground/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring md:hidden"
					aria-label="Open menu"
				>
					<MenuIcon class="size-5" aria-hidden />
				</SheetTrigger>
				<SheetContent side="left" class="w-80">
					<SheetHeader class="border-b px-4 pb-3">
						<SheetTitle>Menu</SheetTitle>
					</SheetHeader>
					<div class="flex flex-col gap-0.5 overflow-y-auto p-3">
						<a
							href="/teams"
							onclick={closeMobile}
							class={`${mobileNavItemClass} ${isActive('/teams') ? mobileNavItemActiveClass : ''}`}
							aria-current={isActive('/teams') ? 'page' : undefined}
							><ShieldIcon class="size-4 shrink-0" aria-hidden />Teams</a
						>
						<p
							class="px-3 pb-1 pt-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground"
							>Seasons</p
						>
						{#each seasonsLinks as link}
							{@const Icon = link.icon}
							<a
								href={link.href}
								onclick={closeMobile}
								class={`${mobileNavItemClass} ${isActive(link.href) ? mobileNavItemActiveClass : ''}`}
								aria-current={isActive(link.href) ? 'page' : undefined}
								><Icon class="size-4 shrink-0" aria-hidden />{link.label}</a
							>
						{/each}
						<p
							class="px-3 pb-1 pt-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground"
							>Settings</p
						>
						{#each settingsLinks as link}
							{@const Icon = link.icon}
							<a
								href={link.href}
								onclick={closeMobile}
								class={`${mobileNavItemClass} ${isActive(link.href) ? mobileNavItemActiveClass : ''}`}
								aria-current={isActive(link.href) ? 'page' : undefined}
								><Icon class="size-4 shrink-0" aria-hidden />{link.label}</a
							>
						{/each}
						<div class="my-2 border-t"></div>
						{#each accountLinks as link}
							{@const Icon = link.icon}
							<a
								href={link.href}
								onclick={closeMobile}
								class={`${mobileNavItemClass} ${isActive(link.href) ? mobileNavItemActiveClass : ''}`}
								aria-current={isActive(link.href) ? 'page' : undefined}
								><Icon class="size-4 shrink-0" aria-hidden />{link.label}</a
							>
						{/each}
					</div>
					<div class="mt-auto border-t p-3">
						{#if showAccount}
							<Button variant="ghost" class="w-full justify-start text-base" onclick={logout}>
								Log out
							</Button>
						{:else}
							<a
								href="/login"
								onclick={closeMobile}
								class="flex h-9 items-center rounded-md px-3 text-base font-medium transition-colors hover:bg-muted"
								>Log in</a
							>
						{/if}
					</div>
				</SheetContent>
			</Sheet>
			<div class="hidden items-center gap-1 md:flex">
				<a
					href="/teams"
					class={`${navItemClass} ${isActive('/teams') ? navItemActiveClass : ''}`}
					aria-current={isActive('/teams') ? 'page' : undefined}
					><ShieldIcon class="size-4 shrink-0" aria-hidden />Teams</a
				>
				<Popover>
					<PopoverTrigger
						class={`${navItemClass} ${seasonsActive ? navItemActiveClass : ''} aria-expanded:bg-foreground/15 aria-expanded:text-primary-foreground`}
						aria-current={seasonsActive ? 'true' : undefined}
						><CalendarIcon class="size-4 shrink-0" aria-hidden />Seasons<ChevronDownIcon class="size-4 shrink-0 opacity-70" aria-hidden /></PopoverTrigger
					>
					<PopoverContent class="flex flex-col gap-0.5 p-1.5" align="start">
						{#each seasonsLinks as link}
							{@const Icon = link.icon}
							<a
								href={link.href}
								class="flex items-center gap-2 rounded-md px-2.5 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground {isActive(link.href) ? 'bg-muted font-semibold text-foreground' : ''}"
								aria-current={isActive(link.href) ? 'page' : undefined}
								><Icon class="size-4 shrink-0" aria-hidden />{link.label}</a
							>
						{/each}
					</PopoverContent>
				</Popover>
				<Popover>
					<PopoverTrigger
						class={`${navItemClass} ${settingsActive ? navItemActiveClass : ''} aria-expanded:bg-foreground/15 aria-expanded:text-primary-foreground`}
						aria-current={settingsActive ? 'true' : undefined}
						><SettingsIcon class="size-4 shrink-0" aria-hidden />Settings<ChevronDownIcon class="size-4 shrink-0 opacity-70" aria-hidden /></PopoverTrigger
					>
					<PopoverContent class="flex flex-col gap-0.5 p-1.5" align="start">
						{#each settingsLinks as link}
							{@const Icon = link.icon}
							<a
								href={link.href}
								class="flex items-center gap-2 rounded-md px-2.5 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground {isActive(link.href) ? 'bg-muted font-semibold text-foreground' : ''}"
								aria-current={isActive(link.href) ? 'page' : undefined}
								><Icon class="size-4 shrink-0" aria-hidden />{link.label}</a
							>
						{/each}
						<div class="my-1 border-t"></div>
						{#each accountLinks as link}
							{@const Icon = link.icon}
							<a
								href={link.href}
								class="flex items-center gap-2 rounded-md px-2.5 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground {isActive(link.href) ? 'bg-muted font-semibold text-foreground' : ''}"
								aria-current={isActive(link.href) ? 'page' : undefined}
								><Icon class="size-4 shrink-0" aria-hidden />{link.label}</a
							>
						{/each}
					</PopoverContent>
				</Popover>
			</div>
		</nav>
		<div class="ml-auto flex items-center gap-2">
			<!-- Theme picker: Light / Dark / System. Defaults to the OS preference
			     (mode-watcher's `system`) and persists an explicit choice in
			     localStorage, so the `.dark` class on <html> survives reloads. -->
			<DropdownMenu>
				<DropdownMenuTrigger
					class="flex size-9 items-center justify-center rounded-md text-primary-foreground transition-colors hover:bg-foreground/10 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-expanded:bg-foreground/15"
					aria-label="Toggle theme"
				>
					{#if mode.current === 'dark'}
						<MoonIcon class="size-5" aria-hidden />
					{:else}
						<SunIcon class="size-5" aria-hidden />
					{/if}
				</DropdownMenuTrigger>
				<DropdownMenuContent align="end">
					<DropdownMenuRadioGroup
						value={userPrefersMode.current}
						onValueChange={(v) => setMode(v as 'light' | 'dark' | 'system')}
					>
						<DropdownMenuRadioItem value="light">
							<SunIcon class="size-4" aria-hidden />
							Light
						</DropdownMenuRadioItem>
						<DropdownMenuRadioItem value="dark">
							<MoonIcon class="size-4" aria-hidden />
							Dark
						</DropdownMenuRadioItem>
						<DropdownMenuRadioItem value="system">
							<MonitorIcon class="size-4" aria-hidden />
							System
						</DropdownMenuRadioItem>
					</DropdownMenuRadioGroup>
				</DropdownMenuContent>
			</DropdownMenu>
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
							<span class="max-sm:hidden max-w-40 truncate">{identity.displayName}</span>
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
