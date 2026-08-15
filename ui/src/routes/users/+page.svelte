<script lang="ts">
	import { onMount } from 'svelte';
	import { Async } from '$lib/async.svelte';
	import {
		AsyncSection,
		DataTable,
		EmptyState,
		PageHeader
	} from '$lib/components/app/index.js';
	import type { Column } from '$lib/components/app/data-table.svelte';
	import { listUsers, fullName } from '$lib/user';
	import type { User } from '$lib/user';
	import { Button } from '$lib/components/ui/button/index.js';
	import UsersIcon from '@lucide/svelte/icons/users';

	const users = new Async<User[]>();
	onMount(() => users.run(() => listUsers()));

	const columns: Column<User>[] = [
		{ key: 'name', header: 'Name', sortable: true, cell: nameCell },
		{ key: 'email', header: 'Email', hideBelow: 'sm', value: (u) => u.email },
		{
			key: 'phone',
			header: 'Phone',
			hideBelow: 'md',
			value: (u) => u.phone_number || '—'
		},
		{ key: 'verified', header: 'Verified', value: (u) => (u.verified ? 'Yes' : 'No') }
	];
</script>

{#snippet nameCell(u: User)}
	<a href={`/users/${u.id}`} class="font-medium text-primary underline-offset-4 hover:underline">
		{fullName(u)}
	</a>
{/snippet}

<svelte:head>
	<title>Intraclub | Users</title>
</svelte:head>

<PageHeader title="Users" description="Everyone registered to play" icon={UsersIcon}>
	{#snippet actions()}
		<Button href="/users/import">Import users</Button>
	{/snippet}
</PageHeader>

<AsyncSection state={users} isEmpty={(u) => u.length === 0}>
	{#snippet loading()}
		<DataTable rows={[]} {columns} getKey={(u) => u.id} loading />
	{/snippet}
	{#snippet empty()}
		<EmptyState title="No users yet." description="Import or invite players to get them registered." />
	{/snippet}
	{#snippet children(us)}
		<DataTable
			rows={us}
			{columns}
			getKey={(u) => u.id}
			caption="Users"
			filter
			filterLabel="Filter users"
		/>
	{/snippet}
</AsyncSection>
