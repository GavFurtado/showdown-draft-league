import { Component, input } from '@angular/core';
import { AbstractControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { TuiIcon, TuiInput, TuiTextfield, TuiTitle } from '@taiga-ui/core';
import { TuiBlock, TuiInputNumber, TuiTextarea } from '@taiga-ui/kit';
import { EnumLabelPipe } from '../../../shared/pipes/enum-label.pipe';
import { LeagueVisibility } from '../../league/models/enums/league-visibility';

@Component({
  selector: 'app-basic-info',
  imports: [
    ReactiveFormsModule,
    TuiIcon,
    TuiInput,
    TuiInputNumber,
    TuiTextfield,
    TuiTextarea,
    TuiBlock,
    TuiTitle,
    EnumLabelPipe,
  ],
  templateUrl: './basic-info.html',
  styleUrl: './basic-info.css',
})
export class BasicInfo {
  readonly form = input.required<FormGroup>();
  readonly fieldError = input.required<(control: AbstractControl | null) => string | null>();
  readonly showRulesetWarning = input<boolean>(false);
  protected readonly LeagueVisibility = LeagueVisibility;
}
