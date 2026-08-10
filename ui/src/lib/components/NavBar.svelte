<script lang="ts">
	import { isLoggedIn, clearToken } from '$lib/auth';
	import { goto } from '$app/navigation';

	let loggedIn = $state(isLoggedIn());

	function logout() {
		clearToken();
		loggedIn = false;
		goto('/');
	}
</script>

<nav class="navbar">
	<a href="/" class="brand">IntraClub</a>
	<div class="links">
		<a href="/facilities">Facilities</a>
		<a href="/formats">Formats</a>
		<a href="/ratings">Ratings</a>
	</div>
	<div class="actions">
		{#if loggedIn}
			<button type="button" onclick={logout}>Log out</button>
		{:else}
			<a href="/login" class="login">Log in</a>
		{/if}
	</div>
</nav>

<style>
	.navbar {
		display: flex;
		align-items: center;
		gap: 1.5rem;
		padding: 0.75rem 1.25rem;
		background: #1f2937;
		color: #f9fafb;
	}
	.brand {
		color: #fff;
		font-weight: 700;
		text-decoration: none;
	}
	.links {
		display: flex;
		gap: 1rem;
	}
	.links a {
		color: #d1d5db;
		text-decoration: none;
	}
	.links a:hover {
		color: #fff;
	}
	.actions {
		margin-left: auto;
	}
	.login {
		color: #d1d5db;
		text-decoration: none;
	}
	.login:hover {
		color: #fff;
	}
	.actions button {
		background: transparent;
		border: 1px solid #d1d5db;
		color: #d1d5db;
		padding: 0.25rem 0.6rem;
		border-radius: 4px;
		cursor: pointer;
	}
	.actions button:hover {
		color: #fff;
		border-color: #fff;
	}
</style>
