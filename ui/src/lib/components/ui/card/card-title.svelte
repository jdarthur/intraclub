<script lang="ts">
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type { HTMLAttributes } from "svelte/elements";

	let {
		ref = $bindable(null),
		class: className,
		level = 3,
		children,
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLHeadingElement>> & {
		/** Heading level to render (`1`–`6`); defaults to `3`. */
		level?: 1 | 2 | 3 | 4 | 5 | 6;
	} = $props();
</script>

<svelte:element
	this={`h${level}`}
	bind:this={ref}
	data-slot="card-title"
	class={cn("text-base leading-snug font-medium group-data-[size=sm]/card:text-sm", className)}
	{...restProps}
>
	{@render children?.()}
</svelte:element>
