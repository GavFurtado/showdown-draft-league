import { Service, inject, signal } from '@angular/core';
import { ClientError } from '../../core/api/api.model';
import { MyLeagueService } from './my-league-service';
import { catchError, EMPTY, finalize, Subject, switchMap } from 'rxjs';
import { toSignal } from '@angular/core/rxjs-interop';

@Service()
export class MyLeagueStore {
  private readonly service = inject(MyLeagueService);

  readonly activeTab = signal<'myLeagues' | 'publicLeagues'>('myLeagues');
  readonly loading = signal<boolean>(false);
  readonly error = signal<ClientError | null>(null);

  private readonly trigger$ = new Subject<void>();

  readonly myLeagues = toSignal(
    this.trigger$.pipe(
      switchMap(() => {
        this.loading.set(true);
        return this.service.getMyLeagues().pipe(
          catchError((err: ClientError) => {
            this.error.set(err);
            return EMPTY;
          }),
          finalize(() => this.loading.set(false)),
        );
      }),
    ),
    { initialValue: [] },
  );

  constructor() {
    this.trigger$.next();
  }
}
