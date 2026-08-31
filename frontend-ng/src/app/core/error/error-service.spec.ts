import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';

import { ErrorService } from './error-service';

describe('ErrorService', () => {
  let service: ErrorService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(ErrorService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('falls back to a friendly message when a 500 body is not parseable', () => {
    const httpError = new HttpErrorResponse({
      status: 500,
      statusText: 'Internal Server Error',
      url: '/api/leagues',
    });

    let caught: unknown;
    service.handle(httpError).subscribe({ error: (e: unknown) => (caught = e) });

    expect((caught as { message: string }).message).toBe('Something went wrong on the server (500)');
  });

  it('uses the API error message when the body is parseable', () => {
    const httpError = new HttpErrorResponse({
      status: 400,
      statusText: 'Bad Request',
      error: { Message: 'GroupCount must be at least 1' },
      url: '/api/leagues',
    });

    let caught: unknown;
    service.handle(httpError).subscribe({ error: (e: unknown) => (caught = e) });

    expect((caught as { message: string }).message).toBe('GroupCount must be at least 1');
  });
});
