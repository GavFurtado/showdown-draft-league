import { Component, input, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TuiCard, TuiCardRow, TuiHeader } from '@taiga-ui/layout';
import { TuiButton, TuiExpand, TuiIcon, TuiTitle } from '@taiga-ui/core';
import { TuiChevron } from '@taiga-ui/kit';
import { EnumLabelPipe } from '../../../shared/pipes/enum-label.pipe';
import { AutoShrink } from '../../../shared/directives/auto-shrink';
import { StatusBadge } from '../../../shared/components/status-badge/status-badge';
import { League } from '../../league/models/league.model';

@Component({
  selector: 'app-league-card',
  imports: [AutoShrink, EnumLabelPipe, RouterLink, TuiCard, TuiCardRow, TuiHeader, TuiButton, TuiChevron, TuiExpand, TuiIcon, TuiTitle, StatusBadge],
  templateUrl: './league-card.html',
  styleUrl: './league-card.css',
})
export class LeagueCard {
  readonly league = input.required<League>();
  protected readonly collapsed = signal(true);
}
