import { Component, inject } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-league-detail',
  imports: [],
  template: `
    <section class="mx-auto flex w-full max-w-3xl flex-col gap-4">
      <h1 class="text-2xl font-semibold">League {{ leagueId }}</h1>
      <p class="text-sm text-(--tui-text-secondary)">League details are coming soon.</p>
    </section>
  `,
})
export class LeagueDetail {
  private readonly route = inject(ActivatedRoute);

  protected readonly leagueId = this.route.snapshot.paramMap.get('id');
}
