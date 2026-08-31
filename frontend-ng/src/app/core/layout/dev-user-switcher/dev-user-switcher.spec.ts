import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { User } from '../../../shared/models/user.model';
import { UserRole } from '../../../shared/models/enums/user-role';
import { MemberRole } from '../../../features/league/models/enums/member-role';
import { asUuid } from '../../../shared/types/branded-strings';
import { AuthService, TOKEN_KEY } from '../../auth/auth-service';
import { DevUserSwitcher } from './dev-user-switcher';

const users: User[] = [
  {
    ID: asUuid('33333333-3333-4333-8333-333333333333'),
    DiscordID: '111',
    DiscordUsername: 'Alice',
    DiscordAvatarURL: '',
    ShowdownUsername: 'alice_shows',
    Role: UserRole.USER,
  },
  {
    ID: asUuid('44444444-4444-4444-8444-444444444444'),
    DiscordID: '222',
    DiscordUsername: 'Bob',
    DiscordAvatarURL: '',
    ShowdownUsername: null,
    Role: UserRole.USER,
  },
];

describe('DevUserSwitcher', () => {
  let component: DevUserSwitcher;
  let auth: AuthService;
  let http: HttpTestingController;
  const settle = () => new Promise((resolve) => setTimeout(resolve, 0));

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [DevUserSwitcher],
      providers: [provideHttpClientTesting()],
    });
    component = TestBed.createComponent(DevUserSwitcher).componentInstance;
    auth = TestBed.inject(AuthService);
    http = TestBed.inject(HttpTestingController);
    localStorage.clear();
    vi.stubGlobal('location', { reload: vi.fn() });
  });

  afterEach(() => {
    http.verify();
    localStorage.clear();
    vi.unstubAllGlobals();
  });

  it('refresh loads the user list', async () => {
    const req = component.refresh();
    http.expectOne('/auth/dev/users').flush(users);
    await req;

    expect(component.users()).toEqual(users);
  });

  it('refresh pre-selects the current session user', async () => {
    auth.user.set(users[0]);
    const req = component.refresh();
    http.expectOne('/auth/dev/users').flush(users);
    await req;

    expect(component.selectedUserId()).toBe(users[0].ID);
    expect(component.roleTargetUserId()).toBe(users[0].ID);
  });

  it('selecting another user stores the minted JWT and reloads the app', async () => {
    component.selectUser('44444444-4444-4444-8444-444444444444');
    const req = http.expectOne('/auth/dev/login');
    expect(req.request.body).toEqual({ UserId: '44444444-4444-4444-8444-444444444444' });
    req.flush({ Token: 'jwt-for-bob' });
    await settle();

    expect(localStorage.getItem(TOKEN_KEY)).toBe('jwt-for-bob');
    expect(location.reload).toHaveBeenCalled();
  });

  it('selecting the already-active user does not re-impersonate', async () => {
    auth.user.set(users[0]);
    await component.selectUser(users[0].ID);
    http.expectNone('/auth/dev/login');
  });

  it('upsertMembership posts the picked user, league and role', async () => {
    component.roleTargetUserId.set('44444444-4444-4444-8444-444444444444');
    component.leagueId.set('league-abc');
    component.role.set(MemberRole.MODERATOR);
    const submission = component.upsertMembership();

    const req = http.expectOne('/auth/dev/memberships');
    expect(req.request.body).toEqual({ LeagueId: 'league-abc', UserId: '44444444-4444-4444-8444-444444444444', Role: MemberRole.MODERATOR });
    req.flush({});
    await submission;
    expect(component.error()).toBeNull();
  });

  it('upsertMembership sets a success message on completion', async () => {
    component.users.set(users);
    component.roleTargetUserId.set(users[0].ID);
    component.leagueId.set('league-abc');
    const submission = component.upsertMembership();

    const req = http.expectOne('/auth/dev/memberships');
    req.flush({});
    await submission;

    expect(component.success()).toContain(users[0].DiscordUsername);
    expect(component.error()).toBeNull();
  });

  it('upsertMembership defaults to the current user when no role target is picked', async () => {
    auth.user.set(users[0]);
    component.leagueId.set('league-abc');
    const submission = component.upsertMembership();

    const req = http.expectOne('/auth/dev/memberships');
    expect(req.request.body).toEqual({ LeagueId: 'league-abc', UserId: users[0].ID, Role: MemberRole.MEMBER });
    req.flush({});
    await submission;
    expect(component.error()).toBeNull();
  });
});
