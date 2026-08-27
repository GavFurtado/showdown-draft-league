import { Routes } from '@angular/router';

import { authGuard, onboardingGuard } from './core/auth/auth-guard';
import { MyLeagueStore } from './features/my-leagues/my-league-store';
import { CreateLeagueStore } from './features/create-league/create-league-store';
import { JoinLeagueStore } from './features/join-league/join-league-store';

export const routes: Routes = [
  { path: '', loadComponent: () => import('./features/landing/landing').then((m) => m.Landing) },
  { path: 'login', loadComponent: () => import('./features/auth/login/login').then((m) => m.Login) },
  { path: 'callback', loadComponent: () => import('./features/auth/callback/callback').then((m) => m.Callback) },
  {
    path: 'onboarding',
    canActivate: [authGuard],
    loadComponent: () =>
      import('./features/auth/onboarding/onboarding').then((m) => m.Onboarding),
  },
  {
    path: 'my-leagues',
    canActivate: [onboardingGuard],
    loadComponent: () => import('./core/layout/shell').then((m) => m.Shell),
    providers: [MyLeagueStore],
    children: [
      { path: '', loadComponent: () => import('./features/my-leagues/my-leagues/my-leagues').then((m) => m.MyLeagues) },
      {
        path: ':id',
        loadComponent: () =>
          import('./features/my-leagues/league-detail/league-detail').then((m) => m.LeagueDetail),
      },
    ],
  },
  {
    path: 'create-league',
    canActivate: [onboardingGuard],
    loadComponent: () => import('./core/layout/shell').then((m) => m.Shell),
    providers: [CreateLeagueStore],
    children: [
      {
        path: '',
        loadComponent: () => import('./features/create-league/create-league/create-league').then((m) => m.CreateLeague),
      },
    ],
  },
  // Invite links land here: /:leagueId/join
  {
    path: ':leagueId/join',
    canActivate: [onboardingGuard],
    loadComponent: () => import('./core/layout/shell').then((m) => m.Shell),
    providers: [JoinLeagueStore],
    children: [
      {
        path: '',
        loadComponent: () => import('./features/join-league/join-league').then((m) => m.JoinLeague),
      },
    ],
  },

  { path: '**', redirectTo: '' },
];
