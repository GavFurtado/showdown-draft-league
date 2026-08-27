import { provideHttpClientTesting } from '@angular/common/http/testing';
import { ActivatedRouteSnapshot, RouterStateSnapshot, UrlTree, provideRouter } from '@angular/router';
import { TestBed } from '@angular/core/testing';

import { User } from '../../shared/models/user.model';
import { AuthService } from './auth-service';
import { authGuard, onboardingGuard } from './auth-guard';

const user: User = {
  ID: 'user-1',
  DiscordID: '1234',
  DiscordUsername: 'Gavin',
  DiscordAvatarURL: '',
  ShowdownUsername: 'GavinTest',
  Role: 'user',
};

describe('authGuard', () => {
  const executeGuard: (route?: ActivatedRouteSnapshot, state?: RouterStateSnapshot) => unknown = (route, state) =>
    TestBed.runInInjectionContext(() =>
      authGuard(route ?? ({} as ActivatedRouteSnapshot), state ?? ({} as RouterStateSnapshot)),
    );

  let auth: AuthService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideRouter([]), provideHttpClientTesting()],
    });
    auth = TestBed.inject(AuthService);
  });

  it('allows access when logged in', () => {
    auth.user.set(user);

    expect(executeGuard()).toBe(true);
  });

  it('redirects to /login when logged out', () => {
    const result = executeGuard();

    expect((result as UrlTree).toString()).toBe('/login');
  });
});

describe('onboardingGuard', () => {
  const executeGuard: () => unknown = () => TestBed.runInInjectionContext(() => onboardingGuard({} as ActivatedRouteSnapshot, {} as RouterStateSnapshot));

  let auth: AuthService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideRouter([]), provideHttpClientTesting()],
    });
    auth = TestBed.inject(AuthService);
  });

  it('redirects to /login when logged out', () => {
    const result = executeGuard();

    expect((result as UrlTree).toString()).toBe('/login');
  });

  it('redirects to /onboarding when the user has no Showdown username yet', () => {
    auth.user.set({ ...user, ShowdownUsername: null });

    const result = executeGuard();

    expect((result as UrlTree).toString()).toBe('/onboarding');
  });

  it('allows access once the user finished onboarding', () => {
    auth.user.set(user);

    expect(executeGuard()).toBe(true);
  });
});
