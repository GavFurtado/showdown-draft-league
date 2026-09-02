import { Component, inject, input } from '@angular/core';

import { LeagueContextStore } from '../league-context-store';

@Component({
  selector: 'app-league-dashboard',
  imports: [],
  templateUrl: './dashboard.html',
  styleUrl: './dashboard.less',
})
export class LeagueDashboard {
  protected readonly context = inject(LeagueContextStore);
  protected readonly leagueId = input<string | undefined>();
}
