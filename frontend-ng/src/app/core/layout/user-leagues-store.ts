import { Service, inject, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { EMPTY, Subject, catchError, finalize, switchMap } from 'rxjs';

import { ClientError } from '../api/api.model';
import { League } from '../../features/league/models/league.model';
import { MyLeagueService } from '../../features/my-leagues/my-league-service';

@Service()
export class UserLeaguesStore {
  private readonly service = inject(MyLeagueService);

  private readonly trigger$ = new Subject<void>();

  readonly loading = signal(false);
  readonly error = signal<ClientError | null>(null);

  readonly leagues = toSignal(
    this.trigger$.pipe(
      switchMap(() => {
        this.loading.set(true);
        this.error.set(null);
        return this.service.getMyLeagues().pipe(
          catchError((err: ClientError) => {
            this.error.set(err);
            return EMPTY;
          }),
          finalize(() => this.loading.set(false)),
        );
      }),
    ),
    { initialValue: [] as League[] },
  );

  constructor() {
    this.trigger$.next();
  }

  refetch(): void {
    this.trigger$.next();
  }
}