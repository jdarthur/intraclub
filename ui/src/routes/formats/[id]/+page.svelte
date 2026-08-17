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
	import { getCurrentUserId } from '$lib/auth';
	import { toast } from '$lib/toast';
	import {
		getFormat,
		updateFormat,
		deleteFormat,
		getFormatPossibleRatings,
		setFormatPossibleRatings
	} from '$lib/format';
	import type { Format } from '$lib/format';
	import { listRatings } from '$lib/rating';
	import type { Rating } from '$lib/rating';
	import ListTreeIcon from '@lucide/svelte/icons/list-tree';

	const id = () => page.params.id as string;

	const format = new Async<Format>();
	let name = $state('');
	let error = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let deleteOpen = $state(false);
	let isOwner = $state(false);

	// Possible ratings assigned to this format.
	let possibleRatings = $state<Rating[]>([]);
	let allRatings = $state<Rating[]>([]);
	let ratingsError = $state('');
	let ratingsSaving = $state(false);

	// Rating assignments (owner only). The transfer list core holds the
	// in-memory source (unassigned ratings) / target (assigned ratings) lists
	// and selection; we persist diffs on save.
	let transferCore = $state<TransferList.Core<Rating> | null>(null);

	// Which section is shown in the sidebar at a time.
	let activeTab = $state('details');

	onMount(() => format.run(load));

	async function load(): Promise<Format> {
		const f = await getFormat(id());
		name = f.name;
		isOwner = getCurrentUserId() === f.user_id;

		try {
			possibleRatings = await getFormatPossibleRatings(id());
		} catch (e) {
			ratingsError = e instanceof Error ? e.message : 'Failed to load possible ratings';
		}

		// The owner can assign any rating in the catalog; build the transfer
		// list source from the ratings not already assigned.
		if (isOwner) {
			try {
				allRatings = await listRatings();
			} catch {
				// member-edit controls stay usable; the source just lists members
			}
			const assignedIds = new Set(possibleRatings.map((r) => r.id));
			transferCore = new TransferList.Core<Rating>({
				initialSource: allRatings.filter((r) => !assignedIds.has(r.id)),
				initialTarget: possibleRatings,
				filterPredicate: (r, search) => r.name.toLowerCase().includes(search.toLowerCase())
			});
		}
		return f;
	}

	async function handleSave(e: Event) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const updated = await updateFormat(id(), { name: name.trim() });
			format.data = updated;
			name = updated.name;
			toast.success('Format saved');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update format';
		} finally {
			saving = false;
		}
	}

	async function handleDelete() {
		deleteOpen = false;
		error = '';
		deleting = true;
		try {
			await deleteFormat(id());
			toast.success('Format deleted');
			await goto('/formats');
		} catch (err) {
			// e.g. the format is assigned to a Draft and PreDelete blocks it
			error = err instanceof Error ? err.message : 'Failed to delete format';
			deleting = false;
		}
	}

	// Persists the transfer-list state: the target order (highest-skill to
	// lowest-skill) is sent as-is and becomes the FormatRating.RatingIndex
	// ordering on the backend.
	async function handleSaveRatings() {
		ratingsError = '';
		if (!transferCore) return;
		const currentIds = possibleRatings.map((r) => r.id);
		const targetIds = transferCore.target.map((r) => r.id);
		if (targetIds.length === 0) {
			ratingsError = 'A format must keep at least one rating.';
			return;
		}
		if (currentIds.length === targetIds.length && currentIds.every((c, i) => c === targetIds[i])) {
			return;
		}
		ratingsSaving = true;
		try {
			possibleRatings = await setFormatPossibleRatings(id(), targetIds);
		} catch (err) {
			ratingsError = err instanceof Error ? err.message : 'Failed to update possible ratings';
		} finally {
			ratingsSaving = false;
		}
	}
</script>

<svelte:head>
	<title>Intraclub | {format.data ? format.data.name : 'Format'}</title>
</svelte:head>

<PageHeader
	title={format.data?.name}
	icon={ListTreeIcon}
	backHref="/formats"
	backLabel="Back to formats"
/>

