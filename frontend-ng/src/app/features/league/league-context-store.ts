import { Service, computed, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';

import { AuthService } from '../../core/auth/auth-service';
import { ApiService } from '../../core/api/api-service';
import { ClientError } from '../../core/api/api.model';
import { routePaths } from '../../core/api/route-paths';
import { UUID } from '../../shared/types/branded-strings';
import { League } from './models/league.model';
import { LeagueMember } from './models/league-member.model';
import { MemberRole } from './models/enums/member-role';
import { Draft } from './models/draft.model';

@Service()
export class LeagueContextStore {
  private readonly api = inject(ApiService);
  private readonly auth = inject(AuthService);

  league = signal<League | null>(null);
  userMember = signal<LeagueMember | null>(null);

  currentDraft = signal<Draft | null>(null);

  readonly loading = signal(false);
  readonly loadError = signal<ClientError | null>(null);

  isOwner = computed(() => this.userMember()?.Role === MemberRole.OWNER);

  isModerator = computed(() => {
    const role = this.userMember()?.Role;
    return role === MemberRole.OWNER || role === MemberRole.MODERATOR;
  });

  // Tracks the one-time initial load so the league guard can cold-block once.
  private loaded = false;
  private loadPromise: Promise<void> | null = null;

  /**
   * Guards (and the shell) call this on first navigation. Blocks cold entry while the
   * context loads, so children render fully-populated. Subsequent navigations resolve
   * immediately because the store already holds the data.
   * Returns whether the current user is a verified member of the league.
   */
  async ensureLoaded(leagueId: UUID): Promise<boolean> {
    if (this.loaded) return this.userMember() !== null;
    this.loadPromise ??= this.refresh(leagueId).then(() => {
      this.loaded = true;
    });
    try {
      await this.loadPromise;
      return this.userMember() !== null;
    } catch {
      return false;
    }
  }

  /**
   * Re-fetches the league, the current user's membership, and the active draft.
   * Should be called after any mutation that changes league state.
   */
  async refresh(leagueId: UUID): Promise<void> {
    this.loading.set(true);
    this.loadError.set(null);
    try {
      // Fetch league + the current user's membership independently so a sparse
      // league.Members subset can never leave currentMember unpopulated.
      const [league, members] = await Promise.all([
        firstValueFrom(this.api.get<League>(routePaths.leagues.byId(leagueId))),
        firstValueFrom(this.api.get<LeagueMember[]>(routePaths.leagues.members.base(leagueId))),
      ]);
      this.league.set(league);

      const userId = this.auth.user()?.ID;
      this.userMember.set(members.find((m) => m.UserID === userId) ?? null);

      try {
        const draft = await firstValueFrom(
          this.api.get<Draft>(routePaths.leagues.draft.base(leagueId), {
            // Drafts aren't created until the DRAFTING phase
            suppressStatuses: [404],
          }),
        );
        this.currentDraft.set(draft);
      } catch {
        // No draft exists yet for this league (404); clear any stale value.
        this.currentDraft.set(null);
      }
    } catch (err) {
      this.loadError.set(err as ClientError);
      throw err;
    } finally {
      this.loading.set(false);
    }
  }
}
