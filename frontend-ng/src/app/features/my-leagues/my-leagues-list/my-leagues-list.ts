import { Component, input } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TuiButton, TuiIcon } from '@taiga-ui/core';
import { LeagueCard } from '../league-card/league-card';
import { League } from '../../league/models/league.model';
import { ClientError } from '../../../core/api/api.model';

@Component({
  selector: 'app-my-leagues-list',
  imports: [RouterLink, TuiButton, TuiIcon, LeagueCard],
  templateUrl: './my-leagues-list.html',
  styleUrl: './my-leagues-list.css',
})
export class MyLeaguesList {
  leagues = input.required<League[]>();
  loading = input(false);
  error = input<ClientError | null>(null);
}
