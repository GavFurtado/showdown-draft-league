import { inject, Service, signal } from '@angular/core';
import { LeagueCreateRequest } from './models/league-create-request.model';
import { ClientError } from '../../core/api/api.model';
import { LeagueVisibility } from '../league/models/enums/league-visibility';
import { LeagueDraftOrderType } from '../league/models/enums/league-draft-order-type';
import { LeagueSeasonType } from '../league/models/enums/league-season-type';
import { LeaguePlayoffType } from '../league/models/enums/league-playoff-type';
import { LeaguePlayoffSeedingType } from '../league/models/enums/league-playoff-seeding-type';
import { LeagueFormatRequest } from './models/league-format-request.model';
import { Router } from '@angular/router';
import { CreateLeagueService } from './create-league-service';
import { catchError, EMPTY, finalize, Subject, switchMap, tap } from 'rxjs';
import { toSignal } from '@angular/core/rxjs-interop';

export const DEFAULT_FORM: LeagueCreateRequest = {
  Name: '',
  RulesetDescription: '',
  MaxPlayers: 10,
  MaxPokemonPerPlayer: 10,
  MinPokemonPerPlayer: 8,
  StartingDraftPoints: 140,
  Visibility: LeagueVisibility.PRIVATE,
  Format: {
    IsSnakeRoundDraft: true,
    DraftOrderType: LeagueDraftOrderType.RANDOM,
    SeasonType: LeagueSeasonType.ROUND_ROBIN_ONLY,
    GroupCount: 1,
    PlayoffType: LeaguePlayoffType.NONE,
    PlayoffParticipantCount: 4,
    PlayoffByesCount: 0,
    PlayoffSeedingType: LeaguePlayoffSeedingType.STANDARD,
    AllowTransfers: true,
    TransfersCostCredits: true,
    TransferCreditsPerWindow: 2,
    TransferCreditCap: 6,
    TransferWindowFrequencyDays: 7,
    TransferWindowDuration: 48,
    DropCost: 1,
    PickupCost: 1,
  },
};

@Service()
export class CreateLeagueStore {
  private readonly router = inject(Router);
  private readonly service = inject(CreateLeagueService);
  private readonly submit$ = new Subject<void>();

  readonly isSubmitting = signal<boolean>(false);
  readonly currentStep = signal(0); // It's a 0 based index
  readonly form = signal<LeagueCreateRequest>({ ...DEFAULT_FORM });
  readonly error = signal<ClientError | null>(null);

  readonly MAX_STEPS = 6;

  readonly submitting = toSignal(
    this.submit$.pipe(
      switchMap(() => {
        this.isSubmitting.set(true);
        this.error.set(null);
        return this.service.createLeague(this.form()).pipe(
          tap((league) => void this.router.navigateByUrl(`/my-leagues/${league.ID}`)),
          catchError((err) => {
            this.error.set(err);
            return EMPTY;
          }),
          finalize(() => this.isSubmitting.set(false)),
        );
      }),
    ),
  );

  updateField<K extends keyof LeagueCreateRequest>(key: K, value: LeagueCreateRequest[K]): void {
    this.form.update((form) => ({ ...form, [key]: value }));
  }

  updateFormatField<K extends keyof LeagueFormatRequest>(key: K, value: LeagueFormatRequest[K]): void {
    this.form.update((form) => ({
      ...form,
      Format: {
        ...form.Format,
        [key]: value,
      },
    }));
  }

  isPlayoffStepSkipped(): boolean {
    return this.form().Format.SeasonType === LeagueSeasonType.ROUND_ROBIN_ONLY;
  }

  next(formValid: boolean): void {
    if (!formValid) {
      return;
    }

    let nextIdx = this.currentStep() + 1;
    // Skip playoff step (step 4) if ROUND_ROBIN_ONLY season type
    if (this.currentStep() === 2 && this.isPlayoffStepSkipped()) {
      nextIdx = 4;
    }

    this.currentStep.set(nextIdx);
  }

  back(): void {
    this.error.set(null);
    let prevIdx = this.currentStep() - 1;

    // Reverse-skip playoff step if ROUND_ROBIN_ONLY season type
    if (this.currentStep() === 4 && this.isPlayoffStepSkipped()) {
      prevIdx = 2;
    }

    this.currentStep.set(Math.max(prevIdx, 0));
  }

  jumpTo(step: number): void {
    let target = Math.max(step, 0);

    // Skip playoff step if ROUND_ROBIN_ONLY season type
    if (target === 3 && this.isPlayoffStepSkipped()) {
      target = 2;
    }

    this.error.set(null);
    this.currentStep.set(target);
  }

  submitForm() {
    this.submit$.next();
  }
}
