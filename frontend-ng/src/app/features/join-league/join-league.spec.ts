import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Component } from '@angular/core';
import { ActivatedRoute, convertToParamMap, provideRouter } from '@angular/router';
import { TuiRoot, provideTaiga } from '@taiga-ui/core';

import { errorInterceptor } from '../../core/error/error-interceptor';
import { AuthService } from '../../core/auth/auth-service';
import { asIsoDateTime, asUuid } from '../../shared/types/branded-strings';
import { TEST_USER, makeLeague } from '../../shared/testing/test-league';
import { LeagueStatus } from '../league/models/enums/league-status';
import { MemberRole } from '../league/models/enums/member-role';
import { JoinLeagueStore } from './join-league-store';
import { JoinLeague } from './join-league';

// Toast/notification portals render into <tui-root>, so host the component under one.
@Component({
  selector: 'app-host',
  imports: [TuiRoot, JoinLeague],
  template: `<tui-root><app-join-league /></tui-root>`,
})
class Host {}

describe('JoinLeague', () => {
  let fixture: ComponentFixture<Host>;
  let http: HttpTestingController;
  let store: JoinLeagueStore;

  function setup(): void {
    TestBed.resetTestingModule();
    TestBed.configureTestingModule({
      imports: [Host],
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
              parent: {
                paramMap: convertToParamMap({ leagueId: '11111111-1111-4111-8111-111111111111' }),
                parent: null,
              },
            },
          },
        },
      ],
    }).compileComponents();

    http = TestBed.inject(HttpTestingController);
    store = TestBed.inject(JoinLeagueStore);
    TestBed.inject(AuthService).user.set(TEST_USER);
    fixture = TestBed.createComponent(Host);
  }

  function flushLeague(overrides: Parameters<typeof makeLeague>[0] = {}): void {
    http.expectOne('/api/leagues/11111111-1111-4111-8111-111111111111').flush(makeLeague(overrides));
  }

  // The walkthrough starts on the league-details (accordion) step; advance to the form.
  function goToForm(): void {
    const buttons = Array.from<HTMLElement>(fixture.nativeElement.querySelectorAll('button'));
    const continueBtn = buttons.find((b) => b.textContent?.trim() === 'Continue');
    (continueBtn as HTMLButtonElement).click();
    fixture.detectChanges();
  }

  afterEach(() => http.verify());

  it('renders league info as accordions, then continues to the join form', async () => {
    setup();
    fixture.detectChanges(); // triggers ngOnInit → load

    flushLeague();
    await fixture.whenStable();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Test League');
    expect(text).toContain('Overview');
    expect(text).toContain('Players');
    expect(text).toContain('Season & Format');
    expect(fixture.nativeElement.querySelector('form')).toBeNull();

    goToForm();
    expect(fixture.nativeElement.querySelector('form')).toBeTruthy();
    expect(fixture.nativeElement.textContent).toContain('In-league name');
  });

  it('shows the joined-landing state instead of the form', async () => {
    setup();
    fixture.detectChanges();

    flushLeague({
      Members: [
        {
          ID: asUuid('66666666-6666-4666-8666-666666666666'),
          LeagueID: asUuid('11111111-1111-4111-8111-111111111111'),
          UserID: asUuid('33333333-3333-4333-8333-333333333333'),
          InLeagueName: 'Tester',
          Role: MemberRole.MEMBER,
          IsActive: true,
          JoinedAt: asIsoDateTime('2026-08-30T12:00:00.000Z'),
        },
      ],
    });
    await fixture.whenStable();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain("You're already a member of");
    expect(text).toContain('Test League');
    expect(text).toContain("Redirecting to the league's dashboard");
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

  it('blocks joining while the league is pending', async () => {
    setup();
    fixture.detectChanges();

    flushLeague({ Status: LeagueStatus.PENDING });
    await fixture.whenStable();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain("isn't accepting new members");
    expect(fixture.nativeElement.querySelector('form')).toBeNull();
  });

  it('allows joining while the league is in setup', async () => {
    setup();
    fixture.detectChanges();

    flushLeague({ Status: LeagueStatus.SETUP });
    await fixture.whenStable();
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).not.toContain("isn't accepting new members");
    goToForm();
    expect(fixture.nativeElement.querySelector('form')).toBeTruthy();
  });

  it('submits with trimmed names and shows the success state', async () => {
    setup();
    fixture.detectChanges();

    flushLeague();
    await fixture.whenStable();
    fixture.detectChanges();

    const el = fixture.nativeElement;
    goToForm();
    el.querySelector('input[formcontrolname="InLeagueName"]').value = '  Coach  ';
    el.querySelector('input[formcontrolname="InLeagueName"]').dispatchEvent(new Event('input', { bubbles: true }));
    el.querySelector('input[formcontrolname="TeamName"]').value = '';
    el.querySelector('input[formcontrolname="TeamName"]').dispatchEvent(new Event('input', { bubbles: true }));

    el.querySelector('button[type="submit"]').click();
    await fixture.whenStable();

    const req = http.expectOne('/api/leagues/11111111-1111-4111-8111-111111111111/members/join');
    expect(req.request.body).toEqual({
      UserID: '33333333-3333-4333-8333-333333333333',
      LeagueID: '11111111-1111-4111-8111-111111111111',
      InLeagueName: 'Coach',
      TeamName: undefined,
    });
    req.flush({ ID: '77777777-7777-4777-8777-777777777777', InLeagueName: 'Coach', TeamName: "Tester's Team" });
    await fixture.whenStable();
    fixture.detectChanges();

    expect(el.textContent).toContain('You have successfully joined');
    expect(el.textContent).toContain('Test League');
    expect(store.joinedMember()?.ID).toBe('77777777-7777-4777-8777-777777777777');
  });

  it('shows server errors on failed joins (409 name taken)', async () => {
    setup();
    fixture.detectChanges();

    flushLeague();
    await fixture.whenStable();
    fixture.detectChanges();

    const el = fixture.nativeElement;
    goToForm();
    el.querySelector('button[type="submit"]').click();
    await fixture.whenStable();

    http
      .expectOne('/api/leagues/11111111-1111-4111-8111-111111111111/members/join')
      .flush({ Message: 'team name is taken' }, { status: 409, statusText: 'Conflict' });
    await fixture.whenStable();
    fixture.detectChanges();

    expect(el.textContent).toContain('team name is taken');
    expect(el.querySelector('form')).toBeTruthy();
  });
});
