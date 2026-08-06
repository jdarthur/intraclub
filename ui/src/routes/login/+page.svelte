<script lang="ts">
	import { setToken } from '$lib/auth';

	let email = $state('');
	let message = $state('');
	let error = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		message = '';
		error = '';

		const res = await fetch('/api/one_time_password', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ email })
		});

		if (!res.ok) {
			const body = await res.json();
			error = body.error ?? 'Failed to send login link';
			return;
		}

		message = 'Check your email for the login link.';
	}
</script>

<svelte:head>
	<title>Log in</title>
</svelte:head>

<h1>Log in</h1>

<form onsubmit={handleSubmit}>
	<label>
		Email
		<input type="email" bind:value={email} required />
	</label>
	<button type="submit">Send login link</button>
</form>

{#if message}
	<p>{message}</p>
{/if}

{#if error}
	<p>{error}</p>
{/if}
