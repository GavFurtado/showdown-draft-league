import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { AuthService, TOKEN_KEY } from '../auth/auth-service';
import { User } from '../../shared/models/user.model';
import { DevUserSwitcher } from './dev-user-switcher';

// Members are `protected` per codebase convention; tests go through this narrow view.
interface SwitcherInternals {
  open(): boolean;
  toggle(): void;
  users(): User[];
  refresh(): Promise<void>;
  impersonate(userId: string): void;
  createUser(): Promise<void>;
  upsertMembership(): Promise<void>;
  newName: { set(value: string): void };
  leagueId: { set(value: string): void };
  role: { set(value: 'OWNER' | 'MODERATOR' | 'MEMBER'): void };
  selectedUserId: { set(value: string | null): void };
  error(): string | null;
}

const users: User[] = [
  {
    ID: 'user-1',
    DiscordID: '111',
    DiscordUsername: 'Alice',
    DiscordAvatarURL: '',
    ShowdownUsername: 'alice-shows',
    Role: 'user',
  },
  {
    ID: 'user-2',
    DiscordID: '222',
    DiscordUsername: 'Bob',
    DiscordAvatarURL: '',
    ShowdownUsername: null,
    Role: 'user',
  },
];

describe('DevUserSwitcher', () => {
  let c: SwitcherInternals;
  let http: HttpTestingController;

  const settle = () => new Promise((resolve) => setTimeout(resolve, 0));

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [DevUserSwitcher],
      providers: [provideHttpClientTesting()],
    });
    http = TestBed.inject(HttpTestingController);
    localStorage.clear();
    vi.stubGlobal('location', { reload: vi.fn() });

    const fixture = TestBed.createComponent(DevUserSwitcher);
    c = fixture.componentInstance as unknown as SwitcherInternals;
  });

  afterEach(() => {
    http.verify();
    localStorage.clear();
    vi.unstubAllGlobals();
  });

  it('is collapsed by default and fetches users when opened', async () => {
    expect(c.open()).toBe(false);

    c.toggle();
    http.expectOne('/auth/dev/users').flush(users);
    await settle();

    expect(c.open()).toBe(true);
    expect(c.users()).toEqual(users);
  });

  it('impersonate stores the minted JWT and reloads the app', () => {
    c.impersonate('user-2');

    const req = http.expectOne('/auth/dev/login');
    expect(req.request.body).toEqual({ UserId: 'user-2' });
    req.flush({ Token: 'jwt-for-bob' });

    expect(localStorage.getItem(TOKEN_KEY)).toBe('jwt-for-bob');
    expect(location.reload).toHaveBeenCalled();
  });

  it('ignores impersonating the already-active user', () => {
    TestBed.inject(AuthService).user.set(users[0]);

    c.impersonate(users[0].ID);

    http.expectNone('/auth/dev/login');
  });

  it('createUser posts a new fake user then refreshes the list', async () => {
    c.newName.set('Carol');
    const creation = c.createUser();

    const createReq = http.expectOne('/auth/dev/users');
    expect(createReq.request.method).toBe('POST');
    expect(createReq.request.body).toEqual({ Name: 'Carol' });
    createReq.flush(users[1], { status: 201, statusText: 'Created' });

    const listReq = http.expectOne('/auth/dev/users');
    expect(listReq.request.method).toBe('GET');
    listReq.flush(users);

    await creation;
    expect(c.users().length).toBe(2);
  });

  it('upsertMembership posts the selected user, league and role', async () => {
    c.selectedUserId.set('user-2');
    c.leagueId.set('league-abc');
    c.role.set('MODERATOR');
    const submission = c.upsertMembership();

    const req = http.expectOne('/auth/dev/memberships');
    expect(req.request.body).toEqual({ LeagueId: 'league-abc', UserId: 'user-2', Role: 'MODERATOR' });
    req.flush({});

    await submission;
    expect(c.error()).toBeNull();
  });

  it('surfaces an error when the backend is unreachable', async () => {
    const loading = c.refresh();
    http.expectOne('/auth/dev/users').flush({}, { status: 404, statusText: 'Not Found' });
    await loading;
    await settle();

    expect(c.error()).toContain('ENV=dev');
  });
});
