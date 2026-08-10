<script lang="ts">
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

{#if devLink}
	<p>
		<a href={devLink}>Log in</a>
	</p>
{/if}

{#if error}
	<p class="error">{error}</p>
{/if}

<style>
	.error {
		color: #c00;
	}
	form {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		max-width: 24rem;
		margin-top: 1rem;
	}
	label {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	input {
		padding: 0.35rem;
	}
	button {
		padding: 0.35rem 0.6rem;
		width: fit-content;
	}
</style>
