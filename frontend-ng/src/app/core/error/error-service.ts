import { HttpErrorResponse } from '@angular/common/http';
import { Service, signal } from '@angular/core';
import { throwError } from 'rxjs';
import type { Observable } from 'rxjs';

import { ApiErrorResponse, ClientError } from '../api/api.model';

export const MAX_ERRORS = 5;

export interface ErrorHandlingOptions {
  suppressErrorReporting?: boolean;
  suppressStatuses?: number[];
}

@Service()
export class ErrorService {
  readonly errors = signal<ClientError[]>([]);

  handle(error: unknown, options?: ErrorHandlingOptions): Observable<never> {
    const clientError = this.normalize(error);
    const suppressed = options?.suppressErrorReporting || options?.suppressStatuses?.includes(clientError.status ?? -1);
    if (!suppressed) {
      this.report(clientError);
    }
    return throwError(() => clientError);
  }

  private normalize(error: unknown): ClientError {
    if (error instanceof HttpErrorResponse) {
      const apiError = error.error as ApiErrorResponse | null;
      return {
        message: apiError?.Message ?? this.fallbackMessage(error),
        status: error.status,
        statusText: error.statusText,
        url: error.url ?? undefined,
        code: null,
        details: apiError,
        meta: { requestId: error.headers.get('X-Request-ID'), timestamp: new Date().toISOString() },
      };
    }
    if (error instanceof Error) {
      return {
        message: error.message,
        code: null,
        details: error,
        stack: error.stack,
        meta: { requestId: null, timestamp: new Date().toISOString() },
      };
    }
    return {
      message: String(error),
      code: null,
      details: error,
      meta: { requestId: null, timestamp: new Date().toISOString() },
    };
  }

  /**
   * Builds a human-readable fallback when the server didn't return a
   * parseable API error body (e.g. a panic or empty 500 response).
   */
  private fallbackMessage(error: HttpErrorResponse): string {
    if (error.status >= 500) {
      return `Something went wrong on the server (${error.status})`;
    }
    if (error.status === 0) {
      return 'Unable to reach the server. Please try again.';
    }
    if (error.status > 0) {
      const statusText = error.statusText?.trim();
      return statusText ? `Request failed (${error.status} ${statusText})` : `Request failed (${error.status})`;
    }
    return error.message;
  }

  private report(clientError: ClientError): void {
    this.errors.update((current) => {
      const last = current.at(-1);
      if (last && areErrorsEquivalent(last, clientError)) {
        return current;
      }
      return [...current, clientError].slice(-MAX_ERRORS);
    });
  }
}

function areErrorsEquivalent(a: ClientError, b: ClientError): boolean {
  return a.message === b.message && a.status === b.status && a.url === b.url;
}
