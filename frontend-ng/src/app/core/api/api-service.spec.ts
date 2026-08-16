import { provideHttpClient, withInterceptors } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import type { Observable } from 'rxjs';

import { ApiService } from './api-service';
import { ApiErrorResponse } from './api.model';
import { errorInterceptor } from '../error/error-interceptor';
import { ErrorService } from '../error/error-service';

describe('ApiService.get', () => {
  let service: ApiService;
  let controller: HttpTestingController;
  let errors: ErrorService;

  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [provideHttpClient(withInterceptors([errorInterceptor])), provideHttpClientTesting()],
    });
    service = TestBed.inject(ApiService);
    controller = TestBed.inject(HttpTestingController);
    errors = TestBed.inject(ErrorService);
  });

  afterEach(() => {
    controller.verify();
  });

  function expectUrl(url: string) {
    return controller.expectOne((req) => req.url === url);
  }

  function once<T>(source: Observable<T>): Promise<{ next?: T; error?: unknown }> {
    return new Promise((resolve) => {
      source.subscribe({
        next: (value) => resolve({ next: value }),
        error: (err) => resolve({ error: err }),
      });
    });
  }

  function serverError(message: string): ApiErrorResponse {
    return {
      Timestamp: '2024-01-01T12:00:00Z',
      Status: 500,
      Error: 'Internal Server Error',
      Message: message,
      Path: '/api/things',
    };
  }

  describe('happy path', () => {
    it('requests baseUrl + path and emits the payload', async () => {
      const payload = { ID: 'league-1' };
      const result = once(service.get<{ ID: string }>('/api/leagues/league-1'));

      const req = expectUrl('/api/leagues/league-1');
      req.flush(payload);

      await expect(result).resolves.toEqual({ next: payload });
    });

    it('dedupes concurrent gets of the same path into a single request', async () => {
      const payload = { id: 1 };
      const first = once(service.get<{ id: number }>('/api/things'));
      const second = once(service.get<{ id: number }>('/api/things'));

      const req = expectUrl('/api/things');
      req.flush(payload);

      await expect(first).resolves.toEqual({ next: payload });
      await expect(second).resolves.toEqual({ next: payload });
    });

    it('does not dedupe different paths', async () => {
      const first = once(service.get('/api/a'));
      const second = once(service.get('/api/b'));

      const reqA = expectUrl('/api/a');
      const reqB = expectUrl('/api/b');
      reqA.flush({ id: 1 });
      reqB.flush({ id: 2 });

      await expect(first).resolves.toEqual({ next: { id: 1 } });
      await expect(second).resolves.toEqual({ next: { id: 2 } });
    });

    it('does not dedupe the same path with different params', async () => {
      const first = once(service.get('/api/things', { params: { limit: 1 } }));
      const second = once(service.get('/api/things', { params: { limit: 2 } }));

      const reqA = controller.expectOne((req) => req.url === '/api/things' && req.params.get('limit') === '1');
      const reqB = controller.expectOne((req) => req.url === '/api/things' && req.params.get('limit') === '2');
      expect(reqA.request.params.get('limit')).toBe('1');
      expect(reqB.request.params.get('limit')).toBe('2');
      reqA.flush([1]);
      reqB.flush([2]);

      await expect(first).resolves.toEqual({ next: [1] });
      await expect(second).resolves.toEqual({ next: [2] });
    });

    it('serializes params and drops null/undefined values', async () => {
      const result = once(
        service.get('/api/search', { params: { q: 'pikachu', cost: 0, skip: null, extra: undefined } }),
      );

      const req = expectUrl('/api/search');
      expect(req.request.params.get('q')).toBe('pikachu');
      expect(req.request.params.get('cost')).toBe('0');
      expect(req.request.params.has('skip')).toBe(false);
      expect(req.request.params.has('extra')).toBe(false);
      req.flush([]);

      await expect(result).resolves.toEqual({ next: [] });
    });
  });

  describe('unhappy path', () => {
    it('normalizes a server error into a ClientError, reports it, and rethrows', async () => {
      const result = once(service.get('/api/things'));

      const req = expectUrl('/api/things');
      req.flush(serverError('boom'), {
        status: 500,
        statusText: 'Internal Server Error',
        // eslint-disable-next-line @typescript-eslint/naming-convention -- HTTP header name
        headers: { 'X-Request-ID': 'req-123' },
      });

      const { error } = await result;
      expect(error).toMatchObject({
        message: 'boom',
        status: 500,
        meta: { requestId: 'req-123' },
      });
      expect(errors.errors()).toHaveLength(1);
    });

    it('skips reporting when suppressErrorReporting is set, but still rethrows', async () => {
      const result = once(service.get('/api/things', { suppressErrorReporting: true }));

      const req = expectUrl('/api/things');
      req.flush(serverError('boom'), { status: 500, statusText: 'Internal Server Error' });

      const { error } = await result;
      expect(error).toBeTruthy();
      expect(errors.errors()).toHaveLength(0);
    });

    it('skips reporting only the suppressed statuses', async () => {
      const notFound = once(service.get('/api/things/1', { suppressStatuses: [404] }));
      const req404 = expectUrl('/api/things/1');
      req404.flush({ Message: 'missing' }, { status: 404, statusText: 'Not Found' });
      await notFound;
      expect(errors.errors()).toHaveLength(0);

      const serverErrorCall = once(service.get('/api/things/2'));
      const req500 = expectUrl('/api/things/2');
      req500.flush({ Message: 'boom' }, { status: 500, statusText: 'Internal Server Error' });
      await serverErrorCall;
      expect(errors.errors()).toHaveLength(1);
    });
  });

  describe('edge cases', () => {
    it('issues a fresh request after a get completes (no stale replay)', async () => {
      const first = once(service.get('/api/things'));
      const req1 = expectUrl('/api/things');
      req1.flush({ id: 1 });
      await first;

      const second = once(service.get('/api/things'));
      const req2 = expectUrl('/api/things');
      req2.flush({ id: 2 });

      await expect(second).resolves.toEqual({ next: { id: 2 } });
    });

    it('issues a fresh request after a get errors (failed requests leave the map)', async () => {
      const failing = once(service.get('/api/things'));
      const req1 = expectUrl('/api/things');
      req1.flush(serverError('boom'), { status: 500, statusText: 'Internal Server Error' });
      await failing;

      const retry = once(service.get('/api/things'));
      const req2 = expectUrl('/api/things');
      req2.flush({ id: 3 });

      await expect(retry).resolves.toEqual({ next: { id: 3 } });
    });

    it('caps reported errors at MAX_ERRORS', async () => {
      for (let i = 1; i <= 7; i++) {
        const result = once(service.get(`/api/things/${i}`));
        const req = expectUrl(`/api/things/${i}`);
        req.flush(serverError(`boom ${i}`), { status: 500, statusText: 'Internal Server Error' });
        await result;
      }

      expect(errors.errors()).toHaveLength(5);
      expect(errors.errors().at(-1)?.message).toBe('boom 7');
    });

    it('collapses identical consecutive errors into one report', async () => {
      for (let i = 0; i < 2; i++) {
        const result = once(service.get('/api/things'));
        const req = expectUrl('/api/things');
        req.flush(serverError('boom'), { status: 500, statusText: 'Internal Server Error' });
        await result;
      }

      expect(errors.errors()).toHaveLength(1);
    });

    it('invalidatePending evicts an in-flight get so the next get refetches', async () => {
      const first = once(service.get('/api/things'));
      const post = once(service.post('/api/things', { name: 'x' }, { invalidatePending: ['/api/things'] }));

      const reqGet = controller.expectOne((r) => r.method === 'GET' && r.url === '/api/things');
      const reqPost = controller.expectOne((r) => r.method === 'POST' && r.url === '/api/things');
      reqPost.flush({ id: 9 });
      reqGet.flush({ id: 1 });
      await first;
      await post;

      const second = once(service.get('/api/things'));
      const reqGet2 = controller.expectOne((r) => r.method === 'GET' && r.url === '/api/things');
      reqGet2.flush({ id: 2 });

      await expect(second).resolves.toEqual({ next: { id: 2 } });
    });
  });
});
