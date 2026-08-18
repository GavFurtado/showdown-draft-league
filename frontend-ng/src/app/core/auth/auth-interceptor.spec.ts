import { HttpClient, provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';

import { TOKEN_KEY } from './auth-service';
import { authInterceptor } from './auth-interceptor';

describe('authInterceptor', () => {
  let http: HttpClient;
  let testing: HttpTestingController;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(withInterceptors([authInterceptor])), provideHttpClientTesting()],
    });
    http = TestBed.inject(HttpClient);
    testing = TestBed.inject(HttpTestingController);
    localStorage.clear();
  });

  afterEach(() => {
    testing.verify();
    localStorage.clear();
  });

  it('attaches the Bearer token to /api/* requests', () => {
    localStorage.setItem(TOKEN_KEY, 'jwt');
    http.get('/api/users/me').subscribe();

    const req = testing.expectOne('/api/users/me');
    expect(req.request.headers.get('Authorization')).toBe('Bearer jwt');
  });

  it('does not attach a token to /auth/* requests', () => {
    localStorage.setItem(TOKEN_KEY, 'jwt');
    http.get('/auth/discord/login').subscribe();

    const req = testing.expectOne('/auth/discord/login');
    expect(req.request.headers.has('Authorization')).toBe(false);
  });

  it('does not attach a token when none is stored', () => {
    http.get('/api/users/me').subscribe();

    const req = testing.expectOne('/api/users/me');
    expect(req.request.headers.has('Authorization')).toBe(false);
  });
});
