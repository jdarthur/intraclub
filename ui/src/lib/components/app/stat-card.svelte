<script lang="ts" module>
	import { tv, type VariantProps } from 'tailwind-variants';

	/** Tone drives the icon colour (and any accent), not the label. */
	export const statCardTone = tv({
		variants: {
			tone: {
				default: 'text-muted-foreground',
				primary: 'text-primary',
				success: 'text-success',
				warning: 'text-warning',
				info: 'text-info',
				destructive: 'text-destructive'
			}
		},
		defaultVariants: { tone: 'default' }
	});

	export type StatCardTone = VariantProps<typeof statCardTone>['tone'];
</script>

<script lang="ts">
	import type { Component, Snippet } from 'svelte';
	import { cn } from '$lib/utils.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';

	let {
		label,
		value,
		hint,
		icon,
		href,
		tone = 'default',
		loading = false,
		children,
		class: className
	}: {
		/** Stat name. Rendered as a `<p>`, never a heading. */
		label: string;
		/** The value. Rendered in `text-2xl font-semibold tabular-nums`. */
		value: string | number;
		/** Optional supporting line under the value. */
		hint?: string;
		/** Optional Lucide icon component, tinted by `tone`. */
		icon?: Component;
		/** When set, the value becomes a link. */
		href?: string;
		/** Colour tone applied to the icon. */
		tone?: StatCardTone;
		/** When true, renders skeleton placeholders instead of content. */
		loading?: boolean;
		/** Extra content rendered below the hint. */
		children?: Snippet;
		class?: string;
	} = $props();
</script>

{#if loading}
	<div class={cn('rounded-lg border p-4', className)}>
		<Skeleton class="h-4 w-24" />
		<Skeleton class="mt-2 h-8 w-16" />
	</div>
{:else}
	<div class={cn('rounded-lg border p-4', className)}>
		<div class="flex items-center justify-between gap-2">
			<p class="text-sm font-medium text-muted-foreground">{label}</p>
			{#if icon}
				{@const Icon = icon}
				<Icon class={cn('size-4 shrink-0', statCardTone({ tone }))} />
			{/if}
		</div>
		<div class="mt-1 text-2xl font-semibold tabular-nums">
			{#if href}
				<a href={href} class="transition-colors hover:text-foreground hover:underline">
					{value}
				</a>
			{:else}
				{value}
			{/if}
		</div>
		{#if hint}
			<p class="mt-1 text-xs text-muted-foreground">{hint}</p>
		{/if}
		{@render children?.()}
	</div>
{/if}