<AsyncSection state={format}>
	{#snippet children(f)}
		{#if error}
			<Alert variant="destructive" class="mt-4">
				<AlertDescription>{error}</AlertDescription>
			</Alert>
		{/if}

		<Tabs bind:value={activeTab} orientation="vertical" class="mt-6 flex flex-col gap-6 md:flex-row">
			<TabsList class="h-fit w-full items-stretch md:w-56 md:shrink-0">
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
					value="ratings"
					class="group justify-start gap-2.5 px-3 py-2 text-base"
				>
					<span
						class="size-1.5 shrink-0 rounded-full bg-primary opacity-0 transition-opacity group-data-[state=active]:opacity-100"
						aria-hidden="true"
					></span>
					Ratings
				</TabsTrigger>
			</TabsList>

			<TabsContent value="details" class="flex-1">
				{#if isOwner}
					<Card class="max-w-md">
						<CardHeader>
							<CardTitle>Format details</CardTitle>
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
								{deleting ? 'Deleting...' : 'Delete format'}
							</AlertDialogTrigger>
							<AlertDialogContent>
								<AlertDialogHeader>
									<AlertDialogTitle>Delete format?</AlertDialogTitle>
									<AlertDialogDescription>
										This permanently removes this format and cannot be undone.
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
							<CardTitle>Format</CardTitle>
						</CardHeader>
						<CardContent>
							<p class="text-sm text-muted-foreground">
								Only the owner can edit this format's details.
							</p>
						</CardContent>
					</Card>
				{/if}
			</TabsContent>

			<TabsContent value="ratings" class="flex-1">
				{#if ratingsError}
					<Alert variant="destructive">
						<AlertDescription>{ratingsError}</AlertDescription>
					</Alert>
				{/if}

				{#if isOwner}
					{#if transferCore}
						<Card class="max-w-2xl">
							<CardHeader>
								<CardTitle>Possible ratings</CardTitle>
							</CardHeader>
							<CardContent>
								<p class="text-sm text-muted-foreground">
									Move ratings into the roster to make them assignable in this format.
									The order shown (highest-skill to lowest-skill) is preserved when you
									save. Changes apply when you save.
								</p>
								<div class="mt-4">
									<TransferList.Root direction="horizontal">
										<TransferList.Container>
											<TransferList.Title title="Available ratings" />
											<TransferList.Toolbar
												variant="source"
												core={transferCore}
												inputPlaceholder="Search ratings..."
											/>
											<TransferList.Body>
												{#each transferCore.filteredSource as row (row.id)}
													<TransferList.Item side="source" {row} core={transferCore}>
														{row.name}
													</TransferList.Item>
												{/each}
											</TransferList.Body>
										</TransferList.Container>
										<TransferList.Container>
											<TransferList.Title title="Assigned ratings" />
											<TransferList.Toolbar
												variant="target"
												core={transferCore}
												inputPlaceholder="Search ratings..."
											/>
											<TransferList.Body>
												{#each transferCore.filteredTarget as row (row.id)}
													<TransferList.Item side="target" {row} core={transferCore}>
														{row.name}
													</TransferList.Item>
												{/each}
											</TransferList.Body>
										</TransferList.Container>
									</TransferList.Root>
									<div class="mt-4 flex items-center gap-3">
										<Button
											type="button"
											onclick={handleSaveRatings}
											disabled={ratingsSaving || transferCore.target.length === 0}
										>
											{ratingsSaving ? 'Saving...' : 'Save ratings'}
										</Button>
										<span class="text-sm text-muted-foreground">
											{transferCore.target.length} rating{transferCore.target.length === 1 ? '' : 's'}
										</span>
									</div>
									{#if transferCore.target.length === 0}
										<p class="mt-2 text-sm text-muted-foreground">
											A format must keep at least one rating.
										</p>
									{/if}
								</div>
							</CardContent>
						</Card>
					{:else}
						<p class="text-sm text-muted-foreground">Loading ratings...</p>
					{/if}
				{:else}
					<Card class="max-w-md">
						<CardHeader>
							<CardTitle>Possible ratings</CardTitle>
						</CardHeader>
						<CardContent>
							{#if possibleRatings.length === 0}
								<p class="text-sm text-muted-foreground">No ratings assigned yet.</p>
							{:else}
								<ul class="flex flex-col gap-2">
									{#each possibleRatings as rating}
										<li class="text-sm">{rating.name}</li>
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
