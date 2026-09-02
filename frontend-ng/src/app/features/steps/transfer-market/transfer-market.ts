import { Component, input } from '@angular/core';
import { AbstractControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { TuiInput, TuiTextfield, TuiTitle } from '@taiga-ui/core';
import { TuiInputNumber, TuiSwitch } from '@taiga-ui/kit';

@Component({
  selector: 'app-transfer-market',
  imports: [ReactiveFormsModule, TuiInput, TuiInputNumber, TuiTextfield, TuiTitle, TuiSwitch],
  templateUrl: './transfer-market.html',
  styleUrl: './transfer-market.css',
})
export class TransferMarket {
  readonly form = input.required<FormGroup>();
  readonly fieldError = input.required<(control: AbstractControl | null) => string | null>();
}
