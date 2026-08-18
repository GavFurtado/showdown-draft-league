import { Pipe, PipeTransform } from '@angular/core';

/**
 * Transforms SCREAMING_SNAKE_CASE into Title Case.
 * Optional second arg: an override map, e.g.
 *   { ROUND_ROBIN_ONLY: 'Round Robin', TOURNAMENT_ONLY: 'Tournament' }
 */
@Pipe({ name: 'enumLabel', standalone: true })
export class EnumLabelPipe implements PipeTransform {
  private static readonly OVERRIDES: Record<string, string> = {
    ROUND_ROBIN_ONLY: 'Round-Robin',
    BRACKET_ONLY: 'Tournament',
    TOURNAMENT_ONLY: 'Tournament',
    HYBRID: 'Round-Robin + Playoffs',
  };

  transform(value: string): string {
    if (!value) return '';
    if (EnumLabelPipe.OVERRIDES[value]) return EnumLabelPipe.OVERRIDES[value];
    return value
      .replace(/_/g, ' ')
      .toLowerCase()
      .replace(/\b\w/g, (c) => c.toUpperCase());
  }
}
