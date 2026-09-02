import { HttpClient, HttpContext, HttpParams } from '@angular/common/http';
import { Service, inject } from '@angular/core';
import { finalize, shareReplay, tap } from 'rxjs';
import type { Observable } from 'rxjs';

import { environment } from '../../../environments/environment';
import { SUPPRESS_ERROR_REPORTING, SUPPRESS_ERROR_STATUSES } from '../error/error-interceptor';

export interface ApiGetOptions {
  params?: Record<string, string | number | boolean | null | undefined>;
  suppressErrorReporting?: boolean;
  suppressStatuses?: number[];
}

export interface ApiMutationOptions extends ApiGetOptions {
  invalidatePending?: string[];
}

// In-flight dedup + mutation invalidation pattern borrowed from pdz:
// https://github.com/LumarisX/pokemon-draftzone-client
// Original: core/services/api.service.ts. Differences: dedup key includes serialized params,
// invalidation derives from route-paths.ts prefixes, errors flow through the interceptor via HttpContext.
@Service()
export class ApiService {
  private readonly http = inject(HttpClient);
  private readonly pendingRequests = new Map<string, Observable<unknown>>();
  readonly baseUrl = environment.apiUrl;

  get<T>(path: string, options: ApiGetOptions = {}): Observable<T> {
    const key = this.dedupKey(path, options.params);

    const inFlightRequests = this.pendingRequests.get(key);
    // Check if the request to be made is already inFlight,
    if (inFlightRequests) {
      // return that request's Observable if so
      return inFlightRequests as Observable<T>;
    }

    const request$ = this.http
      .get<T>(this.url(path), {
        params: toHttpParams(options.params),
        context: toContext(options),
      })
      .pipe(
        shareReplay({ bufferSize: 1, refCount: true }),
        finalize(() => this.pendingRequests.delete(key)),
      );
    this.pendingRequests.set(key, request$);
    return request$;
  }

  post<T>(path: string, body?: unknown, options: ApiMutationOptions = {}): Observable<T> {
    return this.request('POST', path, body, options);
  }

  put<T>(path: string, body?: unknown, options: ApiMutationOptions = {}): Observable<T> {
    return this.request('PUT', path, body, options);
  }

  patch<T>(path: string, body?: unknown, options: ApiMutationOptions = {}): Observable<T> {
    return this.request('PATCH', path, body, options);
  }

  delete<T>(path: string, options: ApiMutationOptions = {}): Observable<T> {
    return this.request('DELETE', path, undefined, options);
  }

  private request<T>(
    method: 'POST' | 'PUT' | 'PATCH' | 'DELETE',
    path: string,
    body: unknown,
    options: ApiMutationOptions,
  ): Observable<T> {
    return this.http
      .request<T>(method, this.url(path), { body, context: toContext(options) })
      .pipe(tap(() => options.invalidatePending?.forEach((p) => this.invalidate(p))));
  }

  private url(path: string): string {
    return `${this.baseUrl}${path}`;
  }

  private invalidate(prefix: string): void {
    for (const key of this.pendingRequests.keys()) {
      if (key.startsWith(prefix)) {
        this.pendingRequests.delete(key);
      }
    }
  }

  private dedupKey(path: string, params?: ApiGetOptions['params']): string {
    return params ? `${path}?${JSON.stringify(params)}` : path;
  }
}

function toHttpParams(params?: ApiGetOptions['params']): HttpParams | undefined {
  if (!params) return undefined;
  let httpParams = new HttpParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === null || value === undefined) continue;
    httpParams = httpParams.set(key, String(value));
  }
  return httpParams;
}

function toContext(options: ApiGetOptions): HttpContext {
  return new HttpContext()
    .set(SUPPRESS_ERROR_REPORTING, options.suppressErrorReporting ?? false)
    .set(SUPPRESS_ERROR_STATUSES, options.suppressStatuses ?? []);
}
