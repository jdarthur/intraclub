<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getTeam, promoteCoCaptain } from '$lib/team';
	import type { TeamAssignment, TeamRoster, TeamRole } from '$lib/team';
	import { getCurrentUserId } from '$lib/auth';
	import { listUsers, fullName } from '$lib/user';
	import type { User } from '$lib/user';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Card,
		CardContent,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card/index.js';
	import {
		Tabs,
		TabsContent,
		TabsList,
		TabsTrigger
	} from '$lib/components/ui/tabs/index.js';

	const id = () => page.params.id as string;

	let roster = $state<TeamRoster | null>(null);
	let users = $state<User[]>([]);
	let activeTab = $state('members');
	let loading = $state(true);
	let loadError = $state('');
	let actionError = $state('');
	let actionMessage = $state('');

	onMount(load);

	async function load() {
		loading = true;
		loadError = '';
		try {
			const [teamRoster, userList] = await Promise.all([getTeam(id()), listUsers()]);
			roster = teamRoster;
			users = userList;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load team';
		} finally {
			loading = false;
		}
	}

	function userName(userId: string): string {
		const u = users.find((x) => x.id === userId);
		return u ? fullName(u) : userId;
	}

	function roleLabel(role: TeamRole): string {
		switch (role) {
			case 'captain':
				return 'Captain';
			case 'co_captain':
				return 'Co-captain';
			default:
				return 'Member';
		}
	}

	// The current user can assign co-captains if they are this team's captain or
	// a co-captain themselves (mirrors the backend's EditableBy).
	const canManage = $derived.by(() => {
		const me = getCurrentUserId();
		if (!me || !roster) return false;
		return roster.assignments.some((a) => a.user_id === me && a.role !== 'member');
	});

	const sortedAssignments = $derived(
		roster ? [...roster.assignments].sort((a, b) => userName(a.user_id).localeCompare(userName(b.user_id))) : []
	);

	// Members eligible to be promoted to co-captain (regular members only).
	const promotableMembers = $derived(
		sortedAssignments.filter((a) => a.role === 'member')
	);

	async function promote(userId: string) {
		actionError = '';
		actionMessage = '';
		try {
			await promoteCoCaptain(id(), userId);
			actionMessage = `${userName(userId)} is now a co-captain.`;
			await load();
		} catch (e) {
			actionError = e instanceof Error ? e.message : 'Failed to promote member';
		}
	}
</script>

<svelte:head>
	<title>Intraclub | {roster ? roster.team.name : 'Team'}</title>
</svelte:head>

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">{roster?.team.name ?? 'Team'}</h1>
	<a href="/teams" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to teams</a>
</div>

{#if loading}
	<p class="mt-4 text-muted-foreground">Loading...</p>
{:else if loadError}
	<p class="mt-4 text-sm font-medium text-destructive">{loadError}</p>
{:else if roster}
	<Card class="mt-6">
		<CardHeader>
			<CardTitle class="text-base">Team</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="flex flex-wrap items-center gap-4 text-sm">
				<Badge variant="secondary">{roster.assignments.length} members</Badge>
				<span class="text-muted-foreground">{roster.team.color.name}</span>
			</div>
		</CardContent>
	</Card>

	<Tabs bind:value={activeTab} class="mt-6 w-full">
		<TabsList>
			<TabsTrigger value="members">Members</TabsTrigger>
			<TabsTrigger value="co-captains">Co-captains</TabsTrigger>
		</TabsList>
		<TabsContent value="members">
			{#if sortedAssignments.length === 0}
				<p class="mt-4 text-sm text-muted-foreground">No players assigned.</p>
			{:else}
				<ul class="mt-4 space-y-2 text-sm">
					{#each sortedAssignments as assignment}
						<li class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2">
							<span class="font-medium">{userName(assignment.user_id)}</span>
							<Badge variant={assignment.role === 'captain' ? 'default' : 'secondary'}>
								{roleLabel(assignment.role)}
							</Badge>
						</li>
					{/each}
				</ul>
			{/if}
		</TabsContent>
		<TabsContent value="co-captains">
			<p class="mt-4 text-sm text-muted-foreground">
				Co-captains help manage the team roster. Rosters are otherwise fixed after the draft is finalized.
			</p>

			{#if actionError}
				<p class="mt-3 text-sm font-medium text-destructive">{actionError}</p>
			{/if}
			{#if actionMessage}
				<p class="mt-3 text-sm font-medium text-emerald-600">{actionMessage}</p>
			{/if}

			{#if canManage}
				{#if promotableMembers.length === 0}
					<p class="mt-4 text-sm text-muted-foreground">There are no members to promote.</p>
				{:else}
					<ul class="mt-4 space-y-2 text-sm">
						{#each promotableMembers as assignment}
							<li class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2">
								<span class="font-medium">{userName(assignment.user_id)}</span>
								<Button size="sm" onclick={() => promote(assignment.user_id)}>
									Promote to co-captain
								</Button>
							</li>
						{/each}
					</ul>
				{/if}
			{:else}
				<p class="mt-4 text-sm text-muted-foreground">
					Only this team's captain or co-captains can assign roles.
				</p>
			{/if}
		</TabsContent>
	</Tabs>
{/if}
