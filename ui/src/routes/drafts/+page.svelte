<script lang="ts">
	import { onMount } from 'svelte';
	import { listDrafts } from '$lib/draft';
	import type { Draft } from '$lib/draft';
	import { listFormats } from '$lib/format';
	import { listUsers, fullName } from '$lib/user';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let drafts = $state<Draft[]>([]);
	let formats = $state<Record<string, string>>({});
	let users = $state<Record<string, string>>({});
	let loading = $state(true);
	let error = $state('');

	// The zero Go time marshals to this literal; treat it as "not set".
	const ZERO_TIME = '0001-01-01T00:00:00Z';

	function isCompleted(draft: Draft): boolean {
		return !!draft.completed_at && draft.completed_at !== ZERO_TIME;
	}

	onMount(async () => {
		try {
			const [draftList, formatList, userList] = await Promise.all([
				listDrafts(),
				listFormats(),
				listUsers()
			]);
			drafts = draftList;
			formats = Object.fromEntries(formatList.map((f) => [f.id, f.name]));
			users = Object.fromEntries(userList.map((u) => [u.id, fullName(u)]));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load drafts';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Drafts</title>
</svelte:head>

<div class="flex items-center justify-between gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Drafts</h1>
	<Button href="/drafts/new">New draft</Button>
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{:else if drafts.length === 0}
	<p class="text-muted-foreground">No drafts yet.</p>
{:else}
	<div class="mt-4 overflow-hidden rounded-lg border">
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead>Name</TableHead>
					<TableHead>Owner</TableHead>
					<TableHead>Format</TableHead>
					<TableHead>Status</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each drafts as draft}
					<TableRow>
						<TableCell>
							<a
								href={`/drafts/${draft.id}`}
								class="font-medium text-primary underline-offset-4 hover:underline"
							>
								{draft.name}
							</a>
						</TableCell>
						<TableCell>{users[draft.owner] ?? draft.owner}</TableCell>
						<TableCell>{formats[draft.format] ?? draft.format}</TableCell>
						<TableCell>
							<Badge variant={isCompleted(draft) ? 'secondary' : 'default'}>
								{isCompleted(draft) ? 'Completed' : 'In progress'}
							</Badge>
						</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
