<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { Async } from '$lib/async.svelte';
	import { AsyncSection, PageHeader } from '$lib/components/app/index.js';
	import { getUser, fullName } from '$lib/user';
	import type { User } from '$lib/user';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import UsersIcon from '@lucide/svelte/icons/users';

	const id = () => page.params.id as string;

	const user = new Async<User>();
	onMount(() => user.run(() => getUser(id())));
</script>

<svelte:head>
	<title>Intraclub | {user.data ? fullName(user.data) : 'User'}</title>
</svelte:head>

<PageHeader
	title={user.data ? fullName(user.data) : undefined}
	icon={UsersIcon}
	backHref="/users"
	backLabel="Back to users"
/>

<AsyncSection state={user}>
	{#snippet children(u)}
		<Card class="mt-6 max-w-md">
			<CardHeader>
				<CardTitle>User details</CardTitle>
			</CardHeader>
			<CardContent>
				<dl class="grid gap-4 text-sm">
					<div class="flex justify-between gap-4">
						<dt class="text-muted-foreground">Name</dt>
						<dd>{fullName(u)}</dd>
					</div>
					<div class="flex justify-between gap-4">
						<dt class="text-muted-foreground">Email</dt>
						<dd>{u.email}</dd>
					</div>
					<div class="flex justify-between gap-4">
						<dt class="text-muted-foreground">Phone</dt>
						<dd>{u.phone_number || '—'}</dd>
					</div>
					<div class="flex justify-between gap-4">
						<dt class="text-muted-foreground">Verified</dt>
						<dd>{u.verified ? 'Yes' : 'No'}</dd>
					</div>
				</dl>
			</CardContent>
		</Card>
	{/snippet}
</AsyncSection>
