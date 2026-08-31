import { Component, input, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TuiCard, TuiCardRow, TuiHeader } from '@taiga-ui/layout';
import { TuiButton, TuiExpand, TuiIcon, TuiTitle } from '@taiga-ui/core';
import { TuiChevron } from '@taiga-ui/kit';
import { EnumLabelPipe } from '../../../shared/pipes/enum-label.pipe';
import { StatusBadge } from '../../../shared/components/status-badge/status-badge';
import { League } from '../../league/models/league.model';
import { MemberRole } from '../../league/models/enums/member-role';

@Component({
  selector: 'app-league-card',
  imports: [
    EnumLabelPipe,
    RouterLink,
    TuiCard,
    TuiCardRow,
    TuiHeader,
    TuiButton,
    TuiChevron,
    TuiExpand,
    TuiIcon,
    TuiTitle,
    StatusBadge,
  ],
  templateUrl: './league-card.html',
  styleUrl: './league-card.css',
})
export class LeagueCard {
  readonly league = input.required<League>();
  protected readonly collapsed = signal(true);
  protected readonly memberRole = MemberRole;
}
