import { HttpContext, HttpClient } from '@angular/common/http';
import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { TuiLoader } from '@taiga-ui/core/components/loader';
import { firstValueFrom } from 'rxjs';

import { routePaths } from '../../../core/api/route-paths';
import { AuthService } from '../../../core/auth/auth-service';
import { SUPPRESS_ERROR_REPORTING } from '../../../core/error/error-interceptor';

// OAuth redirect target (interim handoff). Runs inside the Discord login popup:
// self-fetches the backend callback (same-origin via the dev proxy, oauthstate
// cookie rides along), then posts the JWT back to the opener and closes. When
// opened directly (popup blocked, then  full-page fallback) it completes in-place.
@Component({
  selector: 'app-callback',
  imports: [TuiLoader],
  templateUrl: './callback.html',
  styleUrl: './callback.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class Callback {
  private readonly http = inject(HttpClient);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly auth = inject(AuthService);

  constructor() {
    void this.completeLogin();
  }

  private async completeLogin(): Promise<void> {
    const params = this.route.snapshot.queryParamMap;
    const code = params.get('code');
    const state = params.get('state');
    if (!code || !state) {
      this.finish(false);
      return;
    }

    try {
      const { Token: token } = await firstValueFrom(
        this.http.get<{ Token: string }>(routePaths.auth.discordCallback, {
          params: { code, state },
          context: new HttpContext().set(SUPPRESS_ERROR_REPORTING, true),
        }),
      );
      await this.auth.setToken(token);
      this.finish(true, token);
    } catch {
      this.finish(false);
    }
  }

  private finish(success: boolean, token?: string): void {
    if (window.opener) {
      // popup path -> hand the JWT back to the main window, then close
      window.opener.postMessage({ type: success ? 'auth:success' : 'auth:error', token }, window.location.origin);
      window.close();
    } else {
      // full-page fallback / direct visit → complete in-place
      void this.router.navigateByUrl(success ? '/my-leagues' : '/login');
    }
  }
}
