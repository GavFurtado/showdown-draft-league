import { Routes } from '@angular/router';

import { authGuard } from './core/auth/auth-guard';

export const routes: Routes = [
  { path: '', loadComponent: () => import('./features/landing/landing').then((m) => m.Landing) },
  { path: 'login', loadComponent: () => import('./features/auth/login/login').then((m) => m.Login) },
  { path: 'callback', loadComponent: () => import('./features/auth/callback/callback').then((m) => m.Callback) },
  {
    path: 'my-leagues',
    canActivate: [authGuard],
    loadComponent: () => import('./core/layout/shell').then((m) => m.Shell),
    children: [
      { path: '', loadComponent: () => import('./features/my-leagues/my-leagues/my-leagues').then((m) => m.MyLeagues) },
    ],
  },
  { path: '**', redirectTo: '' },
];
