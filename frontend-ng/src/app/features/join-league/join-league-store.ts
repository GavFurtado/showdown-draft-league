import { Service, inject, signal } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { EMPTY, Subject, catchError, finalize, switchMap } from 'rxjs';

import { ClientError } from '../../core/api/api.model';
import { LeagueMember } from '../league/models/league-member.model';
import { JoinLeagueService, LeagueJoinRequest } from './join-league-service';

@Service()
export class JoinLeagueStore {
  private readonly service = inject(JoinLeagueService);

  readonly loading = signal(false);
  readonly loadError = signal<ClientError | null>(null);

  readonly joining = signal(false);
  readonly joinError = signal<ClientError | null>(null);
  readonly joinedMember = signal<LeagueMember | null>(null);

  private readonly leagueId$ = new Subject<string>();

  readonly league = toSignal(
    this.leagueId$.pipe(
      switchMap((leagueId) => {
        this.loading.set(true);
        this.loadError.set(null);
        return this.service.getLeague(leagueId).pipe(
          catchError((err: ClientError) => {
            this.loadError.set(err);
            return EMPTY;
          }),
          finalize(() => this.loading.set(false)),
        );
      }),
    ),
    { initialValue: null },
  );

  load(leagueId: string): void {
    this.leagueId$.next(leagueId);
  }

  join(request: LeagueJoinRequest): void {
    this.joining.set(true);
    this.joinError.set(null);
    this.service
      .join(request)
      .pipe(finalize(() => this.joining.set(false)))
      .subscribe({
        next: (member) => this.joinedMember.set(member),
        error: (err: ClientError) => this.joinError.set(err),
      });
  }
}
