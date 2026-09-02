import { Component, computed, inject, input } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { TuiItem } from '@taiga-ui/cdk';
import { TuiTabs } from '@taiga-ui/kit';

import { LeagueContextStore } from '../../../../features/league/league-context-store';

@Component({
  selector: 'app-league-subnav',
  imports: [RouterLink, RouterLinkActive, TuiItem, TuiTabs],
  templateUrl: './league-subnav.html',
  styleUrl: './league-subnav.css',
})
export class LeagueSubnav {
  readonly leagueId = input.required<string>();

  private readonly context = inject(LeagueContextStore);

  protected readonly isStaff = computed(() => this.context.isOwner() || this.context.isModerator());

  protected readonly tabs = [
    { label: 'Dashboard', segment: 'dashboard' },
    { label: 'Team Sheets', segment: 'teamsheets' },
    { label: 'Draftboard', segment: 'draftboard' },
    { label: 'Draft History', segment: 'draft-history' },
    { label: 'Games', segment: 'games' },
    { label: 'Standings', segment: 'standings' },
    { label: 'Transfers', segment: 'transfers' },
  ];
}
