import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { errorInterceptor } from '../../core/error/error-interceptor';
import { asUuid } from '../../shared/types/branded-strings';
import { makeLeague } from '../../shared/testing/test-league';
import { UserLeaguesStore } from './user-leagues-store';

describe('UserLeaguesStore', () => {
  let http: HttpTestingController;
  let store: UserLeaguesStore;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(withInterceptors([errorInterceptor])), provideHttpClientTesting()],
    });
    http = TestBed.inject(HttpTestingController);
    store = TestBed.inject(UserLeaguesStore);
  });

  afterEach(() => http.verify());

  it('loads the authenticated user\'s leagues on creation', () => {
    http.expectOne('/api/users/me/leagues').flush([makeLeague(), makeLeague({ ID: asUuid('22222222-2222-4222-8222-222222222222') })]);

    expect(store.leagues().length).toBe(2);
    expect(store.loading()).toBe(false);
    expect(store.error()).toBeNull();
  });

  it('reloads on refetch', () => {
    http.expectOne('/api/users/me/leagues').flush([makeLeague()]);

    store.refetch();

    http.expectOne('/api/users/me/leagues').flush([makeLeague({ ID: asUuid('22222222-2222-4222-8222-222222222222') })]);
    expect(store.leagues()[0].ID).toBe('22222222-2222-4222-8222-222222222222');
  });

  it('surfaces load errors', () => {
    http
      .expectOne('/api/users/me/leagues')
      .flush({ Message: 'nope' }, { status: 500, statusText: 'Server Error' });

    expect(store.leagues()).toEqual([]);
    expect(store.loading()).toBe(false);
    expect(store.error()?.status).toBe(500);
  });
});