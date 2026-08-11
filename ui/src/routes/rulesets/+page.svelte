<script lang="ts">
	import { onMount } from 'svelte';
	import { listRulesets } from '$lib/ruleset';
	import type { Ruleset } from '$lib/ruleset';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let rulesets = $state<Ruleset[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			rulesets = await listRulesets();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load rulesets';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Rulesets</title>
</svelte:head>

<div class="flex items-center justify-between gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Rulesets</h1>
	<Button href="/rulesets/new">New ruleset</Button>
</div>

{#if loading}
	<p class="text-muted-foreground">Loading...</p>
{:else if error}
	<p class="text-sm font-medium text-destructive">{error}</p>
{:else if rulesets.length === 0}
	<p class="text-muted-foreground">No rulesets yet.</p>
{:else}
	<div class="mt-4 overflow-hidden rounded-lg border">
		<Table>
			<TableHeader>
				<TableRow>
					<TableHead>Name</TableHead>
					<TableHead>Revision</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each rulesets as ruleset}
					<TableRow>
						<TableCell>
							<a href={`/rulesets/${ruleset.id}`} class="font-medium text-primary underline-offset-4 hover:underline">
								{ruleset.name}
							</a>
						</TableCell>
						<TableCell>{ruleset.revision}</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</div>
{/if}
