import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { errorInterceptor } from '../../core/error/error-interceptor';
import { makeLeague } from '../../shared/testing/test-league';
import { JoinLeagueStore } from './join-league-store';

describe('JoinLeagueStore', () => {
  let http: HttpTestingController;
  let store: JoinLeagueStore;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(withInterceptors([errorInterceptor])), provideHttpClientTesting()],
    });
    http = TestBed.inject(HttpTestingController);
    store = TestBed.inject(JoinLeagueStore);
  });

  afterEach(() => http.verify());

  it('loads a league by id', () => {
    store.load('league-1');

    http.expectOne('/api/leagues/league-1').flush(makeLeague());

    expect(store.league()?.Name).toBe('Test League');
    expect(store.loading()).toBe(false);
    expect(store.loadError()).toBeNull();
  });

  it('surfaces load errors', () => {
    store.load('league-1');

    http
      .expectOne('/api/leagues/league-1')
      .flush({ Message: 'league not found' }, { status: 404, statusText: 'Not Found' });

    expect(store.league()).toBeNull();
    expect(store.loadError()?.message).toBe('league not found');
    expect(store.loading()).toBe(false);
  });

  it('joins and exposes the created member', () => {
    store.join({ UserID: 'user-1', LeagueID: 'league-1', InLeagueName: 'Coach', TeamName: 'Team Rocket' });

    const req = http.expectOne('/api/leagues/league-1/members/join');
    expect(req.request.body).toEqual({
      UserID: 'user-1',
      LeagueID: 'league-1',
      InLeagueName: 'Coach',
      TeamName: 'Team Rocket',
    });
    req.flush({ ID: 'member-1', UserID: 'user-1', InLeagueName: 'Coach', TeamName: 'Team Rocket' });

    expect(store.joinedMember()?.ID).toBe('member-1');
    expect(store.joining()).toBe(false);
    expect(store.joinError()).toBeNull();
  });

  it('omits blank names so the server applies its defaults', () => {
    store.join({ UserID: 'user-1', LeagueID: 'league-1' });

    const req = http.expectOne('/api/leagues/league-1/members/join');
    expect(req.request.body).toEqual({ UserID: 'user-1', LeagueID: 'league-1' });
    req.flush({ ID: 'member-1' });
  });

  it('surfaces join errors (e.g. name taken → 409)', () => {
    store.join({ UserID: 'user-1', LeagueID: 'league-1' });

    http
      .expectOne('/api/leagues/league-1/members/join')
      .flush({ Message: "team name is taken" }, { status: 409, statusText: 'Conflict' });

    expect(store.joinedMember()).toBeNull();
    expect(store.joinError()?.status).toBe(409);
    expect(store.joining()).toBe(false);
  });
});
