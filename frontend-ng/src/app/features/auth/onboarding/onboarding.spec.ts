import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { Router, provideRouter } from '@angular/router';
import { TestBed } from '@angular/core/testing';

import { AuthService } from '../../../core/auth/auth-service';
import { User } from '../../../shared/models/user.model';
import { Onboarding } from './onboarding';

// Members are `protected` per codebase convention; tests go through this narrow view.
interface OnboardingInternals {
  username: { set(value: string): void };
  submit(): Promise<void>;
  error(): string | null;
}

const onboarded: User = {
  ID: 'user-1',
  DiscordID: '1234',
  DiscordUsername: 'Gavin',
  DiscordAvatarURL: '',
  ShowdownUsername: 'GavinTest',
  Role: 'user',
};

const freshSignup: User = { ...onboarded, ShowdownUsername: null };

describe('Onboarding', () => {
  let http: HttpTestingController;
  let router: Router;
  let auth: AuthService;
  let navigate: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [Onboarding],
      providers: [provideRouter([]), provideHttpClientTesting()],
    });
    http = TestBed.inject(HttpTestingController);
    router = TestBed.inject(Router);
    auth = TestBed.inject(AuthService);
    navigate = vi.spyOn(router, 'navigateByUrl').mockResolvedValue(true);
  });

  afterEach(() => {
    http.verify();
    vi.restoreAllMocks();
  });

  function create(): OnboardingInternals {
    const fixture = TestBed.createComponent(Onboarding);
    return fixture.componentInstance as unknown as OnboardingInternals;
  }

  it('bounces fully-onboarded users away immediately', () => {
    auth.user.set(onboarded);

    create();

    expect(navigate).toHaveBeenCalledWith('/my-leagues');
  });

  it('saves the showdown username and moves on', async () => {
    auth.user.set(freshSignup);
    const c = create();

    c.username.set('gavin_shows');
    const submission = c.submit();

    const put = http.expectOne('/api/users/profile');
    expect(put.request.method).toBe('PUT');
    expect(put.request.body).toEqual({ ShowdownName: 'gavin_shows' });
    put.flush({ ...freshSignup, ShowdownUsername: 'gavin_shows' });

    const me = http.expectOne('/api/users/me');
    me.flush({ ...freshSignup, ShowdownUsername: 'gavin_shows' });

    await submission;

    expect(auth.needsOnboarding()).toBe(false);
    expect(navigate).toHaveBeenCalledWith('/my-leagues');
  });

  it('stays on the form when the server rejects the name', async () => {
    auth.user.set(freshSignup);
    const c = create();

    c.username.set('taken_name');
    const submission = c.submit();

    http.expectOne('/api/users/profile').flush({}, { status: 500, statusText: 'Server Error' });

    await submission;

    expect(c.error()).toBeTruthy();
    expect(navigate).not.toHaveBeenCalledWith('/my-leagues');
  });
});
