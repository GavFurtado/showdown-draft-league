import { HttpContextToken, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError } from 'rxjs';

import { ErrorService } from './error-service';

export const SUPPRESS_ERROR_REPORTING = new HttpContextToken<boolean>(() => false);
export const SUPPRESS_ERROR_STATUSES = new HttpContextToken<number[]>(() => []);

export const errorInterceptor: HttpInterceptorFn = (req, next) => {
  const errors = inject(ErrorService);
  return next(req).pipe(
    catchError((error: unknown) =>
      errors.handle(error, {
        suppressErrorReporting: req.context.get(SUPPRESS_ERROR_REPORTING),
        suppressStatuses: req.context.get(SUPPRESS_ERROR_STATUSES),
      }),
    ),
  );
};
