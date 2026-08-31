import { Service, inject, signal } from '@angular/core';
import { firstValueFrom, shareReplay } from 'rxjs';
import type { Observable } from 'rxjs';

import { User } from '../../shared/models/user.model';
import { ApiService } from '../api/api-service';
import { routePaths } from '../api/route-paths';

export const TOKEN_KEY = 'jwt_token';

export class LoginError extends Error {
  constructor(readonly code: 'popup-blocked' | 'popup-closed') {
    super(code);
  }
}

/**
 * Auth state + the interim Discord OAuth handoff.
 *
 * INTERIM FLOW (dev-only, until the refresh-token rewrite): the whole round-trip runs on the SPA
 * origin through the dev proxy (proxy.conf.json forwards `/auth/*` and `/api/*` to the backend) —
 * Discord is never pointed at the backend:
 *
 *   1. `login()` opens a popup at `/auth/discord/login` (proxied). The backend sets a host-only
 *      `oauthstate` cookie for :4200 and 307s to Discord.
 *   2. Discord redirects to DISCORD_REDIRECT_URI = `http://localhost:4200/callback?code=&state=`,
 *      which loads CallbackComponent inside the popup (NOT proxied — served by the SPA).
 *   3. CallbackComponent self-fetches `/auth/discord/callback?code=&state=` (proxied; the oauthstate
 *      cookie rides along). The backend exchanges the code and returns `200 {"Token": "<jwt>"}` —
 *      a normal same-origin JSON response the SPA can read.
 *   4. CallbackComponent posts `{type: 'auth:success', token}` to `window.opener` (same-origin
 *      postMessage) and closes. If the popup was blocked (`window.opener === null`) it completes
 *      in-place instead; login always lands on `/my-leagues`.
 *   5. The main window calls `setToken()` → localStorage → `loadUser()` → `/api/users/me`.
 *
 * The httpOnly `token` cookie the backend also sets never survives the proxy (its Domain is the
 * backend host), so the JWT rides in localStorage and the authInterceptor forwards it as a Bearer
 * header. When the refresh-token rewrite lands, only the handoff surface changes: `login()` +
 * `setToken()` + CallbackComponent become cookie-driven and this doc gets rewritten.
 */
@Service()
export class AuthService {
  private readonly api = inject(ApiService);

  readonly user = signal<User | null>(null);

  // Permanent session cache (idea borrowed from pdz)
  // ONE fetch of /api/users/me, shared app-wide. Reset by refreshUser().
  private me$: Observable<User> | null = null;

  private getMe(): Observable<User> {
    this.me$ ??= this.api.get<User>(routePaths.users.me).pipe(shareReplay({ refCount: false }));
    return this.me$;
  }

  // Persists the JWT from the interim handoff and fetches the current user.
  // Along with CallbackComponent's navigation, this is the whole handoff
  // surface the refresh-token rewrite replaces.
  setToken(jwt: string): Promise<void> {
    localStorage.setItem(TOKEN_KEY, jwt);
    this.me$ = null;
    return this.loadUser();
  }

  async loadUser(): Promise<void> {
    return firstValueFrom(this.getMe())
      .then((user) => this.user.set(user))
      .catch(() => {
        // If 401 or a network failure,
        // we drop stale session, and
        // let the guard redirecting to /login
        this.user.set(null);
        localStorage.removeItem(TOKEN_KEY);
      });
  }

  isLoggedIn(): boolean {
    return this.user() !== null;
  }

  // Re-fetches the user after profile changes (e.g. updating the Showdown username).
  async refreshUser(): Promise<void> {
    this.me$ = null;
    await this.loadUser();
  }

  logout(): void {
    this.api.post(routePaths.auth.logout).subscribe({ error: () => undefined });
    localStorage.removeItem(TOKEN_KEY);
    this.user.set(null);
  }

  // provideAppInitializer(() => inject(AuthService).prime())
  prime(): Promise<void> {
    if (!localStorage.getItem(TOKEN_KEY)) return Promise.resolve();
    return this.loadUser();
  }

  /**
   * Starts the interim Discord OAuth login in a popup and resolves once the handoff completes.
   *
   * Opens `/auth/discord/login` (dev-proxied) in a popup; the backend 307s to Discord and
   * CallbackComponent completes the exchange inside the popup, posting the JWT back here via
   * same-origin postMessage (source + origin are verified). The main window never reloads.
   *
   * Resolves after `setToken()` has loaded the user. Rejects with:
   *  - `LoginError('popup-blocked')` — the popup could not open; LoginComponent falls back to a
   *    full-page navigation to the login URL, where CallbackComponent completes in-place.
   *  - `LoginError('popup-closed')` — the popup was closed before the handshake finished.
   */
  login(): Promise<void> {
    return new Promise((resolve, reject) => {
      const popup = window.open(routePaths.auth.discordLogin, 'discord-auth', 'popup=yes,width=640,height=720');
      if (!popup) {
        reject(new LoginError('popup-blocked'));
        return;
      }

      const onMessage = (event: MessageEvent) => {
        if (event.source !== popup) return; // only our popup
        if (event.origin !== window.location.origin) return; // same-origin handshake
        const data = event.data as { type?: string; token?: string };
        if (data?.type === 'auth:success' && typeof data.token === 'string') {
          cleanup();
          void this.setToken(data.token).then(() => resolve());
        } else if (data?.type === 'auth:error') {
          cleanup();
          reject(new Error('Discord login failed'));
        }
      };

      const pollClosed = setInterval(() => {
        if (popup.closed) {
          cleanup();
          reject(new LoginError('popup-closed'));
        }
      }, 300);

      const cleanup = () => {
        clearInterval(pollClosed);
        window.removeEventListener('message', onMessage);
      };

      window.addEventListener('message', onMessage);
    });
  }
}
