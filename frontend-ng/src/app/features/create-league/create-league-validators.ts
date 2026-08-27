import { AbstractControl, ValidationErrors, ValidatorFn } from '@angular/forms';

// --- FormGroup-level cross-field validators ---

export function minMaxRosterValidator(): ValidatorFn {
  return (group: AbstractControl): ValidationErrors | null => {
    const min = group.get('MinPokemonPerPlayer')?.value;
    const max = group.get('MaxPokemonPerPlayer')?.value;
    return min > 0 && min > max ? { minExceedsMax: true } : null;
  };
}

export function playoffByesValidator(): ValidatorFn {
  return (group: AbstractControl): ValidationErrors | null => {
    const format = group.get('Format');
    const type = format?.get('PlayoffType')?.value;
    const byes = format?.get('PlayoffByesCount')?.value;
    const participants = format?.get('PlayoffParticipantCount')?.value;
    return type !== 'NONE' && byes >= participants ? { byesExceedsParticipants: true } : null;
  };
}

export function transferFrequencyValidator(): ValidatorFn {
  return (group: AbstractControl): ValidationErrors | null => {
    const format = group.get('Format');
    const allowed = format?.get('AllowTransfers')?.value;
    const freq = format?.get('TransferWindowFrequencyDays')?.value;
    return allowed && (freq % 7 !== 0 || freq === 0) ? { invalidFrequency: true } : null;
  };
}

export function singleElimSeedingValidator(): ValidatorFn {
  return (group: AbstractControl): ValidationErrors | null => {
    const format = group.get('Format');
    const type = format?.get('PlayoffType')?.value;
    const seeding = format?.get('PlayoffSeedingType')?.value;
    return type === 'SINGLE_ELIM' && seeding === 'FULLY_SEEDED' ? { invalidSeeding: true } : null;
  };
}

export const ERROR_MESSAGES: Record<string, string> = {
  required: 'This field is required.',
  minlength: 'Value is too short.',
  containsHtml: 'HTML tags are not allowed.',
  forbiddenChars: "The '%' and '\\' characters are not allowed.",
  minExceedsMax: 'Minimum roster size cannot exceed maximum.',
  byesExceedsParticipants: 'Byes must be fewer than participants.',
  invalidFrequency: 'Transfer window frequency must be a multiple of 7.',
  invalidSeeding: 'Fully seeded is not available with single elimination.',
};

const FIELD_LABELS: Record<string, string> = {
  Name: 'League name',
  RulesetDescription: 'Ruleset description',
  MaxPlayers: 'Max players',
  MaxPokemonPerPlayer: 'Max roster size',
  MinPokemonPerPlayer: 'Min roster size',
  StartingDraftPoints: 'Starting draft points',
  Visibility: 'Visibility',
  'Format.GroupCount': 'Group count',
  'Format.PlayoffType': 'Playoff type',
  'Format.PlayoffParticipantCount': 'Playoff participants',
  'Format.TransferWindowFrequencyDays': 'Transfer window frequency',
  'Format.TransferWindowDuration': 'Transfer window duration',
  'Format.TransferCreditsPerWindow': 'Transfer credits per window',
  'Format.TransferCreditCap': 'Transfer credit cap',
  'Format.DropCost': 'Drop cost',
  'Format.PickupCost': 'Pickup cost',
};

export function getErrorMessage(control: AbstractControl, path: string): string | null {
  if (!control.errors) {
    return null;
  }

  const key = Object.keys(control.errors)[0];
  const label = FIELD_LABELS[path] ?? path;

  if (key === 'min') {
    const min = control.errors['min']?.min;
    return min != null ? `${label} must be at least ${min}.` : (ERROR_MESSAGES['min'] ?? `${label} is too low.`);
  }

  if (key === 'max') {
    const max = control.errors['max']?.max;
    return max != null ? `${label} must be at most ${max}.` : (ERROR_MESSAGES['max'] ?? `${label} is too high.`);
  }

  if (key === 'minlength') {
    const requiredLength = control.errors['minlength']?.requiredLength;
    return requiredLength != null
      ? `${label} must be at least ${requiredLength} characters.`
      : ERROR_MESSAGES['minlength'];
  }

  return ERROR_MESSAGES[key] ?? `${label} is invalid.`;
}
