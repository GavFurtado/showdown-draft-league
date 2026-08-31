import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { AuthService } from '../../core/auth/auth-service';
import { LeagueContextStore } from './league-context-store';
import { MemberRole } from './models/enums/member-role';

// leagueGuard just keeps non-members out of the UI and
// blocks cold first navigation once while LeagueContext loads.
export const leagueGuard: CanActivateFn = async (route) => {
  const auth = inject(AuthService);
  const context = inject(LeagueContextStore);
  const router = inject(Router);

  if (!auth.isLoggedIn()) return router.createUrlTree(['/login']);

  const leagueId = route.paramMap.get('leagueId') as string | null;
  if (!leagueId) return router.createUrlTree(['/my-leagues']);

  const isMember = await context.ensureLoaded(leagueId as never);
  return isMember ? true : router.createUrlTree(['/my-leagues']);
};

/**
 * Enforces a per-route `data: { role }` on league children. Mirrors the backend's RBAC
 * permissions as a UX gate only — the server is the source of truth.
 */
export const leagueRoleGuard: CanActivateFn = (route) => {
  const context = inject(LeagueContextStore);
  const router = inject(Router);

  const required = route.data['role'] as MemberRole | undefined;
  if (!required) return true;
  if (required === MemberRole.MEMBER) return true;

  const role = context.userMember()?.Role;
  const allowed = role === MemberRole.OWNER || (required === MemberRole.MODERATOR && role === MemberRole.MODERATOR);
  return allowed ? true : router.createUrlTree(['/my-leagues']);
};
