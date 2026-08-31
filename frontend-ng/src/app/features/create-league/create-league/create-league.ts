import { Component, inject, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { CreateLeagueStore, DEFAULT_FORM } from '../create-league-store';
import { AbstractControl, FormControl, FormGroup, Validators } from '@angular/forms';
import {
  getErrorMessage,
  minMaxRosterValidator,
  playoffByesValidator,
  singleElimSeedingValidator,
  transferFrequencyValidator,
} from '../create-league-validators';
import { forbiddenChars, htmlRejection } from '../../../shared/validation/text-validators';
import { LeaguePlayoffSeedingType } from '../../league/models/enums/league-playoff-seeding-type';
import { LeaguePlayoffType } from '../../league/models/enums/league-playoff-type';
import { LeagueSeasonType } from '../../league/models/enums/league-season-type';
import { LeagueVisibility } from '../../league/models/enums/league-visibility';
import { LeagueCreateRequest } from '../models/league-create-request.model';
import { TuiButton, TuiHint, TuiHintManual, TuiIcon } from '@taiga-ui/core';
import { TuiCardLarge, TuiSurface } from '@taiga-ui/layout';
import { Layout } from '../../../shared/components/layout/layout';
import { TuiConnected, TuiStep, TuiStepper } from '@taiga-ui/kit';
import { BasicInfo } from '../../steps/basic-info/basic-info';
import { DraftSetup } from '../../steps/draft-setup/draft-setup';
import { SeasonSetup } from '../../steps/season-setup/season-setup';
import { PlayoffSetup } from '../../steps/playoff-setup/playoff-setup';
import { TransferMarket } from '../../steps/transfer-market/transfer-market';
import { Review } from '../../steps/review/review';

@Component({
  selector: 'app-create-league',
  imports: [
    TuiButton,
    TuiCardLarge,
    TuiSurface,
    TuiIcon,
    Layout,
    TuiStepper,
    TuiStep,
    TuiConnected,
    TuiHintManual,
    TuiHint,
    BasicInfo,
    DraftSetup,
    SeasonSetup,
    PlayoffSetup,
    TransferMarket,
    Review,
  ],
  templateUrl: './create-league.html',
  styleUrl: './create-league.css',
})
export class CreateLeague {
  protected readonly store = inject(CreateLeagueStore);

  protected readonly stepError = signal<string | null>(null);
  protected readonly publicNoRuleset = signal(false);

  private readonly controlPaths = new Map<AbstractControl, string>();

  protected readonly steps = [
    { label: 'Basic Info', icon: '@tui.edit' },
    { label: 'Draft Setup', icon: '@tui.file-spreadsheet' },
    { label: 'Season Setup', icon: '@tui.calendar' },
    { label: 'Playoff Setup', icon: '@tui.tournament' },
    { label: 'Transfer Market', icon: '@tui.store' },
    { label: 'Review & Create', icon: '@tui.check' },
  ];

  protected stepState(index: number): 'normal' | 'pass' | 'error' {
    if (index < this.stepIndex) return 'pass';
    if (index === this.stepIndex && this.stepError()) return 'error';
    return 'normal';
  }

  protected isStepDisabled(index: number): boolean {
    if (index === 5) return true;
    if (index === 3 && this.store.isPlayoffStepSkipped()) return true;
    return index > this.stepIndex;
  }

  protected readonly boundFieldError = (control: AbstractControl | null): string | null => {
    return this.fieldError(control);
  };

  protected readonly form = new FormGroup(
    {
      Name: new FormControl(DEFAULT_FORM.Name, {
        validators: [Validators.required, Validators.minLength(3), htmlRejection(), forbiddenChars()],
      }),
      RulesetDescription: new FormControl(DEFAULT_FORM.RulesetDescription, {
        validators: [htmlRejection(), forbiddenChars()],
      }),
      MaxPlayers: new FormControl(DEFAULT_FORM.MaxPlayers, {
        validators: [Validators.required, Validators.min(2), Validators.max(32)],
      }),
      MaxPokemonPerPlayer: new FormControl(DEFAULT_FORM.MaxPokemonPerPlayer, {
        validators: [Validators.required, Validators.min(1), Validators.max(20)],
      }),
      // NOTE: Min. 0 because server treats MinPokemonPerPlayer = 0 specially (i think i don't remember)
      MinPokemonPerPlayer: new FormControl(DEFAULT_FORM.MinPokemonPerPlayer, {
        validators: [Validators.required, Validators.min(0), Validators.max(20)],
      }),
      StartingDraftPoints: new FormControl(DEFAULT_FORM.StartingDraftPoints, {
        validators: [Validators.required, Validators.min(0)],
      }),
      Visibility: new FormControl<'PRIVATE' | 'PUBLIC'>(DEFAULT_FORM.Visibility, { validators: Validators.required }),

      Format: new FormGroup(
        {
          IsSnakeRoundDraft: new FormControl(DEFAULT_FORM.Format.IsSnakeRoundDraft, {
            validators: [Validators.required],
          }),
          DraftOrderType: new FormControl(DEFAULT_FORM.Format.DraftOrderType, { validators: [Validators.required] }),
          SeasonType: new FormControl(DEFAULT_FORM.Format.SeasonType, { validators: [Validators.required] }),
          GroupCount: new FormControl(DEFAULT_FORM.Format.GroupCount, {
            validators: [Validators.required, Validators.min(1), Validators.max(2)],
          }),
          PlayoffSeedingType: new FormControl(DEFAULT_FORM.Format.PlayoffSeedingType),
          PlayoffType: new FormControl(DEFAULT_FORM.Format.PlayoffType, { validators: [Validators.required] }),
          PlayoffParticipantCount: new FormControl(DEFAULT_FORM.Format.PlayoffParticipantCount, {
            validators: Validators.min(2),
          }),
          PlayoffByesCount: new FormControl(DEFAULT_FORM.Format.PlayoffByesCount),
          AllowTransfers: new FormControl(DEFAULT_FORM.Format.AllowTransfers, { validators: [Validators.required] }),
          TransfersCostCredits: new FormControl(DEFAULT_FORM.Format.TransfersCostCredits),
          TransferCreditsPerWindow: new FormControl(DEFAULT_FORM.Format.TransferCreditsPerWindow, {
            validators: [Validators.min(0)],
          }),
          TransferCreditCap: new FormControl(DEFAULT_FORM.Format.TransferCreditCap, {
            validators: [Validators.min(0)],
          }),
          TransferWindowFrequencyDays: new FormControl(DEFAULT_FORM.Format.TransferWindowFrequencyDays, {
            validators: [Validators.min(7)],
          }),
          TransferWindowDuration: new FormControl(DEFAULT_FORM.Format.TransferWindowDuration, {
            validators: [Validators.min(1)],
          }),
          DropCost: new FormControl(DEFAULT_FORM.Format.DropCost, { validators: [Validators.min(0)] }),
          PickupCost: new FormControl(DEFAULT_FORM.Format.PickupCost, { validators: [Validators.min(0)] }),
        },
        { validators: [playoffByesValidator(), transferFrequencyValidator(), singleElimSeedingValidator()] },
      ),
    },
    { validators: [minMaxRosterValidator()] },
  );

  constructor() {
    this.buildControlPaths(this.form.controls, '');

    this.form.valueChanges.subscribe((value) => {
      this.store.form.set(value as LeagueCreateRequest);
      this.publicNoRuleset.set(value.Visibility === LeagueVisibility.PUBLIC && !value.RulesetDescription);
    });

    // Cross-field clamps & forced values
    const minCtrl = this.form.get('MinPokemonPerPlayer')!;
    const maxCtrl = this.form.get('MaxPokemonPerPlayer')!;
    const participantsCtrl = this.form.get('Format.PlayoffParticipantCount')!;
    const byesCtrl = this.form.get('Format.PlayoffByesCount')!;
    const playoffTypeCtrl = this.form.get('Format.PlayoffType')!;
    const seedingCtrl = this.form.get('Format.PlayoffSeedingType')!;

    maxCtrl.valueChanges.pipe(takeUntilDestroyed()).subscribe((max) => {
      if (minCtrl.value && minCtrl.value > 0 && minCtrl.value > max!) {
        minCtrl.setValue(max!, { emitEvent: false });
      }
    });

    participantsCtrl.valueChanges.pipe(takeUntilDestroyed()).subscribe((participants) => {
      if (byesCtrl.value && byesCtrl.value >= participants!) {
        byesCtrl.setValue(Math.max(0, participants! - 1), { emitEvent: false });
      }
    });

    playoffTypeCtrl.valueChanges.pipe(takeUntilDestroyed()).subscribe((type) => {
      if (
        type === LeaguePlayoffType.SINGLE_ELIM &&
        seedingCtrl.value === LeaguePlayoffSeedingType.FULLY_SEEDED
      ) {
        seedingCtrl.setValue(LeaguePlayoffSeedingType.STANDARD);
      }
    });

    seedingCtrl.valueChanges.pipe(takeUntilDestroyed()).subscribe((seeding) => {
      if (seeding === LeaguePlayoffSeedingType.STANDARD && byesCtrl.value !== 0) {
        byesCtrl.setValue(0);
      }
    });

    const seasonTypeCtrl = this.form.get('Format.SeasonType')!;
    seasonTypeCtrl.valueChanges.pipe(takeUntilDestroyed()).subscribe((seasonType) => {
      if (seasonType === LeagueSeasonType.ROUND_ROBIN_ONLY) {
        if (playoffTypeCtrl.value !== LeaguePlayoffType.NONE) {
          playoffTypeCtrl.setValue(LeaguePlayoffType.NONE);
        }
      } else if (playoffTypeCtrl.value === LeaguePlayoffType.NONE) {
        // BRACKET_ONLY / HYBRID imply a top cut; default it when none chosen yet.
        playoffTypeCtrl.setValue(LeaguePlayoffType.SINGLE_ELIM);
      }
    });
  }

  get stepIndex(): number {
    return this.store.currentStep();
  }

  set stepIndex(value: number) {
    // Forward navigation must go through validation (stepper forward clicks are disabled).
    if (value > this.stepIndex) {
      this.onNext();
      return;
    }

    let target = Math.max(value, 0);
    // Land before the playoff step when it is skipped
    if (target === 3 && this.store.isPlayoffStepSkipped()) {
      target = 2;
    }

    this.stepError.set(null);
    this.store.currentStep.set(target);
  }

  protected fieldError(control: AbstractControl | null): string | null {
    if (!control || control.valid || !control.touched) return null;
    const path = this.controlPaths.get(control);
    return path ? getErrorMessage(control, path) : null;
  }

  private buildControlPaths(controls: Record<string, AbstractControl>, prefix: string): void {
    for (const [name, control] of Object.entries(controls)) {
      const path = prefix ? `${prefix}.${name}` : name;
      if (control instanceof FormGroup) {
        this.buildControlPaths(control.controls, path);
      } else {
        this.controlPaths.set(control, path);
      }
    }
  }

  protected onNext(): void {
    this.form.markAllAsTouched();
    this.form.updateValueAndValidity();

    if (!this.form.valid) {
      this.stepError.set('Please fix the errors above.');
      return;
    }

    this.stepError.set(null);
    this.store.next(true);
  }

  protected onBack(): void {
    this.store.back();
  }

  protected onSubmit(): void {
    this.form.markAllAsTouched();
    this.form.updateValueAndValidity();

    if (!this.form.valid) {
      this.stepError.set('Please fix the errors above.');
      return;
    }

    this.stepError.set(null);
    this.store.submitForm(); // we're disregarding the response body
  }
}
