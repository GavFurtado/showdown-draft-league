import { ChangeDetectionStrategy, Component, OnInit, computed, inject } from '@angular/core';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';

import { AuthService } from '../../core/auth/auth-service';
import { LeagueStatus } from '../league/models/enums/league-status';
import { Layout } from '../../shared/components/layout/layout';
import { StatusBadge } from '../../shared/components/status-badge/status-badge';
import { EnumLabelPipe } from '../../shared/pipes/enum-label.pipe';
import { forbiddenChars, htmlRejection } from '../../shared/validation/text-validators';
import { JoinLeagueStore } from './join-league-store';
import { TuiButton, TuiIcon, TuiInput, TuiTextfield } from '@taiga-ui/core';
import { TuiCardLarge, TuiCardRow, TuiHeader } from '@taiga-ui/layout';
import { TuiTitle } from '@taiga-ui/core';
import { TuiTooltip } from '@taiga-ui/kit';

@Component({
  selector: 'app-join-league',
  imports: [
    Layout,
    RouterLink,
    ReactiveFormsModule,
    StatusBadge,
    EnumLabelPipe,
    TuiButton,
    TuiIcon,
    TuiInput,
    TuiTextfield,
    TuiCardLarge,
    TuiCardRow,
    TuiHeader,
    TuiTitle,
    TuiTooltip,
  ],
  templateUrl: './join-league.html',
  styleUrl: './join-league.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class JoinLeague implements OnInit {
  protected readonly store = inject(JoinLeagueStore);
  private readonly auth = inject(AuthService);
  private readonly route = inject(ActivatedRoute);

  // The :leagueId segment belongs to a parent route (invite links land on /:leagueId/join),
  // so walk the tree instead of assuming which level owns the param.
  protected readonly leagueId = this.readLeagueId();

  protected readonly form = new FormGroup({
    InLeagueName: new FormControl('', {
      validators: [Validators.minLength(3), Validators.maxLength(20), htmlRejection(), forbiddenChars()],
    }),
    TeamName: new FormControl('', {
      validators: [Validators.minLength(3), Validators.maxLength(20), htmlRejection(), forbiddenChars()],
    }),
  });

  protected readonly league = this.store.league;

  protected readonly alreadyMember = computed(() => {
    const league = this.store.league();
    const userId = this.auth.user()?.ID;
    return !!league && !!userId && !!league.Members?.some((m) => m.UserID === userId);
  });

  // The backend rejects joins for leagues outside SETUP (401) — mirror that client-side.
  protected readonly joinable = computed(() => this.store.league()?.Status === LeagueStatus.SETUP);

  protected readonly full = computed(() => {
    const league = this.store.league();
    return !!league && league.PlayerCount >= league.MaxPlayers;
  });

  protected readonly ownerName = computed(() => {
    const members = this.store.league()?.Members ?? [];
    const ownerId = this.store.league()?.OwnerUserID;
    return (
      members.find((m) => m.Role === 'OWNER' || m.UserID === ownerId)?.DisplayName ??
      members.find((m) => m.UserID === ownerId)?.User?.DiscordUsername ??
      null
    );
  });

  protected readonly discordUsername = computed(() => this.auth.user()?.DiscordUsername ?? '');

  protected readonly teamPlaceholder = computed(() => `${this.discordUsername()}'s Team`);

  ngOnInit(): void {
    if (this.leagueId) this.store.load(this.leagueId);
  }

  protected fieldError(name: 'InLeagueName' | 'TeamName'): string | null {
    const control = this.form.get(name)!;
    if (!control.invalid || !control.touched) return null;

    if (control.hasError('required')) return 'This field is required.';
    if (control.hasError('minlength'))
      return `Must be at least ${control.getError('minlength').requiredLength} characters.`;
    if (control.hasError('maxlength'))
      return `Must be at most ${control.getError('maxlength').requiredLength} characters.`;
    if (control.hasError('containsHtml')) return 'HTML tags are not allowed.';
    if (control.hasError('forbiddenChars')) return "The '%' and '\\' characters are not allowed.";
    return null;
  }

  protected onSubmit(): void {
    this.form.markAllAsTouched();
    const userId = this.auth.user()?.ID;
    if (this.form.invalid || !this.leagueId || !userId) return;

    const inLeagueName = this.form.getRawValue().InLeagueName?.trim() || undefined;
    const teamName = this.form.getRawValue().TeamName?.trim() || undefined;

    this.store.join({ UserID: userId, LeagueID: this.leagueId, InLeagueName: inLeagueName, TeamName: teamName });
  }

  private readLeagueId(): string | null {
    let snapshot = this.route.snapshot;
    while (snapshot) {
      const id = snapshot.paramMap.get('leagueId');
      if (id) return id;
      snapshot = snapshot.parent!;
    }
    return null;
  }
}
