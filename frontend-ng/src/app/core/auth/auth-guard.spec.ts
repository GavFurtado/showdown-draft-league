import { provideHttpClientTesting } from '@angular/common/http/testing';
import { ActivatedRouteSnapshot, RouterStateSnapshot, UrlTree, provideRouter } from '@angular/router';
import { TestBed } from '@angular/core/testing';

import { User } from '../../shared/models/user.model';
import { UserRole } from '../../shared/models/enums/user-role';
import { asUuid } from '../../shared/types/branded-strings';
import { AuthService } from './auth-service';
import { authGuard } from './auth-guard';

const user: User = {
  ID: asUuid('33333333-3333-4333-8333-333333333333'),
  DiscordID: '1234',
  DiscordUsername: 'Tester',
  DiscordAvatarURL: '',
  ShowdownUsername: 'tester_show',
  Role: UserRole.USER,
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
