import { Routes } from '@angular/router';

import { authGuard } from './core/auth/auth-guard';
import { CreateLeagueStore } from './features/create-league/create-league-store';
import { JoinLeagueStore } from './features/join-league/join-league-store';
import { UserLeaguesStore } from './core/shell/app-shell/user-leagues-store';
import { LeagueContextStore } from './features/league/league-context-store';
import { leagueGuard, leagueRoleGuard } from './features/league/league-guard';
import { MemberRole } from './features/league/models/enums/member-role';

export const routes: Routes = [
  { path: '', loadComponent: () => import('./features/landing/landing').then((m) => m.Landing) },
  { path: 'login', loadComponent: () => import('./features/auth/login/login').then((m) => m.Login) },
  { path: 'callback', loadComponent: () => import('./features/auth/callback/callback').then((m) => m.Callback) },
  {
    path: 'my-leagues',
    canActivate: [authGuard],
    loadComponent: () => import('./core/shell/app-shell/shell/shell').then((m) => m.Shell),
    providers: [UserLeaguesStore],
    children: [
      { path: '', loadComponent: () => import('./features/my-leagues/my-leagues/my-leagues').then((m) => m.MyLeagues) },
    ],
  },
  {
    path: 'create-league',
    canActivate: [authGuard],
    loadComponent: () => import('./core/shell/app-shell/shell/shell').then((m) => m.Shell),
    providers: [CreateLeagueStore, UserLeaguesStore],
    children: [
      {
        path: '',
        loadComponent: () => import('./features/create-league/create-league/create-league').then((m) => m.CreateLeague),
      },
    ],
  },
  // Invite links land here: /leagues/:leagueId/join
  {
    path: 'leagues/:leagueId/join',
    canActivate: [authGuard],
    loadComponent: () => import('./core/shell/app-shell/shell/shell').then((m) => m.Shell),
    providers: [JoinLeagueStore, UserLeaguesStore],
    children: [
      {
        path: '',
        loadComponent: () => import('./features/join-league/join-league').then((m) => m.JoinLeague),
      },
    ],
  },

  // Authenticated shell for all league routes (navbar + league subnav)
  {
    path: 'leagues/:leagueId',
    canActivate: [authGuard],
    loadComponent: () => import('./core/shell/app-shell/shell/shell').then((m) => m.Shell),
    children: [
      {
        path: '',
        canActivate: [leagueGuard],
        providers: [LeagueContextStore],
        loadComponent: () => import('./core/shell/league-shell/league-shell/league-shell').then((m) => m.LeagueShell),
        children: [
          {
            path: 'dashboard',
            canActivate: [leagueRoleGuard],
            data: { role: MemberRole.MEMBER },
            loadComponent: () => import('./features/league/dashboard/dashboard').then((m) => m.LeagueDashboard),
          },
          // Admin is owner-gated; remaining sub-pages (draftboard, teamsheets, draft-history,
          // games, standings, transfers) are added with their feature pages in later phases.
          { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
        ],
      },
    ],
  },

  { path: '**', redirectTo: '' },
];
