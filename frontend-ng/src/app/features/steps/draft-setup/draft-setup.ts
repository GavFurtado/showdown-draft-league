import { Component, DestroyRef, effect, inject, input, untracked } from '@angular/core';
import { AbstractControl, FormControl, FormGroup, ReactiveFormsModule } from '@angular/forms';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { TuiInput, TuiRadio, TuiTextfield, TuiTitle } from '@taiga-ui/core';
import { TuiBlock, TuiInputNumber, TuiInputRange, TuiTextarea } from '@taiga-ui/kit';
import { EnumLabelPipe } from '../../../shared/pipes/enum-label.pipe';
import { LeagueDraftOrderType } from '../../league/models/enums/league-draft-order-type';

@Component({
  selector: 'app-draft-setup',
  imports: [
    ReactiveFormsModule,
    TuiInput,
    TuiInputNumber,
    TuiInputRange,
    TuiTextfield,
    TuiTextarea,
    TuiBlock,
    TuiTitle,
    EnumLabelPipe,
    TuiRadio,
  ],
  templateUrl: './draft-setup.html',
  styleUrl: './draft-setup.css',
})
export class DraftSetup {
  readonly form = input.required<FormGroup>();
  readonly fieldError = input.required<(control: AbstractControl | null) => string | null>();
  protected readonly LeagueDraftOrderType = LeagueDraftOrderType;

  protected readonly pokemonRange = new FormControl<[number, number]>([8, 10]);

  private readonly destroyRef = inject(DestroyRef);

  constructor() {
    effect(() => {
      this.form(); // track

      untracked(() => {
        const group = this.form();
        const minCtrl = group.get('MinPokemonPerPlayer')!;
        const maxCtrl = group.get('MaxPokemonPerPlayer')!;

        this.pokemonRange.setValue([minCtrl.value ?? 8, maxCtrl.value ?? 10], { emitEvent: false });

        this.pokemonRange.valueChanges
          .pipe(takeUntilDestroyed(this.destroyRef))
          .subscribe((val) => {
            if (!val) return;
            minCtrl.setValue(val[0], { emitEvent: false });
            maxCtrl.setValue(val[1], { emitEvent: false });
          });

        minCtrl.valueChanges
          .pipe(takeUntilDestroyed(this.destroyRef))
          .subscribe((val) => {
            const cur = this.pokemonRange.value!;
            if (val !== cur[0]) {
              this.pokemonRange.setValue([val ?? 0, cur[1]], { emitEvent: false });
            }
          });

        maxCtrl.valueChanges
          .pipe(takeUntilDestroyed(this.destroyRef))
          .subscribe((val) => {
            const cur = this.pokemonRange.value!;
            if (val !== cur[1]) {
              this.pokemonRange.setValue([cur[0], val ?? 10], { emitEvent: false });
            }
          });
      });
    });
  }
}
