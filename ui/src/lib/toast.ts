/**
 * Thin wrapper around `svelte-sonner`'s `toast` so pages import toasts from
 * `$lib/toast` rather than reaching into the dependency directly.
 *
 * ```ts
 * import { toast } from '$lib/toast';
 * toast.success('Facility created');
 * ```
 *
 * Keep toast copy in verb-past form that appears nowhere else on the page —
 * toasts live ~4s and a strict `getByText` in that window would see two
 * matches. Do not pass `richColors` to `<Toaster />`: it hardcodes sonner's
 * own palette and bypasses the token system.
 */
export { toast } from 'svelte-sonner';
