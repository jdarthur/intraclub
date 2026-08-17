/**
 * Shared page primitives — deliberately separate from the vendored
 * `ui/src/lib/components/ui/` directory so `shadcn-svelte add` never
 * overwrites them.
 */
export { default as PageHeader } from './page-header.svelte';
export { default as EmptyState } from './empty-state.svelte';
export { default as StatCard, statCardTone, type StatCardTone } from './stat-card.svelte';
export { default as StatusBadge, statusConfig, type Status } from './status-badge.svelte';
export { default as AsyncSection } from './async-section.svelte';
export { default as DataTable, type Column } from './data-table.svelte';
export { default as PhotoPicker } from './photo-picker.svelte';
