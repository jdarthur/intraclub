// Identity store: the signed-in user's profile and role-bearing data. The
// user id comes from the JWT's `sub` claim (getCurrentUserId), and the profile
// is resolved via GET /api/user/:id — /api/whoami returns HTTP 400 and must
// not be used (see issue #192).
//
// Commissioner-ness is per-season and only derivable from ScheduleDetail.
// Rather than fetching schedules, the store is a **write-through cache**: pages
// that already call GET /api/schedule?season_id= push the result in via
// noteCommissioners(), so no extra requests are made.
import { SvelteMap } from 'svelte/reactivity';
import { getCurrentUserId } from '$lib/auth';
import { getUser, fullName, type User } from '$lib/user';
import { listTeams, type TeamRoster } from '$lib/team';

class Identity {
	userId = $state<string | null>(null);
	user = $state<User | null>(null);
	teams = $state<TeamRoster[]>([]);
	loading = $state(false);
	error = $state('');

	// seasonId -> commissioner user ids, as reported by GET /api/schedule.
	// A SvelteMap so writes are reactive. The raw list is stored (rather than a
	// boolean) so isCommissionerOf stays correct even when a page's note
	// arrives before identity.load() resolves — the check is re-derived from
	// both the note and userId, whichever lands last.
	private commissionersBySeason = new SvelteMap<string, string[]>();

	get displayName(): string {
		return this.user ? fullName(this.user) : '';
	}

	/** e.g. 'JA' — for the nav avatar chip. */
	get initials(): string {
		if (!this.user) return '';
		const first = this.user.first_name.trim().charAt(0);
		const last = this.user.last_name.trim().charAt(0);
		return `${first}${last}`.toUpperCase();
	}

	/** Team ids where the current user is captain or co-captain. */
	get captainOf(): string[] {
		if (!this.userId) return [];
		return this.teams
			.filter((roster) =>
				roster.assignments.some(
					(a) =>
						a.user_id === this.userId && (a.role === 'captain' || a.role === 'co_captain')
				)
			)
			.map((roster) => roster.team.id);
	}

	get isCaptain(): boolean {
		return this.captainOf.length > 0;
	}

	isCommissionerOf(seasonId: string): boolean {
		const commissioners = this.commissionersBySeason.get(seasonId);
		return (
			this.userId !== null && commissioners !== undefined && commissioners.includes(this.userId)
		);
	}

	/** Write-through cache: called by pages that already fetched the schedule. */
	noteCommissioners(seasonId: string, commissioners: string[]): void {
		this.commissionersBySeason.set(seasonId, commissioners);
	}

	/**
	 * Loads the signed-in user's profile via GET /api/user/:id. Idempotent and
	 * never rejects — errors are swallowed into `error` (an unhandled rejection
	 * in the layout effect would surface on every route). No-op when there is
	 * no decodable `sub` claim (no token, or a token without one), so it issues
	 * zero requests in that case.
	 */
	async load(): Promise<void> {
		const userId = getCurrentUserId();
		if (userId === null) {
			this.reset();
			return;
		}
		// Already resolved, or a request is in flight, for this user.
		if (this.userId === userId && (this.user !== null || this.loading)) return;
		this.userId = userId;
		this.loading = true;
		this.error = '';
		try {
			this.user = await getUser(userId);
		} catch (e) {
			this.error = e instanceof Error ? e.message : 'Failed to load identity';
		} finally {
			this.loading = false;
		}
	}

	/** Lazy: fetches the teams visible to the current user once. */
	async loadTeams(): Promise<void> {
		const userId = this.userId ?? getCurrentUserId();
		if (!userId) return;
		if (this.teams.length > 0) return;
		try {
			this.teams = await listTeams();
		} catch (e) {
			this.error = e instanceof Error ? e.message : 'Failed to load teams';
		}
	}

	/** Clears all cached identity state (logout / session expiry). */
	reset(): void {
		this.userId = null;
		this.user = null;
		this.teams = [];
		this.loading = false;
		this.error = '';
		this.commissionersBySeason.clear();
	}
}

export const identity = new Identity();
