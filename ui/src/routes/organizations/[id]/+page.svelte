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
		Popover,
		PopoverClose,
		PopoverContent,
		PopoverHeader,
		PopoverTitle,
		PopoverTrigger
	} from '$lib/components/ui/popover/index.js';
	import {
		Tabs,
		TabsContent,
		TabsList,
		TabsTrigger
	} from '$lib/components/ui/tabs/index.js';
	import * as TransferList from '$lib/components/ui/transfer-list/index.js';

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
	let isOwner = $state(false);

	// Membership configuration (owner only). The transfer list core holds the
	// in-memory source (non-members) / target (members) lists and selection;
	// we persist diffs on save.
	let transferCore = $state<TransferList.Core<User> | null>(null);
	let savingMembers = $state(false);

	// Which section is shown in the sidebar at a time.
	let activeTab = $state('details');

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
				// member-edit controls stay usable; the source just lists members
			}
			const memberIds = new Set(members.map((m) => m.id));
			transferCore = new TransferList.Core<User>({
				initialSource: allUsers.filter((u) => !memberIds.has(u.id)),
				initialTarget: members,
				filterPredicate: (u, search) =>
					fullName(u).toLowerCase().includes(search.toLowerCase())
			});
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

	// Persists the transfer-list state: adds users that moved into the target,
	// removes users that moved back to the source.
	async function handleSaveMembers() {
		error = '';
		membersError = '';
		if (!transferCore) return;
		const memberIds = new Set(members.map((m) => m.id));
		const targetIds = new Set(transferCore.target.map((u) => u.id));
		const toAdd = transferCore.target
			.filter((u) => !memberIds.has(u.id))
			.map((u) => u.id);
		const toRemove = members.filter((m) => !targetIds.has(m.id));
		if (toAdd.length === 0 && toRemove.length === 0) return;
		savingMembers = true;
		try {
			for (const userId of toAdd) {
				await addMember(id(), userId);
			}
			for (const member of toRemove) {
				await removeMember(id(), member.id);
			}
			await load();
		} catch (err) {
			membersError = err instanceof Error ? err.message : 'Failed to save members';
		} finally {
			savingMembers = false;
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
		<div class="ml-auto">
			<a href="/organizations" class="text-sm text-muted-foreground hover:text-foreground">&larr; Back to organizations</a>
		</div>
	</div>

	{#if error}
		<p class="mt-4 text-sm font-medium text-destructive">{error}</p>
	{/if}

	<Tabs bind:value={activeTab} orientation="vertical" class="mt-6 flex gap-6">
		<TabsList class="h-fit w-56 shrink-0 items-stretch">
			<TabsTrigger
				value="details"
				class="group justify-start gap-2.5 px-3 py-2 text-base data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:font-semibold data-[state=active]:shadow-sm dark:data-[state=active]:bg-input/30"
			>
				<span
					class="size-1.5 shrink-0 rounded-full bg-primary opacity-0 transition-opacity group-data-[state=active]:opacity-100"
					aria-hidden="true"
				></span>
				Details
			</TabsTrigger>
			<TabsTrigger
				value="members"
				class="group justify-start gap-2.5 px-3 py-2 text-base data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:font-semibold data-[state=active]:shadow-sm dark:data-[state=active]:bg-input/30"
			>
				<span
					class="size-1.5 shrink-0 rounded-full bg-primary opacity-0 transition-opacity group-data-[state=active]:opacity-100"
					aria-hidden="true"
				></span>
				Members
			</TabsTrigger>
		</TabsList>

		<TabsContent value="details" class="flex-1">
			{#if isOwner}
				<Card class="max-w-md">
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
			{:else}
				<Card class="max-w-md">
					<CardHeader>
						<CardTitle class="text-base">Organization</CardTitle>
					</CardHeader>
					<CardContent>
						<p class="text-sm text-muted-foreground">
							Only the owner can edit this organization's details.
						</p>
					</CardContent>
				</Card>
			{/if}
		</TabsContent>

		<TabsContent value="members" class="flex-1">
			{#if membersError}
				<p class="text-sm font-medium text-destructive">{membersError}</p>
			{/if}

			{#if isOwner}
				{#if transferCore}
					<Card class="max-w-2xl">
						<CardHeader>
							<CardTitle class="text-base">Members</CardTitle>
						</CardHeader>
						<CardContent>
							<p class="text-sm text-muted-foreground">
								Move users into the roster to make them members of this organization.
								Changes apply when you save.
							</p>
							<div class="mt-4">
								<TransferList.Root direction="horizontal">
									<TransferList.Container>
										<TransferList.Title title="All users" />
										<TransferList.Toolbar
											variant="source"
											core={transferCore}
											inputPlaceholder="Search users..."
										/>
										<TransferList.Body>
											{#each transferCore.filteredSource as row (row.id)}
												<TransferList.Item side="source" {row} core={transferCore}>
													{fullName(row)}
												</TransferList.Item>
											{/each}
										</TransferList.Body>
									</TransferList.Container>
									<TransferList.Container>
										<TransferList.Title title="Members" />
										<TransferList.Toolbar
											variant="target"
											core={transferCore}
											inputPlaceholder="Search members..."
										/>
										<TransferList.Body>
											{#each transferCore.filteredTarget as row (row.id)}
												<TransferList.Item side="target" {row} core={transferCore}>
													{fullName(row)}
												</TransferList.Item>
											{/each}
										</TransferList.Body>
									</TransferList.Container>
								</TransferList.Root>
								<div class="mt-4 flex items-center gap-3">
									<Button
										type="button"
										onclick={handleSaveMembers}
										disabled={savingMembers}
									>
										{savingMembers ? 'Saving...' : 'Save members'}
									</Button>
									<span class="text-sm text-muted-foreground">
										{transferCore.target.length} member{transferCore.target.length === 1 ? '' : 's'}
									</span>
								</div>
							</div>
						</CardContent>
					</Card>
				{:else}
					<p class="text-sm text-muted-foreground">Loading members...</p>
				{/if}
			{:else}
				<Card class="max-w-md">
					<CardHeader>
						<CardTitle class="text-base">Members</CardTitle>
					</CardHeader>
					<CardContent>
						{#if members.length === 0}
							<p class="text-sm text-muted-foreground">No members yet.</p>
						{:else}
							<ul class="flex flex-col gap-2">
								{#each members as member}
									<li class="text-sm">{fullName(member)} ({member.email})</li>
								{/each}
							</ul>
						{/if}
					</CardContent>
				</Card>
			{/if}
		</TabsContent>
	</Tabs>
{/if}
