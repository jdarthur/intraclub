/**
 * Minimal async-state container built on Svelte 5 runes.
 *
 * Owns the four-way status a page needs — idle / loading / ready / error —
 * alongside the resolved data and the raw backend error message. `run` never
 * throws: failures are captured into `error` and `status` instead, so callers
 * don't need try/catch around the load path.
 *
 * ```ts
 * const facilities = new Async<Facility[]>();
 * onMount(() => facilities.run(() => listFacilities()));
 * ```
 *
 * Pair with `<AsyncSection>` (`$lib/components/app/`) to render the
 * loading / error / empty / content branch chain.
 */
export class Async<T> {
	status = $state<'idle' | 'loading' | 'ready' | 'error'>('idle');
	data = $state<T | undefined>(undefined);
	error = $state('');

	/** Run an async producer, updating `status` / `data` / `error`. Never throws. */
	async run(fn: () => Promise<T>): Promise<void> {
		this.status = 'loading';
		this.error = '';
		try {
			this.data = await fn();
			this.status = 'ready';
		} catch (e) {
			this.error = e instanceof Error ? e.message : 'Something went wrong';
			this.status = 'error';
		}
	}
}
