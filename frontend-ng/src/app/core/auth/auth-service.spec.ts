import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { User } from '../../shared/models/user.model';
import { AuthService, LoginError, TOKEN_KEY } from './auth-service';

const user: User = {
  ID: 'user-1',
  DiscordID: '1234',
  DiscordUsername: 'Gavin',
  DiscordAvatarURL: '',
  ShowdownUsername: 'GavinTest',
  Role: 'user',
};

describe('AuthService', () => {
  let service: AuthService;
  let http: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({ providers: [provideHttpClientTesting()] });
    service = TestBed.inject(AuthService);
    http = TestBed.inject(HttpTestingController);
    localStorage.clear();
  });

  afterEach(() => {
    http.verify();
    localStorage.clear();
  });

  it('starts logged out', () => {
    expect(service.isLoggedIn()).toBe(false);
    expect(service.user()).toBeNull();
  });

  it('setToken stores the JWT and loads the current user', async () => {
    const promise = service.setToken('jwt');
    http.expectOne('/api/users/me').flush(user);
    await promise;

    expect(localStorage.getItem(TOKEN_KEY)).toBe('jwt');
    expect(service.user()).toEqual(user);
    expect(service.isLoggedIn()).toBe(true);
  });

  it('setToken with a failed /users/me clears the session', async () => {
    const promise = service.setToken('jwt');
    http.expectOne('/api/users/me').flush({ error: 'unauthorized' }, { status: 401, statusText: 'Unauthorized' });
    await promise;

    expect(localStorage.getItem(TOKEN_KEY)).toBeNull();
    expect(service.user()).toBeNull();
    expect(service.isLoggedIn()).toBe(false);
  });

  it('caches /users/me for the session (single fetch)', async () => {
    const first = service.loadUser();
    http.expectOne('/api/users/me').flush(user);
    await first;

    const second = service.loadUser();
    await second;

    expect(service.user()).toEqual(user);
  });

  it('refreshUser bypasses the /users/me cache', async () => {
    const first = service.setToken('jwt');
    http.expectOne('/api/users/me').flush(user);
    await first;

    const updated: User = { ...user, ShowdownUsername: 'Renamed' };
    const second = service.refreshUser();
    http.expectOne('/api/users/me').flush(updated);
    await second;

    expect(service.user()).toEqual(updated);
  });

  it('needsOnboarding reflects a missing Showdown username', async () => {
    expect(service.needsOnboarding()).toBe(false);

    const promise = service.setToken('jwt');
    http.expectOne('/api/users/me').flush({ ...user, ShowdownUsername: null });
    await promise;
    expect(service.needsOnboarding()).toBe(true);

    const refreshed = service.refreshUser();
    http.expectOne('/api/users/me').flush(user);
    await refreshed;
    expect(service.needsOnboarding()).toBe(false);
  });

  it('prime resolves without a request when no token is stored', async () => {
    await service.prime();
    http.expectNone('/api/users/me');
  });

  it('prime restores the session when a token is stored', async () => {
    localStorage.setItem(TOKEN_KEY, 'jwt');
    const promise = service.prime();
    http.expectOne('/api/users/me').flush(user);
    await promise;

    expect(service.isLoggedIn()).toBe(true);
  });

  it('logout clears the token and user', async () => {
    const promise = service.setToken('jwt');
    http.expectOne('/api/users/me').flush(user);
    await promise;

    service.logout();
    http.expectOne('/auth/logout').flush({ message: 'ok' });

    expect(localStorage.getItem(TOKEN_KEY)).toBeNull();
    expect(service.user()).toBeNull();
  });

  it('login rejects with popup-blocked when the popup cannot open', async () => {
    vi.spyOn(window, 'open').mockReturnValue(null);

    await expect(service.login()).rejects.toBeInstanceOf(LoginError);
    await expect(service.login()).rejects.toMatchObject({ code: 'popup-blocked' });
  });

  it('login resolves when the popup posts a token back', async () => {
    const popup = { closed: false } as unknown as Window;
    vi.spyOn(window, 'open').mockReturnValue(popup);

    const promise = service.login();
    window.dispatchEvent(
      new MessageEvent('message', {
        data: { type: 'auth:success', token: 'jwt' },
        origin: window.location.origin,
        source: popup,
      }),
    );

    http.expectOne('/api/users/me').flush(user);
    await promise;

    expect(localStorage.getItem(TOKEN_KEY)).toBe('jwt');
    expect(service.user()).toEqual(user);
  });

  it('login rejects with popup-closed when the popup closes without completing', async () => {
    vi.useFakeTimers();
    const popup = { closed: false } as unknown as Window;
    vi.spyOn(window, 'open').mockReturnValue(popup);

    const promise = service.login();
    (popup as { closed: boolean }).closed = true;
    vi.advanceTimersByTime(400);

    await expect(promise).rejects.toMatchObject({ code: 'popup-closed' });
    vi.useRealTimers();
  });
});
