import { Component, input } from '@angular/core';
import { AbstractControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { TuiInput, TuiRadio, TuiTextfield, TuiTitle } from '@taiga-ui/core';
import { TuiBlock, TuiInputNumber } from '@taiga-ui/kit';
import { EnumLabelPipe } from '../../../shared/pipes/enum-label.pipe';
import { LeaguePlayoffType } from '../../league/models/enums/league-playoff-type';
import { LeaguePlayoffSeedingType } from '../../league/models/enums/league-playoff-seeding-type';

@Component({
  selector: 'app-playoff-setup',
  imports: [ReactiveFormsModule, TuiInput, TuiInputNumber, TuiRadio, TuiTextfield, TuiTitle, TuiBlock, EnumLabelPipe],
  templateUrl: './playoff-setup.html',
  styleUrl: './playoff-setup.css',
})
export class PlayoffSetup {
  readonly form = input.required<FormGroup>();
  readonly fieldError = input.required<(control: AbstractControl | null) => string | null>();

  protected readonly PlayoffType = LeaguePlayoffType;
  protected readonly Seeding = LeaguePlayoffSeedingType;

  protected get hasPlayoffs(): boolean {
    return this.form().get('Format.PlayoffType')!.value !== LeaguePlayoffType.NONE;
  }

  protected get isStandardSeeding(): boolean {
    return this.form().get('Format.PlayoffSeedingType')!.value === LeaguePlayoffSeedingType.STANDARD;
  }

  protected get seeds(): LeaguePlayoffSeedingType[] {
    const isSingleElim = this.form().get('Format.PlayoffType')!.value === LeaguePlayoffType.SINGLE_ELIM;

    return isSingleElim
      ? [LeaguePlayoffSeedingType.STANDARD, LeaguePlayoffSeedingType.BYES_ONLY]
      : [
          LeaguePlayoffSeedingType.STANDARD,
          LeaguePlayoffSeedingType.BYES_ONLY,
          LeaguePlayoffSeedingType.FULLY_SEEDED,
        ];
  }
}
