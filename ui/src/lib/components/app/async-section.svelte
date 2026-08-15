<script lang="ts" generics="T">
	import type { Snippet } from 'svelte';
	import type { Async } from '$lib/async.svelte';
	import { Alert, AlertDescription } from '$lib/components/ui/alert/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';

	let {
		state,
		isEmpty,
		loading,
		error,
		empty,
		children
	}: {
		/** The `Async` instance driving this section. */
		state: Async<T>;
		/** Predicate deciding whether the resolved data counts as "empty". */
		isEmpty?: (d: T) => boolean;
		/** Custom loading snippet (defaults to skeleton rows). */
		loading?: Snippet;
		/** Custom error snippet, receiving the raw backend message. */
		error?: Snippet<[string]>;
		/** Custom empty snippet (defaults to a muted one-liner). */
		empty?: Snippet;
		/** Content rendered once data is ready and non-empty. */
		children: Snippet<[T]>;
	} = $props();
</script>

{#if state.status === 'error'}
	{@render (error ?? _error)(state.error)}
{:else if state.status === 'ready' && state.data !== undefined && (!isEmpty || !isEmpty(state.data))}
	{@render children(state.data)}
{:else if state.status === 'ready'}
	{@render (empty ?? _empty)()}
{:else}
	{@render (loading ?? _loading)()}
{/if}

<!-- The default error branch renders the raw backend message as visible text;
	 `role="alert"` does not affect getByText matching. -->
{#snippet _error(message: string)}
	<Alert variant="destructive">
		<TriangleAlertIcon class="size-4" />
		<AlertDescription>{message}</AlertDescription>
	</Alert>
{/snippet}

{#snippet _loading()}
	<div>
		<Skeleton class="h-8 w-full" />
		<Skeleton class="mt-3 h-8 w-full" />
	</div>
{/snippet}

{#snippet _empty()}
	<p class="py-4 text-sm text-muted-foreground">Nothing here yet.</p>
{/snippet}
