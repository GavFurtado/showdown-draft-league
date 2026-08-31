import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TuiButton, TuiDataList, TuiDropdown, TuiInput, TuiOption, TuiTextfield } from '@taiga-ui/core';
import { TuiChevron, TuiComboBox, TuiDataListWrapper, TuiSelect } from '@taiga-ui/kit';
import { firstValueFrom } from 'rxjs';

import { ApiService } from '../../api/api-service';
import { routePaths } from '../../api/route-paths';
import { AuthService, TOKEN_KEY } from '../../auth/auth-service';
import { User } from '../../../shared/models/user.model';
import { MemberRole } from '../../../features/league/models/enums/member-role';

interface TokenResponse {
  Token: string;
}

/**
 * Dev-only panel for impersonating users without Discord OAuth.
 * Rendered by the shell when `environment.devTools` is on; the backing
 * endpoints only exist when the backend runs with ENV=dev.
 */
@Component({
  selector: 'app-dev-user-switcher',
  imports: [
    FormsModule,
    TuiButton,
    TuiChevron,
    TuiComboBox,
    TuiDataList,
    TuiDataListWrapper,
    TuiDropdown,
    TuiInput,
    TuiOption,
    TuiSelect,
    TuiTextfield,
  ],
  templateUrl: './dev-user-switcher.html',
  styleUrl: './dev-user-switcher.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class DevUserSwitcher {
  private readonly api = inject(ApiService);
  private readonly auth = inject(AuthService);

  readonly open = signal(false);
  readonly users = signal<User[]>([]);
  readonly busy = signal(false);
  readonly error = signal<string | null>(null);
  readonly success = signal<string | null>(null);

  readonly newName = signal('');
  readonly leagueId = signal('');
  readonly role = signal<MemberRole>(MemberRole.MEMBER);
  readonly selectedUserId = signal<string | null>(null);
  readonly roleTargetUserId = signal<string | null>(null);
  readonly roles = Object.values(MemberRole);

  readonly stringify = (id: string): string =>
    this.users().find((u) => u.ID === id)?.DiscordUsername ?? id;

  onOpenChange(opened: boolean): void {
    this.open.set(opened);
    if (opened) void this.refresh();
  }

  async refresh(): Promise<void> {
    this.error.set(null);
    this.success.set(null);
    try {
      const users = await firstValueFrom(this.api.get<User[]>(routePaths.auth.dev.users));
      this.users.set(users);
      // keep selections valid after list changes
      const isKnown = (id: string | null) => !id || users.some((u) => u.ID === id);
      if (!isKnown(this.selectedUserId())) this.selectedUserId.set(null);
      if (!isKnown(this.roleTargetUserId())) this.roleTargetUserId.set(null);

      // default the impersonate selection to the current session user so the
      // list highlights them as "current" and the combobox shows their name.
      if (!this.selectedUserId()) this.selectedUserId.set(this.auth.user()?.ID ?? null);
      // default the role target to the current session user too.
      if (!this.roleTargetUserId()) this.roleTargetUserId.set(this.auth.user()?.ID ?? null);
    } catch {
      this.error.set('Is the backend running with ENV=dev?');
    }
  }

  async createUser(): Promise<void> {
    const name = this.newName().trim();
    if (!name || this.busy()) return;

    this.busy.set(true);
    this.error.set(null);
    this.success.set(null);
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
   * Selecting a user from the list marks it active and swaps the session to
   * that user's real JWT (then reloads so every service re-hydrates).
   */
  async selectUser(userId: string): Promise<void> {
    if (userId === this.auth.user()?.ID) {
      this.selectedUserId.set(userId);
      return;
    }

    this.selectedUserId.set(userId);
    if (this.busy()) return;

    this.busy.set(true);
    this.error.set(null);
    try {
      const { Token: token } = await firstValueFrom(
        this.api.post<TokenResponse>(routePaths.auth.dev.login, { UserId: userId }, { suppressErrorReporting: true }),
      );
      localStorage.setItem(TOKEN_KEY, token);
      window.location.reload();
    } catch {
      this.error.set('Impersonation failed.');
      this.busy.set(false);
    }
  }

  async upsertMembership(): Promise<void> {
    const userId = this.roleTargetUserId() ?? this.auth.user()?.ID;
    const leagueId = this.leagueId().trim();
    if (!userId || !leagueId) {
      this.error.set('Pick a user and paste a league ID first.');
      return;
    }
    if (this.busy()) return;

    this.busy.set(true);
    this.error.set(null);
    this.success.set(null);
    try {
      await firstValueFrom(
        this.api.post(
          routePaths.auth.dev.memberships,
          { LeagueId: leagueId, UserId: userId, Role: this.role() },
          { suppressErrorReporting: true },
        ),
      );
      this.success.set(`Role set for ${this.stringify(userId)}.`);
    } catch {
      this.error.set('Could not set membership. Check the league ID.');
    } finally {
      this.busy.set(false);
    }
  }
}
