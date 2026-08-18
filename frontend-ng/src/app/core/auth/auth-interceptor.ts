import { HttpInterceptorFn } from '@angular/common/http';

import { TOKEN_KEY } from './auth-service';

// Attaches the JWT to resource API calls only. /auth/* endpoints (login, callback,
// logout) are deliberately tokenless — the handshake itself issues the token.
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const token = localStorage.getItem(TOKEN_KEY);
  if (token && req.url.includes('/api/')) {
    req = req.clone({ setHeaders: { Authorization: `Bearer ${token}` } });
  }
  return next(req);
};
