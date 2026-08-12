<script lang="ts">
	import { onMount } from 'svelte';
	import { listUsers } from '$lib/user';
	import type { User } from '$lib/user';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let users = $state<User[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			users = await listUsers();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load users';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Intraclub | Users</title>
</svelte:head>

<div class="flex items-center justify-between gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Users</h1>
	<Button href="/users/import">Import users</Button>
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{:else if users.length === 0}
	<p class="text-muted-foreground">No users yet.</p>
{:else}
	<div class="mt-4 overflow-hidden rounded-lg border">
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead>Name</TableHead>
					<TableHead>Email</TableHead>
					<TableHead>Phone</TableHead>
					<TableHead>Verified</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each users as user}
					<TableRow>
						<TableCell>
							<a href={`/users/${user.id}`} class="font-medium text-primary underline-offset-4 hover:underline">
								{user.first_name} {user.last_name}
							</a>
						</TableCell>
						<TableCell>{user.email}</TableCell>
						<TableCell>{user.phone_number || '—'}</TableCell>
						<TableCell>{user.verified ? 'Yes' : 'No'}</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
