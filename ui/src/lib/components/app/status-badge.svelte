<script lang="ts" module>
	import type { Component } from 'svelte';
	import type { BadgeVariant } from '$lib/components/ui/badge/index.js';
	import CircleDotIcon from '@lucide/svelte/icons/circle-dot';
	import LockIcon from '@lucide/svelte/icons/lock';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check';
	import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
	import CheckIcon from '@lucide/svelte/icons/check';
	import BadgeCheckIcon from '@lucide/svelte/icons/badge-check';
	import CircleXIcon from '@lucide/svelte/icons/circle-x';
	import CircleHelpIcon from '@lucide/svelte/icons/circle-help';
	import PencilLineIcon from '@lucide/svelte/icons/pencil-line';

	export type Status =
		| 'open'
		| 'closed'
		| 'pending'
		| 'confirmed'
		| 'official'
		| 'complete'
		| 'accepted'
		| 'rejected'
		| 'available'
		| 'maybe'
		| 'unavailable'
		| 'draft';

	/**
	 * Every status gets an icon as well as a colour: --success and --primary
	 * share a hue in the green theme, so colour alone cannot carry the
	 * distinction between e.g. "confirmed" and "official".
	 *
	 * `tone` classes ride on top of the vendored `Badge`'s structural variant
	 * (twMerge resolves the border/bg/text conflicts).
	 */
	export const statusConfig: Record<
		Status,
		{ variant: BadgeVariant; tone: string; icon: Component }
	> = {
		open: { variant: 'outline', tone: 'text-muted-foreground', icon: CircleDotIcon },
		closed: { variant: 'secondary', tone: 'text-muted-foreground', icon: LockIcon },
		pending: {
			variant: 'outline',
			tone: 'border-warning/40 bg-warning/10 text-warning',
			icon: ClockIcon
		},
		confirmed: {
			variant: 'outline',
			tone: 'border-success/40 bg-success/10 text-success',
			icon: CircleCheckIcon
		},
		official: {
			variant: 'outline',
			tone: 'border-primary/40 bg-primary/10 text-primary',
			icon: ShieldCheckIcon
		},
		complete: {
			variant: 'outline',
			tone: 'border-success/40 bg-success/10 text-success',
			icon: CheckIcon
		},
		accepted: {
			variant: 'outline',
			tone: 'border-success/40 bg-success/10 text-success',
			icon: BadgeCheckIcon
		},
		rejected: { variant: 'destructive', tone: '', icon: CircleXIcon },
		available: {
			variant: 'outline',
			tone: 'border-success/40 bg-success/10 text-success',
			icon: CircleCheckIcon
		},
		maybe: {
			variant: 'outline',
			tone: 'border-warning/40 bg-warning/10 text-warning',
			icon: CircleHelpIcon
		},
		unavailable: {
			variant: 'outline',
			tone: 'border-destructive/40 bg-destructive/10 text-destructive',
			icon: CircleXIcon
		},
		draft: { variant: 'secondary', tone: 'text-muted-foreground', icon: PencilLineIcon }
	};
</script>

<script lang="ts">
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { cn } from '$lib/utils.js';

	let {
		status,
		label,
		icon = true,
		class: className
	}: {
		/** Which status to render (drives colour + icon). */
		status: Status;
		/** Optional label; defaults to the status word itself. */
		label?: string;
		/** Whether to render the status icon (default true). */
		icon?: boolean;
		class?: string;
	} = $props();

	const config = $derived(statusConfig[status]);
</script>

<Badge variant={config.variant} class={cn(config.tone, className)}>
	{#if icon}
		{@const Icon = config.icon}
		<Icon />
	{/if}
	{label ?? status}
</Badge>
