import { Component, computed, input } from '@angular/core';
import { TuiBadge } from '@taiga-ui/kit/components/badge';
import { TuiStatus } from '@taiga-ui/kit/components/status';
import { STATUS_PRESENTATIONS, TONE_CLASSES, type StatusDomain } from './presentations';

@Component({
  selector: 'app-status-badge',
  imports: [TuiBadge, TuiStatus],
  templateUrl: './status-badge.html',
  styleUrl: './status-badge.css',
})
export class StatusBadge {
  readonly status = input<string>();
  readonly domain = input<StatusDomain>('global');
  readonly label = input<string>();
  readonly size = input<'s' | 'm' | 'l' | 'xl'>('m');

  protected readonly presentation = computed(() => {
    const key = (this.status() ?? '').toUpperCase();
    return (
      STATUS_PRESENTATIONS[this.domain()][key] ??
      STATUS_PRESENTATIONS.global[key] ?? { label: this.status() ?? '', tone: 'gray' as const }
    );
  });

  protected readonly classes = computed(() => TONE_CLASSES[this.presentation().tone]);

  protected readonly displayLabel = computed(() => this.label() ?? this.presentation().label);
}
