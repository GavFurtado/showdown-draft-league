import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ActivatedRoute, convertToParamMap, provideRouter } from '@angular/router';
import { provideTaiga } from '@taiga-ui/core';

import { errorInterceptor } from '../../core/error/error-interceptor';
import { AuthService } from '../../core/auth/auth-service';
import { TEST_USER, makeLeague } from '../../shared/testing/test-league';
import { LeagueStatus } from '../league/models/enums/league-status';
import { JoinLeagueStore } from './join-league-store';
import { JoinLeague } from './join-league';

describe('JoinLeague', () => {
  let fixture: ComponentFixture<JoinLeague>;
  let http: HttpTestingController;
  let store: JoinLeagueStore;

  function setup(): void {
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      imports: [JoinLeague],
      providers: [
        provideRouter([]),
        provideTaiga(),
        provideHttpClient(withInterceptors([errorInterceptor])),
        provideHttpClientTesting(),
        JoinLeagueStore,
        {
          provide: ActivatedRoute,
          useValue: {
            // readLeagueId() walks the snapshot chain (snapshot.parent), not the route chain.
            snapshot: {
              paramMap: convertToParamMap({}),
              parent: { paramMap: convertToParamMap({ leagueId: 'league-1' }), parent: null },
            },
          },
        },
      ],
    }).compileComponents();

    http = TestBed.inject(HttpTestingController);
    store = TestBed.inject(JoinLeagueStore);
    TestBed.inject(AuthService).user.set(TEST_USER);
    fixture = TestBed.createComponent(JoinLeague);
  }

  function flushLeague(overrides: Parameters<typeof makeLeague>[0] = {}): void {
    http.expectOne('/api/leagues/league-1').flush(makeLeague(overrides));
  }

  afterEach(() => http.verify());

  it('renders league info and the join form', async () => {
    setup();
    fixture.detectChanges(); // triggers ngOnInit → load

    flushLeague();
    await fixture.whenStable();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Test League');
    expect(text).toContain('Be nice');
    expect(fixture.nativeElement.querySelector('form')).toBeTruthy();
  });

  it('shows the already-a-member state instead of the form', async () => {
    setup();
    fixture.detectChanges();

    flushLeague({
      Members: [{ ID: 'm1', LeagueID: 'league-1', UserID: 'user-1', DisplayName: 'Gavin', Role: 'MEMBER', IsActive: true, JoinedAt: '' }],
    });
    await fixture.whenStable();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain("already a member");
    expect(fixture.nativeElement.querySelector('form')).toBeNull();
  });

  it('blocks joining for leagues outside SETUP', async () => {
    setup();
    fixture.detectChanges();

    flushLeague({ Status: LeagueStatus.REGULAR_SEASON });
    await fixture.whenStable();
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain("isn't accepting new members");
    expect(fixture.nativeElement.querySelector('form')).toBeNull();
  });

  it('submits with trimmed names and shows the success state', async () => {
    setup();
    fixture.detectChanges();

    flushLeague();
    await fixture.whenStable();
    fixture.detectChanges();

    const el = fixture.nativeElement;
    el.querySelector('input[formcontrolname="InLeagueName"]').value = '  Coach  ';
    el.querySelector('input[formcontrolname="InLeagueName"]')
      .dispatchEvent(new Event('input', { bubbles: true }));
    el.querySelector('input[formcontrolname="TeamName"]').value = '';
    el.querySelector('input[formcontrolname="TeamName"]')
      .dispatchEvent(new Event('input', { bubbles: true }));

    el.querySelector('button[type="submit"]').click();
    await fixture.whenStable();

    const req = http.expectOne('/api/leagues/league-1/members/join');
    expect(req.request.body).toEqual({
      UserID: 'user-1',
      LeagueID: 'league-1',
      InLeagueName: 'Coach',
      TeamName: undefined,
    });
    req.flush({ ID: 'member-9', InLeagueName: 'Coach', TeamName: "Gavin's Team" });
    await fixture.whenStable();
    fixture.detectChanges();

    expect(el.textContent).toContain("You're in!");
    expect(store.joinedMember()?.ID).toBe('member-9');
  });

  it('shows server errors on failed joins (409 name taken)', async () => {
    setup();
    fixture.detectChanges();

    flushLeague();
    await fixture.whenStable();
    fixture.detectChanges();

    const el = fixture.nativeElement;
    el.querySelector('button[type="submit"]').click();
    await fixture.whenStable();

    http
      .expectOne('/api/leagues/league-1/members/join')
      .flush({ Message: "team name is taken" }, { status: 409, statusText: 'Conflict' });
    await fixture.whenStable();
    fixture.detectChanges();

    expect(el.textContent).toContain('team name is taken');
    expect(el.querySelector('form')).toBeTruthy();
  });
});
