import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { TuiButton, TuiDropdown, TuiIcon, TuiInput, TuiTextfield } from '@taiga-ui/core';
import { TuiChevron, TuiTooltip } from '@taiga-ui/kit';
import { firstValueFrom } from 'rxjs';

import { ApiService } from '../../api/api-service';
import { routePaths } from '../../api/route-paths';
import { ThemeService } from '../../theme/theme-service';
import { AuthService } from '../../auth/auth-service';
import { User } from '../../../shared/models/user.model';

/**
 * Profile menu: identity at the top, account preferences (optional Showdown
 * username, theme) below. Shown in the shell navbar whenever a user is signed in.
 */
@Component({
  selector: 'app-profile-menu',
  imports: [FormsModule, TuiButton, TuiChevron, TuiDropdown, TuiIcon, TuiInput, TuiTextfield, TuiTooltip],
  templateUrl: './profile-menu.html',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ProfileMenu {
  private readonly api = inject(ApiService);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  private readonly themeService = inject(ThemeService);

  readonly theme = this.themeService.isDark;
  readonly user = this.auth.user;
  readonly username = signal('');
  readonly open = signal(false);
  readonly saving = signal(false);
  readonly saved = signal(false);
  readonly error = signal<string | null>(null);

  toggleTheme(): void {
    this.themeService.toggle();
  }

  toggle(): void {
    const opening = !this.open();
    this.open.set(opening);
    if (opening) {
      this.username.set(this.user()?.ShowdownUsername ?? '');
      this.saved.set(false);
      this.error.set(null);
    }
  }

  async save(): Promise<void> {
    const value = this.username().trim();
    const current = this.user()?.ShowdownUsername;
    if (value === (current ?? '')) {
      this.open.set(false);
      return;
    }
    if (this.saving()) return;

    this.saving.set(true);
    this.error.set(null);
    try {
      await firstValueFrom(
        this.api.put<User>(routePaths.users.profile, { ShowdownName: value }, { suppressErrorReporting: true }),
      );
      await this.auth.refreshUser();
      this.saved.set(true);
    } catch {
      this.error.set('Could not save that username. Is it already taken?');
    } finally {
      this.saving.set(false);
    }
  }

  async clear(): Promise<void> {
    this.username.set('');
    await this.save();
  }

  logout(): void {
    this.open.set(false);
    this.auth.logout();
    void this.router.navigateByUrl('/');
  }
}
