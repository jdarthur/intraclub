<script lang="ts">
	import type { Component, Snippet } from 'svelte';
	import { cn } from '$lib/utils.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';

	let {
		title,
		description,
		icon,
		backHref,
		backLabel,
		actions,
		children,
		class: className
	}: {
		/** Page heading. When `undefined` (still loading) a `Skeleton` renders
		 *  instead of an `<h1>`, so `getByRole('heading')` stays a valid
		 *  "data has loaded" barrier. */
		title?: string;
		/** Optional subtitle under the heading. */
		description?: string;
		/** Optional Lucide icon component rendered beside the heading. */
		icon?: Component;
		/** When set, renders the normalised "← Back" affordance. */
		backHref?: string;
		/** Link text for the back affordance (defaults to "Back"). */
		backLabel?: string;
		/** Header actions (buttons / links), rendered on the right. */
		actions?: Snippet;
		/** Extra content rendered below the header row. */
		children?: Snippet;
		class?: string;
	} = $props();
</script>

<div class={cn('flex flex-wrap items-center justify-between gap-4', className)}>
	<div class="flex min-w-0 items-center gap-4">
		{#if title !== undefined}
			<div class="flex min-w-0 items-center gap-3">
				{#if icon}
					{@const Icon = icon}
					<Icon class="size-6 shrink-0 text-muted-foreground" />
				{/if}
				<div class="min-w-0">
					<h1 class="truncate text-2xl font-semibold tracking-tight">{title}</h1>
					{#if description}
						<p class="mt-0.5 text-sm text-muted-foreground">{description}</p>
					{/if}
				</div>
			</div>
		{:else}
			<Skeleton class="h-8 w-56" />
		{/if}
		{#if backHref}
			<a
				href={backHref}
				class="flex shrink-0 items-center gap-1 text-sm text-muted-foreground transition-colors hover:text-foreground"
			>
				<ArrowLeftIcon class="size-4" />
				{backLabel ?? 'Back'}
			</a>
		{/if}
	</div>
	{#if actions}
		<div class="flex shrink-0 items-center gap-2">
			{@render actions()}
		</div>
	{/if}
</div>
{@render children?.()}
