import { Component, inject } from '@angular/core';
import { RouterOutlet } from '@angular/router';

import { LeagueContextStore } from '../league-context-store';

@Component({
  selector: 'app-league-shell',
  imports: [RouterOutlet],
  templateUrl: './league-shell.html',
  styleUrl: './league-shell.less',
})
export class LeagueShell {
  protected readonly context = inject(LeagueContextStore);
}
