<script lang="ts">
	import type { Component, Snippet } from 'svelte';
	import { cn } from '$lib/utils.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import InboxIcon from '@lucide/svelte/icons/inbox';

	let {
		title,
		description,
		icon,
		actionHref,
		actionLabel,
		children,
		class: className
	}: {
		/** Empty-state headline. */
		title: string;
		/** Optional supporting copy. */
		description?: string;
		/** Optional Lucide icon component (defaults to `Inbox`). */
		icon?: Component;
		/** Optional call-to-action link destination. */
		actionHref?: string;
		/** Optional call-to-action link text. Becomes the link's accessible
		 *  name — on list pages that already have a matching header CTA, use a
		 *  distinct phrase (e.g. "Add your first facility") or omit it. */
		actionLabel?: string;
		/** Extra content rendered below the title / CTA. */
		children?: Snippet;
		class?: string;
	} = $props();
</script>

<div
	class={cn(
		'flex flex-col items-center justify-center gap-1.5 rounded-lg border border-dashed px-6 py-12 text-center',
		className
	)}
>
	{#if icon}
		{@const Icon = icon}
		<Icon class="size-8 text-muted-foreground" />
	{:else}
		<InboxIcon class="size-8 text-muted-foreground" />
	{/if}
	<p class="text-base font-medium">{title}</p>
	{#if description}
		<p class="max-w-sm text-sm text-muted-foreground">{description}</p>
	{/if}
	{#if actionHref && actionLabel}
		<Button href={actionHref} variant="outline" class="mt-2">
			{actionLabel}
		</Button>
	{/if}
	{@render children?.()}
</div>
