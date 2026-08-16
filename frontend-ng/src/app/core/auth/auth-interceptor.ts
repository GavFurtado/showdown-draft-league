import { HttpInterceptorFn } from '@angular/common/http';

// Stub. Phase 1: attach `Authorization: Bearer <jwt>` from AuthService and
// prime the session via provideAppInitializer(AuthService.prime()).
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  return next(req);
};
