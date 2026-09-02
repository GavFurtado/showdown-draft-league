import { Component, input } from '@angular/core';
import { AbstractControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { TuiInput, TuiRadio, TuiTextfield, TuiTitle } from '@taiga-ui/core';
import { TuiBlock, TuiInputNumber, TuiTextarea } from '@taiga-ui/kit';
import { EnumLabelPipe } from '../../../shared/pipes/enum-label.pipe';
import { LeagueSeasonType } from '../../league/models/enums/league-season-type';

@Component({
  selector: 'app-season-setup',
  imports: [
    ReactiveFormsModule,
    TuiInput,
    TuiInputNumber,
    TuiTextfield,
    TuiTextarea,
    TuiBlock,
    TuiTitle,
    EnumLabelPipe,
    TuiRadio,
  ],
  templateUrl: './season-setup.html',
  styleUrl: './season-setup.css',
})
export class SeasonSetup {
  readonly form = input.required<FormGroup>();
  readonly fieldError = input.required<(control: AbstractControl | null) => string | null>();
  protected readonly SeasonType = LeagueSeasonType;

  protected readonly seasonLabels: Record<string, string> = {
    HYBRID: 'Hybrid (Round Robin + Playoffs)',
    BRACKET_ONLY: 'Tournament Only',
  };
}
