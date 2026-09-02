import { FormControl, FormGroup } from '@angular/forms';

/**
 * Minimal FormGroup mirroring the CreateLeague shape so step components
 * can render in isolation inside unit tests.
 */
export function createTestLeagueForm(): FormGroup {
  return new FormGroup({
    Name: new FormControl('Test League'),
    RulesetDescription: new FormControl(''),
    MaxPlayers: new FormControl(10),
    MaxPokemonPerPlayer: new FormControl(10),
    MinPokemonPerPlayer: new FormControl(8),
    StartingDraftPoints: new FormControl(140),
    Visibility: new FormControl<'PRIVATE' | 'PUBLIC'>('PRIVATE'),
    Format: new FormGroup({
      IsSnakeRoundDraft: new FormControl(true),
      DraftOrderType: new FormControl<'RANDOM' | 'MANUAL'>('RANDOM'),
      SeasonType: new FormControl<'ROUND_ROBIN_ONLY' | 'BRACKET_ONLY' | 'HYBRID'>('ROUND_ROBIN_ONLY'),
      GroupCount: new FormControl(1),
      PlayoffType: new FormControl<'NONE' | 'SINGLE_ELIM' | 'DOUBLE_ELIM'>('NONE'),
      PlayoffParticipantCount: new FormControl(4),
      PlayoffByesCount: new FormControl(0),
      PlayoffSeedingType: new FormControl<'STANDARD' | 'BYES_ONLY' | 'FULLY_SEEDED'>('STANDARD'),
      AllowTransfers: new FormControl(true),
      TransfersCostCredits: new FormControl(true),
      TransferCreditsPerWindow: new FormControl(2),
      TransferCreditCap: new FormControl(6),
      TransferWindowFrequencyDays: new FormControl(7),
      TransferWindowDuration: new FormControl(48),
      DropCost: new FormControl(1),
      PickupCost: new FormControl(1),
    }),
  });
}
