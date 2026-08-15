<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getUser } from '$lib/user';
	import type { User } from '$lib/user';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';

	const id = () => page.params.id as string;

	let user = $state<User | null>(null);
	let loadError = $state('');

	onMount(load);

	async function load() {
		loadError = '';
		try {
			user = await getUser(id());
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load user';
		}
	}
</script>

<svelte:head>
	<title>Intraclub | {user ? `${user.first_name} ${user.last_name}` : 'User'}</title>
</svelte:head>

{#if loadError}
	<h1 class="text-2xl font-semibold tracking-tight">User</h1>
	<p class="text-sm font-medium text-destructive">{loadError}</p>
	<a href="/users" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to users</a>
{:else if !user}
	<h1 class="text-2xl font-semibold tracking-tight">User</h1>
	<p class="text-muted-foreground">Loading...</p>
{:else}
	<div class="flex items-center gap-4">
		<h1 class="text-2xl font-semibold tracking-tight">{user.first_name} {user.last_name}</h1>
		<a href="/users" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to users</a>
	</div>

	<Card class="mt-6 max-w-md">
		<CardHeader>
			<CardTitle>User details</CardTitle>
		</CardHeader>
		<CardContent>
			<dl class="grid gap-4 text-sm">
				<div class="flex justify-between gap-4">
					<dt class="text-muted-foreground">Name</dt>
					<dd>{user.first_name} {user.last_name}</dd>
				</div>
				<div class="flex justify-between gap-4">
					<dt class="text-muted-foreground">Email</dt>
					<dd>{user.email}</dd>
				</div>
				<div class="flex justify-between gap-4">
					<dt class="text-muted-foreground">Phone</dt>
					<dd>{user.phone_number || '—'}</dd>
				</div>
				<div class="flex justify-between gap-4">
					<dt class="text-muted-foreground">Verified</dt>
					<dd>{user.verified ? 'Yes' : 'No'}</dd>
				</div>
			</dl>
		</CardContent>
	</Card>
{/if}
