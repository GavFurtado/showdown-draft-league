import { ChangeDetectionStrategy, Component, computed, inject } from '@angular/core';
import { toSignal } from '@angular/core/rxjs-interop';
import { NavigationEnd, Router, RouterLink } from '@angular/router';
import { TuiButton, TuiDataList, TuiDropdown, TuiIcon, TuiOption } from '@taiga-ui/core';
import { filter, map } from 'rxjs';

import { UserLeaguesStore } from '../user-leagues-store';

@Component({
  selector: 'app-league-dropdown',
  imports: [RouterLink, TuiButton, TuiDataList, TuiDropdown, TuiIcon, TuiOption],
  templateUrl: './league-dropdown.html',
  styleUrl: './league-dropdown.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class LeagueDropdown {
  private readonly router = inject(Router);
  protected readonly store = inject(UserLeaguesStore);

  protected open = false;

  private readonly url = toSignal(
    this.router.events.pipe(
      filter((e): e is NavigationEnd => e instanceof NavigationEnd),
      map(() => this.router.url),
    ),
    { initialValue: this.router.url },
  );

  protected readonly currentLeagueId = computed(
    () => this.url().match(/\/leagues\/([^/?#]+)/)?.[1] ?? null,
  );

  protected readonly currentLeague = computed(
    () => this.store.leagues().find((league) => league.ID === this.currentLeagueId()) ?? null,
  );

  protected goTo(id: string): void {
    this.open = false;
    if (id !== this.currentLeagueId()) {
      this.router.navigate(['/leagues', id, 'dashboard']);
    }
  }
}