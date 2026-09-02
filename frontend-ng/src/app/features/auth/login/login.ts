import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { Router } from '@angular/router';
import { TuiButton } from '@taiga-ui/core/components/button';

import { routePaths } from '../../../core/api/route-paths';
import { AuthService, LoginError } from '../../../core/auth/auth-service';

@Component({
  selector: 'app-login',
  imports: [TuiButton],
  templateUrl: './login.html',
  styleUrl: './login.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class Login {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);

  protected readonly state = signal<'idle' | 'logging-in'>('idle');
  protected readonly error = signal<string | null>(null);

  constructor() {
    if (this.auth.isLoggedIn()) void this.router.navigateByUrl('/my-leagues');
  }

  protected async login(): Promise<void> {
    if (this.state() === 'logging-in') return;
    this.state.set('logging-in');
    this.error.set(null);

    try {
      await this.auth.login();
      await this.router.navigateByUrl('/my-leagues');
    } catch (err) {
      if (err instanceof LoginError && err.code === 'popup-blocked') {
        // Popup blocked -> full-page fallback through the dev proxy; the callback
        // completes in-place and lands on /my-leagues.
        window.location.href = routePaths.auth.discordLogin;
        return;
      }
      this.error.set(
        err instanceof LoginError ? 'Login window was closed before completing.' : 'Login failed. Try again.',
      );
      this.state.set('idle');
    }
  }
}
