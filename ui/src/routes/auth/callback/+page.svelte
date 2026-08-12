<script lang="ts">
	import { setToken } from '$lib/auth';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	let status = $state('exchanging');
	let error = $state('');

	async function exchangeToken(token: string) {
		try {
			const res = await fetch('/api/token', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ token })
			});

			if (!res.ok) {
				const body = await res.json();
				error = body.error ?? 'Token exchange failed';
				status = 'error';
				return;
			}

			const data = await res.json();
			setToken(data.jwt);

			if (data.return_to) {
				await goto(data.return_to);
			} else {
				await goto('/');
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'An unexpected error occurred';
			status = 'error';
		}
	}

	$effect(() => {
		if (status !== 'exchanging') return;
		const token = $page.url.searchParams.get('token');
		if (!token) {
			status = 'error';
			error = 'No login token provided.';
			return;
		}
		exchangeToken(token);
	});
</script>

<svelte:head>
	<title>Intraclub | Logging in...</title>
</svelte:head>

{#if status === 'exchanging'}
	<p>Logging you in...</p>
{:else if status === 'error'}
	<h1>Login failed</h1>
	<p>{error}</p>
{/if}
