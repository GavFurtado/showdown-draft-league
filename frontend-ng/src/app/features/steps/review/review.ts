import { Component, input } from '@angular/core';
import { FormGroup, ReactiveFormsModule } from '@angular/forms';
import { EnumLabelPipe } from '../../../shared/pipes/enum-label.pipe';
import { LeaguePlayoffType } from '../../league/models/enums/league-playoff-type';

@Component({
  selector: 'app-review',
  imports: [ReactiveFormsModule, EnumLabelPipe],
  templateUrl: './review.html',
  styleUrl: './review.css',
})
export class Review {
  readonly form = input.required<FormGroup>();

  protected readonly seasonLabels: Record<string, string> = {
    HYBRID: 'Hybrid (Round Robin + Playoffs)',
    BRACKET_ONLY: 'Tournament Only',
  };

  protected get hasPlayoffs(): boolean {
    return this.form().get('Format.PlayoffType')!.value !== LeaguePlayoffType.NONE;
  }

  protected get allowTransfers(): boolean {
    return !!this.form().get('Format.AllowTransfers')!.value;
  }

  protected get rosterSize(): string {
    const min = this.form().get('MinPokemonPerPlayer')!.value;
    const max = this.form().get('MaxPokemonPerPlayer')!.value;

    return `${min && min > 0 ? min : 'No Min'} – ${max} Pokémon`;
  }
}
