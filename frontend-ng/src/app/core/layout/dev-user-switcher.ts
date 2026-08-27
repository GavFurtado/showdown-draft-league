import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { TuiButton } from '@taiga-ui/core/components/button';
import { firstValueFrom } from 'rxjs';

import { ApiService } from '../api/api-service';
import { routePaths } from '../api/route-paths';
import { AuthService, TOKEN_KEY } from '../auth/auth-service';
import { User } from '../../shared/models/user.model';

interface TokenResponse {
  Token: string;
}

const ROLES = ['OWNER', 'MODERATOR', 'MEMBER'] as const;

/**
 * Dev-only panel for impersonating users without Discord OAuth.
 * Rendered by the shell when `environment.devTools` is on; the backing
 * endpoints only exist when the backend runs with ENV=dev.
 */
@Component({
  selector: 'app-dev-user-switcher',
  imports: [TuiButton],
  templateUrl: './dev-user-switcher.html',
  styleUrl: './dev-user-switcher.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DevUserSwitcher {
  private readonly api = inject(ApiService);
  private readonly auth = inject(AuthService);

  protected readonly open = signal(false);
  protected readonly users = signal<User[]>([]);
  protected readonly busy = signal(false);
  protected readonly error = signal<string | null>(null);

  protected readonly newName = signal('');
  protected readonly leagueId = signal('');
  protected readonly role = signal<(typeof ROLES)[number]>('MEMBER');
  protected readonly selectedUserId = signal<string | null>(null);
  protected readonly roles = ROLES;

  protected toggle(): void {
    this.open.update((open) => !open);
    if (this.open()) void this.refresh();
  }

  protected async refresh(): Promise<void> {
    this.error.set(null);
    try {
      const users = await firstValueFrom(this.api.get<User[]>(routePaths.auth.dev.users));
      this.users.set(users);
      // keep selection valid after list changes
      if (this.selectedUserId() && !users.some((u) => u.ID === this.selectedUserId())) {
        this.selectedUserId.set(null);
      }
    } catch {
      this.error.set('Is the backend running with ENV=dev?');
    }
  }

  protected async createUser(): Promise<void> {
    const name = this.newName().trim();
    if (!name || this.busy()) return;

    this.busy.set(true);
    this.error.set(null);
    try {
      await firstValueFrom(
        this.api.post<User>(routePaths.auth.dev.users, { Name: name }, { suppressErrorReporting: true }),
      );
      this.newName.set('');
      await this.refresh();
    } catch {
      this.error.set('Could not create user.');
    } finally {
      this.busy.set(false);
    }
  }

  /**
   * Swaps the local session for the selected user's real JWT and reloads so
   * every service re-hydrates against /api/users/me with the new identity.
   */
  protected impersonate(userId: string | null): void {
    if (!userId || this.busy()) return;
    if (userId === this.auth.user()?.ID) return; // already this user

    this.selectedUserId.set(userId);

    this.busy.set(true);
    this.error.set(null);
    this.api
      .post<TokenResponse>(routePaths.auth.dev.login, { UserId: userId }, { suppressErrorReporting: true })
      .subscribe({
        next: ({ Token }) => {
          localStorage.setItem(TOKEN_KEY, Token);
          window.location.reload();
        },
        error: () => {
          this.error.set('Impersonation failed.');
          this.busy.set(false);
        },
      });
  }

  protected async upsertMembership(): Promise<void> {
    const userId = this.selectedUserId() ?? this.auth.user()?.ID;
    const leagueId = this.leagueId().trim();
    if (!userId || !leagueId) {
      this.error.set('Pick a user and paste a league ID first.');
      return;
    }
    if (this.busy()) return;

    this.busy.set(true);
    this.error.set(null);
    try {
      await firstValueFrom(
        this.api.post(
          routePaths.auth.dev.memberships,
          { LeagueId: leagueId, UserId: userId, Role: this.role() },
          { suppressErrorReporting: true },
        ),
      );
    } catch {
      this.error.set('Could not set membership. Check the league ID.');
    } finally {
      this.busy.set(false);
    }
  }
}
