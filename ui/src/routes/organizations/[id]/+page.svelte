<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { getCurrentUserId } from '$lib/auth';
	import {
		getOrganization,
		updateOrganization,
		deleteOrganization,
		listMembers,
		addMember,
		removeMember
	} from '$lib/organization';
	import type { Organization } from '$lib/organization';
	import { listUsers, fullName } from '$lib/user';
	import type { User } from '$lib/user';
	import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		NativeSelect,
		NativeSelectOption
	} from '$lib/components/ui/native-select/index.js';
	import {
		Popover,
		PopoverClose,
		PopoverContent,
		PopoverHeader,
		PopoverTitle,
		PopoverTrigger
	} from '$lib/components/ui/popover/index.js';

	const id = () => page.params.id as string;

	let organization = $state<Organization | null>(null);
	let loadError = $state('');
	let name = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);

	let members = $state<User[]>([]);
	let membersError = $state('');
	let allUsers = $state<User[]>([]);
	let selectedUserId = $state('');
	let addingMember = $state(false);

	let isOwner = $state(false);

	onMount(load);

	async function load() {
		loadError = '';
		try {
			const org = await getOrganization(id());
			organization = org;
			name = org.name;
			isOwner = getCurrentUserId() === org.user_id;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load organization';
			return;
		}

		try {
			members = await listMembers(id());
		} catch (e) {
			membersError = e instanceof Error ? e.message : 'Failed to load members';
		}

		// The owner can add any registered user to the membership roster.
		if (isOwner) {
			try {
				allUsers = await listUsers();
			} catch {
				// member-edit controls stay usable; the picker just lists members
			}
		}
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const updated = await updateOrganization(id(), { name: name.trim() });
			organization = updated;
			name = updated.name;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update organization';
		} finally {
			saving = false;
		}
	}

	async function handleAddMember() {
		if (!selectedUserId) return;
		error = '';
		membersError = '';
		addingMember = true;
		try {
			await addMember(id(), selectedUserId);
			selectedUserId = '';
			members = await listMembers(id());
		} catch (err) {
			membersError = err instanceof Error ? err.message : 'Failed to add member';
		} finally {
			addingMember = false;
		}
	}

	async function handleRemoveMember(userId: string) {
		error = '';
		membersError = '';
		try {
			await removeMember(id(), userId);
			members = await listMembers(id());
		} catch (err) {
			membersError = err instanceof Error ? err.message : 'Failed to remove member';
		}
	}

	async function handleDelete() {
		deleteOpen = false;
		error = '';
		deleting = true;
		try {
			await deleteOrganization(id());
			await goto('/organizations');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete organization';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | {organization ? organization.name : 'Organization'}</title>
</svelte:head>

{#if loadError}
	<h1 class="text-2xl font-semibold tracking-tight">Organization</h1>
	<p class="text-sm font-medium text-destructive">{loadError}</p>
	<a href="/organizations" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to organizations</a>
{:else if !organization}
	<h1 class="text-2xl font-semibold tracking-tight">Organization</h1>
	<p class="text-muted-foreground">Loading...</p>
{:else}
	<div class="flex items-center gap-4">
		<h1 class="text-2xl font-semibold tracking-tight">{organization.name}</h1>
		<a href="/organizations" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to organizations</a>
	</div>

	{#if isOwner}
		<Card class="mt-6 max-w-md">
			<CardHeader>
				<CardTitle class="text-base">Organization details</CardTitle>
			</CardHeader>
			<CardContent>
				<form onsubmit={handleSave} class="flex flex-col gap-4">
					<div class="flex flex-col gap-2">
						<Label for="name">Name</Label>
						<Input id="name" type="text" bind:value={name} required />
					</div>
					<Button type="submit" disabled={saving} class="w-fit">
						{saving ? 'Saving...' : 'Save changes'}
					</Button>
				</form>
			</CardContent>
		</Card>
	{/if}

	<Card class="mt-6 max-w-md">
		<CardHeader>
			<CardTitle class="text-base">Members</CardTitle>
		</CardHeader>
		<CardContent class="flex flex-col gap-4">
			{#if membersError}
				<p class="text-sm font-medium text-destructive">{membersError}</p>
			{/if}

			{#if members.length === 0}
				<p class="text-sm text-muted-foreground">No members yet.</p>
			{:else}
				<ul class="flex flex-col gap-2">
					{#each members as member}
						<li class="flex items-center justify-between gap-4 text-sm">
							<span>{fullName(member)} ({member.email})</span>
							{#if isOwner}
								<Button variant="outline" size="sm" onclick={() => handleRemoveMember(member.id)}>
									Remove
								</Button>
							{/if}
						</li>
					{/each}
				</ul>
			{/if}

			{#if isOwner}
				<div class="flex flex-col gap-2">
					<Label for="member">Add a member</Label>
					<div class="flex items-center gap-2">
						<NativeSelect id="member" bind:value={selectedUserId} class="flex-1" aria-label="Member to add">
							<NativeSelectOption value="">Select a user...</NativeSelectOption>
							{#each allUsers as user}
								<NativeSelectOption value={user.id}>{fullName(user)} ({user.email})</NativeSelectOption>
							{/each}
						</NativeSelect>
						<Button onclick={handleAddMember} disabled={addingMember || !selectedUserId}>
							{addingMember ? 'Adding...' : 'Add'}
						</Button>
					</div>
				</div>
			{/if}
		</CardContent>
	</Card>

	{#if isOwner}
		<div class="mt-4">
			<Popover bind:open={deleteOpen}>
				<PopoverTrigger disabled={deleting} class={buttonVariants({ variant: 'destructive' })}>
					{deleting ? 'Deleting...' : 'Delete organization'}
				</PopoverTrigger>
				<PopoverContent class="w-80">
					<PopoverHeader>
						<PopoverTitle>Delete organization?</PopoverTitle>
						<p class="text-sm text-muted-foreground">
							This permanently removes this organization and cannot be undone.
						</p>
					</PopoverHeader>
					<div class="flex justify-end gap-2">
						<PopoverClose class={buttonVariants({ variant: 'outline', size: 'sm' })}>Cancel</PopoverClose>
						<Button variant="destructive" size="sm" onclick={handleDelete}>Delete</Button>
					</div>
				</PopoverContent>
			</Popover>
		</div>
	{/if}

	{#if error}
		<p class="mt-4 text-sm font-medium text-destructive">{error}</p>
	{/if}
{/if}
