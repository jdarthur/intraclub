<script lang="ts" module>
	import type { Snippet } from 'svelte';

	export type Column<T> = {
		/** Stable identifier used for sorting and filtering. */
		key: string;
		/** Header text. */
		header: string;
		/** Custom cell content; falls back to `value` / the raw field. */
		cell?: Snippet<[T]>;
		/** String/number extractor used for sorting and filtering. */
		value?: (row: T) => string | number;
		/** Opt in per column — a sort control adds a `<button>` to the `<th>`. */
		sortable?: boolean;
		align?: 'left' | 'right';
		/** Extra classes for the `<th>` / `<td>`. */
		class?: string;
		/** Hide this column below the given breakpoint; the value is restated
		 *  in a stacked line inside the primary cell on small screens. */
		hideBelow?: 'sm' | 'md' | 'lg';
	};
</script>

<script lang="ts" generics="T">
	import { cn } from '$lib/utils.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
	import SearchIcon from '@lucide/svelte/icons/search';
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow
	} from '$lib/components/ui/table/index.js';

	let {
		rows,
		columns,
		getKey,
		caption,
		filter = false,
		filterLabel = 'Filter',
		pageSize,
		loading = false,
		empty,
		class: className
	}: {
		rows: T[];
		columns: Column<T>[];
		/** Stable per-row key (e.g. `(row) => row.id`). */
		getKey: (row: T) => string;
		/** Screen-reader-only table caption. */
		caption?: string;
		/** Show a filter input above the table. */
		filter?: boolean;
		/** Accessible label for the filter input — never a bare field name. */
		filterLabel?: string;
		/** Paginate to this many rows per page (no pagination when unset). */
		pageSize?: number;
		loading?: boolean;
		/** Custom empty-state snippet. */
		empty?: Snippet;
		class?: string;
	} = $props();

	let query = $state('');
	let sortKey = $state<string | null>(null);
	let sortDir = $state<'asc' | 'desc' | null>(null);
	let currentPage = $state(1);

	function fieldValue(row: T, col: Column<T>): string | number | undefined {
		if (col.value) return col.value(row);
		const v = (row as Record<string, unknown>)[col.key];
		return typeof v === 'string' || typeof v === 'number' ? v : undefined;
	}

	const filtered = $derived(
		query.trim() === ''
			? rows
			: rows.filter((row) =>
					columns.some((col) => {
						const v = fieldValue(row, col);
						return v !== undefined && String(v).toLowerCase().includes(query.trim().toLowerCase());
					})
				)
	);

	const sorted = $derived(
		sortKey === null || sortDir === null
			? filtered
			: [...filtered].sort((a, b) => {
					const col = columns.find((c) => c.key === sortKey);
					if (!col) return 0;
					const va = fieldValue(a, col);
					const vb = fieldValue(b, col);
					const cmp = String(va ?? '').localeCompare(String(vb ?? ''), undefined, {
						numeric: true,
						sensitivity: 'base'
					});
					return sortDir === 'asc' ? cmp : -cmp;
				})
	);

	const total = $derived(sorted.length);
	const pageCount = $derived(pageSize ? Math.max(1, Math.ceil(total / pageSize)) : 1);
	const page = $derived(Math.min(currentPage, pageCount));
	const paged = $derived(
		pageSize ? sorted.slice((page - 1) * pageSize, page * pageSize) : sorted
	);

	// A filter/sort change can shrink the result set below the current page.
	$effect(() => {
		if (currentPage > pageCount) currentPage = 1;
	});

	function toggleSort(key: string) {
		if (sortKey !== key) {
			sortKey = key;
			sortDir = 'asc';
		} else if (sortDir === 'asc') {
			sortDir = 'desc';
		} else {
			sortKey = null;
			sortDir = null;
		}
	}

	function ariaSort(key: string): 'ascending' | 'descending' | 'none' {
		if (sortKey !== key) return 'none';
		return sortDir === 'asc' ? 'ascending' : 'descending';
	}

	function hideBelowClass(level?: 'sm' | 'md' | 'lg') {
		if (level === 'sm') return 'hidden sm:table-cell';
		if (level === 'md') return 'hidden md:table-cell';
		if (level === 'lg') return 'hidden lg:table-cell';
		return '';
	}

	const hiddenColumns = $derived(columns.filter((c) => c.hideBelow));
