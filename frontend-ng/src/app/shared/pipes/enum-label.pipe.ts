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
    HYBRID: 'Round-Robin + Playoffs',
  };

  transform(value: string, overrides?: Record<string, string>): string {
    if (!value) return '';
    const label = overrides?.[value] ?? EnumLabelPipe.OVERRIDES[value];
    if (label) return label;
    return value
      .replace(/_/g, ' ')
      .toLowerCase()
      .replace(/\b\w/g, (c) => c.toUpperCase());
  }
}
