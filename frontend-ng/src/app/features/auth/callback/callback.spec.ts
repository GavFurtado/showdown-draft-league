import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { ActivatedRoute, Router, convertToParamMap, provideRouter } from '@angular/router';
import { provideTaiga } from '@taiga-ui/core';

import { User } from '../../../shared/models/user.model';
import { UserRole } from '../../../shared/models/enums/user-role';
import { asUuid } from '../../../shared/types/branded-strings';
import { TOKEN_KEY } from '../../../core/auth/auth-service';
import { Callback } from './callback';

const user: User = {
  ID: asUuid('33333333-3333-4333-8333-333333333333'),
  DiscordID: '1234',
  DiscordUsername: 'Tester',
  DiscordAvatarURL: '',
  ShowdownUsername: 'tester_show',
  Role: UserRole.USER,
};

describe('Callback', () => {
  let http: HttpTestingController;

  function setup(query: Record<string, string>): void {
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      imports: [Callback],
      providers: [
        provideHttpClientTesting(),
        provideRouter([]),
        provideTaiga(),
        { provide: ActivatedRoute, useValue: { snapshot: { queryParamMap: convertToParamMap(query) } } },
      ],
    });
    http = TestBed.inject(HttpTestingController);
    vi.spyOn(TestBed.inject(Router), 'navigateByUrl').mockResolvedValue(true);
  }

  // Lets the component's promise chain (firstValueFrom → setToken → loadUser → finish) settle.
  function tick(): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, 0));
  }

  afterEach(() => {
    http.verify();
    localStorage.clear();
  });

  it('self-fetches the backend callback and completes login', async () => {
    setup({ code: 'code', state: 'state' });
    TestBed.createComponent(Callback);

    const callbackReq = http.expectOne((req) => req.url === '/auth/discord/callback');
    expect(callbackReq.request.params.get('code')).toBe('code');
    expect(callbackReq.request.params.get('state')).toBe('state');
    callbackReq.flush({ Token: 'jwt' });
    await tick();

    http.expectOne('/api/users/me').flush(user);
    await tick();

    expect(TestBed.inject(Router).navigateByUrl).toHaveBeenCalledWith('/my-leagues');
    expect(localStorage.getItem(TOKEN_KEY)).toBe('jwt');
  });

  it('navigates to /login when code or state are missing', async () => {
    setup({});
    TestBed.createComponent(Callback);
    await tick();

    expect(TestBed.inject(Router).navigateByUrl).toHaveBeenCalledWith('/login');
  });

  it('navigates to /login when the backend callback fails', async () => {
    setup({ code: 'code', state: 'state' });
    TestBed.createComponent(Callback);

    http
      .expectOne((req) => req.url === '/auth/discord/callback')
      .flush({ error: 'bad' }, { status: 500, statusText: 'Internal Server Error' });
    await tick();

    expect(TestBed.inject(Router).navigateByUrl).toHaveBeenCalledWith('/login');
    expect(localStorage.getItem(TOKEN_KEY)).toBeNull();
  });
});
