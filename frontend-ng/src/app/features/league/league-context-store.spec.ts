import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { errorInterceptor } from '../../core/error/error-interceptor';
import { AuthService } from '../../core/auth/auth-service';
import { asIsoDateTime, asUuid } from '../../shared/types/branded-strings';
import { MemberRole } from './models/enums/member-role';
import { LeagueMember } from './models/league-member.model';
import { LeagueContextStore } from './league-context-store';
import { makeLeague } from '../../shared/testing/test-league';
import { TEST_USER } from '../../shared/testing/test-league';

describe('LeagueContextStore', () => {
  let http: HttpTestingController;
  let store: LeagueContextStore;
  let auth: AuthService;

  const leagueId = asUuid('11111111-1111-4111-8111-111111111111');

  const settle = () => new Promise<void>((resolve) => setTimeout(resolve, 0));

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(withInterceptors([errorInterceptor])), provideHttpClientTesting()],
    });
    http = TestBed.inject(HttpTestingController);
    store = TestBed.inject(LeagueContextStore);
    auth = TestBed.inject(AuthService);
    auth.user.set(TEST_USER);
  });

  afterEach(() => http.verify());

  it('should be created', () => {
    expect(store).toBeTruthy();
  });

  const member: LeagueMember = {
    ID: asUuid('66666666-6666-4666-8666-666666666666'),
    LeagueID: leagueId,
    UserID: TEST_USER.ID,
    InLeagueName: 'Tester',
    Role: MemberRole.MEMBER,
    IsActive: true,
    JoinedAt: asIsoDateTime('2026-08-30T12:00:00.000Z'),
  };

  it('refresh loads league, user member (fetched), and current draft', async () => {
    const league = makeLeague({ Members: [] });

    const pending = store.refresh(leagueId);
    // league + members list fetch in parallel
    http.expectOne(`/api/leagues/${leagueId}`).flush(league);
    http.expectOne(`/api/leagues/${leagueId}/members`).flush([member]);
    await settle();
    http.expectOne(`/api/leagues/${leagueId}/draft`).flush({ ID: asUuid('55555555-5555-4555-8555-555555555555') });
    await pending;

    expect(store.league()).toBe(league);
    expect(store.userMember()?.UserID).toBe(TEST_USER.ID);
    expect(store.currentDraft()?.ID).toBe(asUuid('55555555-5555-4555-8555-555555555555'));
  });

  it('refresh leaves the user member null when the user is not a member', async () => {
    const pending = store.refresh(leagueId);
    http.expectOne(`/api/leagues/${leagueId}`).flush(makeLeague({ Members: [] }));
    http.expectOne(`/api/leagues/${leagueId}/members`).flush([]);
    await settle();
    http.expectOne(`/api/leagues/${leagueId}/draft`).flush({});
    await pending;

    expect(store.userMember()).toBeNull();
    expect(store.isOwner()).toBe(false);
    expect(store.isModerator()).toBe(false);
  });

  it('refresh exposes owner and moderator roles from the fetched member', async () => {
    const pending = store.refresh(leagueId);
    http.expectOne(`/api/leagues/${leagueId}`).flush(makeLeague());
    http
      .expectOne(`/api/leagues/${leagueId}/members`)
      .flush([{ ...member, Role: MemberRole.OWNER }]);
    await settle();
    http.expectOne(`/api/leagues/${leagueId}/draft`).flush({});
    await pending;

    expect(store.isOwner()).toBe(true);
    expect(store.isModerator()).toBe(true);
  });

  it('refresh clears the draft when none exists (404)', async () => {
    store.currentDraft.set({ ID: asUuid('55555555-5555-4555-8555-555555555555') } as never);

    const pending = store.refresh(leagueId);
    http.expectOne(`/api/leagues/${leagueId}`).flush(makeLeague());
    http.expectOne(`/api/leagues/${leagueId}/members`).flush([]);
    await settle();
    http
      .expectOne(`/api/leagues/${leagueId}/draft`)
      .flush({ Message: 'draft not found' }, { status: 404, statusText: 'Not Found' });
    await pending;

    expect(store.currentDraft()).toBeNull();
  });

  it('ensureLoaded cold-blocks once, then resolves immediately with membership', async () => {
    const first = store.ensureLoaded(leagueId);
    http.expectOne(`/api/leagues/${leagueId}`).flush(makeLeague());
    http.expectOne(`/api/leagues/${leagueId}/members`).flush([member]);
    await settle();
    http.expectOne(`/api/leagues/${leagueId}/draft`).flush({});
    await expect(first).resolves.toBe(true);

    // Subsequent navigations are non-blocking — no new network requests.
    await expect(store.ensureLoaded(leagueId)).resolves.toBe(true);
  });
});
