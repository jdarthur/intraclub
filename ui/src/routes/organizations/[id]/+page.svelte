<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Async } from '$lib/async.svelte';
	import { AsyncSection, PageHeader } from '$lib/components/app/index.js';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import {
		AlertDialog,
		AlertDialogAction,
		AlertDialogCancel,
		AlertDialogContent,
		AlertDialogDescription,
		AlertDialogFooter,
		AlertDialogHeader,
		AlertDialogTitle,
		AlertDialogTrigger
	} from '$lib/components/ui/alert-dialog/index.js';
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
		Tabs,
		TabsContent,
		TabsList,
		TabsTrigger
	} from '$lib/components/ui/tabs/index.js';
	import * as TransferList from '$lib/components/ui/transfer-list/index.js';
	import { toast } from '$lib/toast';
	import BuildingIcon from '@lucide/svelte/icons/building';

	const id = () => page.params.id as string;

	const organization = new Async<Organization>();
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

	onMount(() => organization.run(load));

	async function load(): Promise<Organization> {
		const org = await getOrganization(id());
		name = org.name;
		isOwner = getCurrentUserId() === org.user_id;

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
		return org;
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const updated = await updateOrganization(id(), { name: name.trim() });
			organization.data = updated;
			name = updated.name;
			toast.success('Organization saved');
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
			toast.success('Members saved');
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
			toast.success('Organization deleted');
			await goto('/organizations');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete organization';
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | {organization.data ? organization.data.name : 'Organization'}</title>
</svelte:head>

<PageHeader
	title={organization.data?.name}
	icon={BuildingIcon}
	backHref="/organizations"
	backLabel="Back to organizations"
/>

<AsyncSection state={organization}>
	{#snippet children(org)}
		{#if error}
			<Alert variant="destructive" class="mt-4">
				<AlertDescription>{error}</AlertDescription>
			</Alert>
		{/if}

		<Tabs bind:value={activeTab} orientation="vertical" class="mt-6 flex gap-6">
			<TabsList class="h-fit w-56 shrink-0 items-stretch">
				<TabsTrigger
					value="details"
					class="group justify-start gap-2.5 px-3 py-2 text-base"
				>
					<span
						class="size-1.5 shrink-0 rounded-full bg-primary opacity-0 transition-opacity group-data-[state=active]:opacity-100"
						aria-hidden="true"
					></span>
					Details
				</TabsTrigger>
				<TabsTrigger
					value="members"
					class="group justify-start gap-2.5 px-3 py-2 text-base"
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
							<CardTitle>Organization details</CardTitle>
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
						<AlertDialog bind:open={deleteOpen}>
							<AlertDialogTrigger
								disabled={deleting}
								class={buttonVariants({ variant: 'destructive' })}
							>
								{deleting ? 'Deleting...' : 'Delete organization'}
							</AlertDialogTrigger>
							<AlertDialogContent>
								<AlertDialogHeader>
									<AlertDialogTitle>Delete organization?</AlertDialogTitle>
									<AlertDialogDescription>
										This permanently removes this organization and cannot be undone.
									</AlertDialogDescription>
								</AlertDialogHeader>
								<AlertDialogFooter>
									<AlertDialogCancel>Cancel</AlertDialogCancel>
									<AlertDialogAction variant="destructive" onclick={handleDelete}>
										Delete
									</AlertDialogAction>
								</AlertDialogFooter>
							</AlertDialogContent>
						</AlertDialog>
					</div>
				{:else}
					<Card class="max-w-md">
						<CardHeader>
							<CardTitle>Organization</CardTitle>
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
					<Alert variant="destructive">
						<AlertDescription>{membersError}</AlertDescription>
					</Alert>
				{/if}

				{#if isOwner}
					{#if transferCore}
						<Card class="max-w-2xl">
							<CardHeader>
								<CardTitle>Members</CardTitle>
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
							<CardTitle>Members</CardTitle>
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
	{/snippet}
</AsyncSection>