</script>

<!-- The wrapper carries the border; the inner table container (already
	 overflow-x-auto in table.svelte) is the only scroll context, so the table
	 scrolls horizontally at 375px with no nested overflow-hidden. -->
<div class={cn('rounded-lg border bg-card', className)}>
	{#if filter}
		<div class="flex items-center gap-2 border-b px-3 py-2">
			<SearchIcon class="size-4 shrink-0 text-muted-foreground" />
			<Input
				bind:value={query}
				type="search"
				aria-label={filterLabel}
				placeholder={filterLabel}
				class="h-8 max-w-60"
			/>
		</div>
	{/if}
	<Table>
		{#if caption}
			<caption class="sr-only">{caption}</caption>
		{/if}
		<TableHeader>
			<TableRow>
				{#each columns as col}
					<TableHead
						aria-sort={col.sortable ? ariaSort(col.key) : undefined}
						class={cn(
							col.align === 'right' && 'text-right',
							hideBelowClass(col.hideBelow),
							col.class
						)}
					>
						{#if col.sortable}
							<button
								type="button"
								onclick={() => toggleSort(col.key)}
								class="inline-flex items-center gap-1 font-medium transition-colors hover:text-foreground"
							>
								{col.header}
								<ChevronsUpDownIcon class="size-3.5" />
							</button>
						{:else}
							{col.header}
						{/if}
					</TableHead>
				{/each}
			</TableRow>
		</TableHeader>
		<TableBody>
			{#if loading}
				{#each Array(5) as _, i}
					<TableRow>
						{#each columns as col}
							<TableCell class={hideBelowClass(col.hideBelow)}>
								<Skeleton class="h-4 w-full" />
							</TableCell>
						{/each}
					</TableRow>
				{/each}
			{:else if paged.length === 0}
				<TableRow>
					<TableCell colspan={columns.length} class="h-24 text-center">
						{@render (empty ?? _empty)()}
					</TableCell>
				</TableRow>
			{:else}
				{#each paged as row (getKey(row))}
					<TableRow>
						{#each columns as col, i}
							<TableCell
								class={cn(
									col.align === 'right' && 'text-right',
									hideBelowClass(col.hideBelow),
									col.class
								)}
							>
								{#if i === 0}
									{@render _cell(row, col)}
									{#if hiddenColumns.length > 0}
										<div class="mt-1 space-y-0.5 md:hidden text-xs text-muted-foreground">
											{#each hiddenColumns as hcol}
												{#if fieldValue(row, hcol) !== undefined}
													<div>
														<span class="font-medium">{hcol.header}:</span>{' '}
														{fieldValue(row, hcol)}
													</div>
												{/if}
											{/each}
										</div>
									{/if}
								{:else}
									{@render _cell(row, col)}
								{/if}
							</TableCell>
						{/each}
					</TableRow>
				{/each}
			{/if}
		</TableBody>
	</Table>
	{#if pageSize && pageCount > 1}
		<div class="flex items-center justify-between gap-4 border-t px-3 py-2">
			<p class="text-sm text-muted-foreground">
				Showing {(page - 1) * pageSize + 1}–{Math.min(page * pageSize, total)} of {total}
			</p>
			<div class="flex items-center gap-1">
				<Button
					variant="outline"
					size="sm"
					disabled={page <= 1}
					onclick={() => (currentPage = page - 1)}
				>
					Previous
				</Button>
				<Button
					variant="outline"
					size="sm"
					disabled={page >= pageCount}
					onclick={() => (currentPage = page + 1)}
				>
					Next
				</Button>
			</div>
		</div>
	{/if}
</div>

{#snippet _cell(row: T, col: Column<T>)}
	{#if col.cell}
		{@render col.cell(row)}
	{:else}
		{fieldValue(row, col) ?? ''}
	{/if}
{/snippet}

{#snippet _empty()}
	<p class="text-sm text-muted-foreground">No results.</p>
{/snippet}
