import { Component, inject, input } from '@angular/core';
import { RouterOutlet } from '@angular/router';

import { LeagueSubnav } from '../league-subnav/league-subnav';
import { LeagueContextStore } from '../../../../features/league/league-context-store';

@Component({
  selector: 'app-league-shell',
  imports: [RouterOutlet, LeagueSubnav],
  templateUrl: './league-shell.html',
  styleUrl: './league-shell.css',
})
export class LeagueShell {
  readonly leagueId = input.required<string>();
  protected readonly context = inject(LeagueContextStore);
}
