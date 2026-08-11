<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';

	let email = $state('');
	let message = $state('');
	let error = $state('');
	let devLink = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		message = '';
		error = '';
		devLink = '';

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

		const body = await res.json();
		if (body?.token) {
			// DEV MODE ONLY: the API returned the magic-link token in the response
			// instead of emailing it. Render it as a clickable link.
			devLink = `/auth/callback?token=${encodeURIComponent(body.token)}`;
			message = 'DEV MODE ONLY — no email was sent. Use the magic link below to log in.';
		} else {
			message = 'Check your email for the login link.';
		}
	}
</script>

<form onsubmit={handleSubmit} class="flex max-w-sm flex-col gap-4">
	<div class="flex flex-col gap-2">
		<Label for="email">Email</Label>
		<Input id="email" type="email" bind:value={email} required />
	</div>
	<Button type="submit" class="w-fit">Send login link</Button>
</form>

{#if message}
	<p class="text-sm text-muted-foreground">{message}</p>
{/if}

{#if devLink}
	<p>
		<a href={devLink} class="text-sm font-medium text-primary underline underline-offset-4 hover:underline">Log in</a>
	</p>
{/if}

{#if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{/if}
