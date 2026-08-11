<script lang="ts">
	import './styles.css';
	import favicon from '$lib/assets/favicon.svg';
	import { startSessionMonitor } from '$lib/auth';
	import NavBar from '$lib/components/NavBar.svelte';

	let { children } = $props();

	// Start the background JWT-expiry ticker for the whole app. The layout
	// persists across client-side navigations, so this runs once per page load
	// and keeps watching until the page is torn down.
	$effect(() => {
		const stop = startSessionMonitor();
		return () => stop();
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<NavBar />
<main class="mx-auto w-full max-w-6xl px-6 py-6">
	{@render children()}
</main>
