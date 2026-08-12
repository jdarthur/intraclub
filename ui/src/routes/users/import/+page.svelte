<script lang="ts">
	import { importUsersFromCsv, type User } from '$lib/user';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';

	let csv = $state('');
	let error = $state('');
	let submitting = $state(false);
	let result = $state<{ created: User[]; existing: User[] } | null>(null);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		result = null;
		submitting = true;
		try {
			const res = await importUsersFromCsv(csv);
			result = { created: res.Created, existing: res.AlreadyExisting };
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to import users';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Import users</title>
</svelte:head>

<div class="flex items-center gap-4">
	<h1 class="text-2xl font-semibold tracking-tight">Import users</h1>
</div>

<Card class="mt-6 max-w-xl">
	<CardHeader>
		<CardTitle class="text-base">CSV import</CardTitle>
	</CardHeader>
	<CardContent>
		<form onsubmit={handleSubmit} class="flex flex-col gap-4">
			<div class="flex flex-col gap-2">
				<Label for="csv">CSV contents</Label>
				<Textarea
					id="csv"
					bind:value={csv}
					rows={10}
					placeholder={'First Name, Last Name, Email\nAda, Lovelace, ada@example.com\nGrace, Hopper, grace@example.com'}
					required
				/>
				<p class="text-sm text-muted-foreground">
					Headers must be <code>First Name, Last Name, Email</code> (any order). Users whose
					email already exists are skipped.
				</p>
			</div>
			<Button type="submit" disabled={submitting} class="w-fit">
				{submitting ? 'Importing...' : 'Import users'}
			</Button>
		</form>
	</CardContent>
</Card>

{#if error}
	<p class="mt-4 text-sm font-medium text-destructive">{error}</p>
{/if}

{#if result}
	<div class="mt-6 space-y-6">
		<div>
			<h2 class="text-lg font-semibold tracking-tight">Created ({result.created.length})</h2>
			{#if result.created.length === 0}
				<p class="text-sm text-muted-foreground">No new users were created.</p>
			{:else}
				<ul class="mt-2 space-y-1 text-sm">
					{#each result.created as user}
						<li>
							{user.first_name} {user.last_name}
							<span class="text-muted-foreground">&lt;{user.email}&gt;</span>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
		<div>
			<h2 class="text-lg font-semibold tracking-tight">Already existing ({result.existing.length})</h2>
			{#if result.existing.length === 0}
				<p class="text-sm text-muted-foreground">No existing users matched.</p>
			{:else}
				<ul class="mt-2 space-y-1 text-sm">
					{#each result.existing as user}
						<li>
							{user.first_name} {user.last_name}
							<span class="text-muted-foreground">&lt;{user.email}&gt;</span>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	</div>
{/if}
