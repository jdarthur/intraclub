<script lang="ts">
	import './styles.css';
	import favicon from '$lib/assets/favicon.svg';
	import { page } from '$app/state';
	import { startSessionMonitor, getCurrentUserId } from '$lib/auth';
	import { identity } from '$lib/identity.svelte';
	import NavBar from '$lib/components/NavBar.svelte';
	import { Toaster } from '$lib/components/ui/sonner/index.js';

	let { children } = $props();

	// Start the background JWT-expiry ticker for the whole app and resolve the
	// signed-in user's identity. The layout persists across client-side
	// navigations, so reading `page.url` re-runs this on every navigation —
	// including the post-login redirect from /auth/callback, which is when the
	// token (and thus getCurrentUserId) first becomes available.
	$effect(() => {
		void page.url;
		const stop = startSessionMonitor();
		if (getCurrentUserId()) void identity.load();
		return () => stop();
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<a
	href="#main-content"
	class="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-md focus:bg-background focus:px-4 focus:py-2 focus:text-sm focus:font-semibold focus:text-foreground focus:shadow-lg"
	>Skip to main content</a
>
<NavBar />
<main id="main-content" class="mx-auto w-full max-w-6xl px-4 py-6 md:px-6">
	{@render children()}
</main>
<Toaster />
