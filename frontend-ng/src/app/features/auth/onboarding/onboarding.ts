import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { TuiButton } from '@taiga-ui/core/components/button';

import { ApiService } from '../../../core/api/api-service';
import { routePaths } from '../../../core/api/route-paths';
import { AuthService } from '../../../core/auth/auth-service';
import { User } from '../../../shared/models/user.model';

// Shown to fresh signups (ShowdownUsername is still null) so they pick their
// in-game identity before using the app. PUTs /api/users/profile.
@Component({
  selector: 'app-onboarding',
  imports: [TuiButton],
  templateUrl: './onboarding.html',
  styleUrl: './onboarding.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class Onboarding {
  private readonly api = inject(ApiService);
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  protected readonly user = this.auth.user;
  protected readonly username = signal('');
  protected readonly saving = signal(false);
  protected readonly error = signal<string | null>(null);

  constructor() {
    // Already onboarded (or logged out) -> nothing to do here.
    if (!this.auth.isLoggedIn() || !this.auth.needsOnboarding()) {
      void this.router.navigateByUrl('/my-leagues');
    }
  }

  protected async submit(): Promise<void> {
    const showdownUsername = this.username().trim();
    if (!showdownUsername || this.saving()) return;

    this.saving.set(true);
    this.error.set(null);
    try {
      await new Promise<void>((resolve, reject) =>
        this.api.put<User>(routePaths.users.profile, { ShowdownName: showdownUsername }).subscribe({
          next: () => resolve(),
          error: reject,
        }),
      );
      await this.auth.refreshUser();
      if (this.auth.needsOnboarding()) {
        // e.g. the name was taken server-side; let them retry with a fresh form.
        this.error.set('Could not save that username. Is it already taken?');
        return;
      }
      await this.router.navigateByUrl('/my-leagues');
    } catch {
      this.error.set('Could not save your username. Try again.');
    } finally {
      this.saving.set(false);
    }
  }
}
