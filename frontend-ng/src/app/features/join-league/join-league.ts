import { ChangeDetectionStrategy, Component, OnInit, computed, effect, inject, signal } from '@angular/core';
import { DatePipe } from '@angular/common';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { TuiNotificationService, TuiButton, TuiCell, TuiIcon, TuiInput, TuiTextfield, TuiTitle } from '@taiga-ui/core';
import { TuiAccordion, TuiBadge, TuiToastService, TuiTooltip } from '@taiga-ui/kit';
import { TuiAutoFocus } from '@taiga-ui/cdk';
import { Subscription } from 'rxjs';

import { AuthService } from '../../core/auth/auth-service';
import { asUuid, type UUID } from '../../shared/types/branded-strings';
import { LeagueStatus } from '../league/models/enums/league-status';
import { MemberRole } from '../league/models/enums/member-role';
import { LeagueMember } from '../league/models/league-member.model';
import { Layout } from '../../shared/components/layout/layout';
import { StatusBadge } from '../../shared/components/status-badge/status-badge';
import { EnumLabelPipe } from '../../shared/pipes/enum-label.pipe';
import { forbiddenChars, htmlRejection } from '../../shared/validation/text-validators';
import { JoinLeagueStore } from './join-league-store';

const REDIRECT_DELAY_MS = 5000;

const ROLE_RANK: Record<MemberRole, number> = {
  [MemberRole.OWNER]: 0,
  [MemberRole.MODERATOR]: 1,
  [MemberRole.MEMBER]: 2,
};

@Component({
  selector: 'app-join-league',
  imports: [
    Layout,
    RouterLink,
    ReactiveFormsModule,
    StatusBadge,
    EnumLabelPipe,
    DatePipe,
    TuiButton,
    TuiCell,
    TuiIcon,
    TuiInput,
    TuiTextfield,
    TuiTitle,
    TuiAccordion,
    TuiAutoFocus,
    TuiBadge,
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
  private readonly router = inject(Router);
  private readonly toast = inject(TuiToastService);
  private readonly notifications = inject(TuiNotificationService);

  // The :leagueId segment belongs to a parent route (invite links land on /leagues/:leagueId/join),
  protected readonly leagueId = this.readLeagueId();

  private joiningToast: Subscription | null = null;
  private successNotified = false;
  private redirectScheduled = false;

  constructor() {
    // Drive the in-transit / result toasts from the store's join lifecycle.
    effect(() => {
      if (this.store.joining()) {
        this.successNotified = false;
        this.joiningToast ??= this.toast.open('Joining league…', { autoClose: 0, closable: false }).subscribe();
      } else {
        this.joiningToast?.unsubscribe();
        this.joiningToast = null;
        if (this.store.joinedMember() && !this.successNotified) {
          this.successNotified = true;
          this.notifications
            .open('You’re in the league now.', {
              label: 'Joined',
              appearance: 'positive',
              autoClose: 3000,
            })
            .subscribe();
        }
      }
    });

    // Post-join landing zone: once we know the user is in the league, show the big
    // confirmation, tick a live countdown, and auto-head to the dashboard after a few seconds.
    effect(() => {
      if (!this.success() || this.redirectScheduled) return;
      this.redirectScheduled = true;
      let remaining = REDIRECT_DELAY_MS / 1000;
      this.countdown.set(remaining);
      const interval = setInterval(() => {
        remaining -= 1;
        this.countdown.set(Math.max(remaining, 0));
      }, 1000);
      setTimeout(() => {
        clearInterval(interval);
        void this.router.navigate(['/leagues', this.leagueId, 'dashboard']);
      }, REDIRECT_DELAY_MS);
    });
  }

  protected readonly form = new FormGroup({
    InLeagueName: new FormControl('', {
      validators: [Validators.minLength(3), Validators.maxLength(20), htmlRejection(), forbiddenChars()],
    }),
    TeamName: new FormControl('', {
      validators: [Validators.minLength(3), Validators.maxLength(20), htmlRejection(), forbiddenChars()],
    }),
  });

  protected readonly league = this.store.league;

  protected readonly memberRole = MemberRole;

  protected readonly alreadyMember = computed(() => {
    const league = this.store.league();
    const userId = this.auth.user()?.ID;
    return !!league && !!userId && !!league.Members?.some((m) => m.UserID === userId);
  });

  // Post-join landing: the current session is in the league. Either they just joined
  // in this view, or they're re-visiting an invite they're already part of.
  protected readonly isNewlyJoined = computed(() => !!this.store.joinedMember());
  protected readonly success = computed(() => this.isNewlyJoined() || this.alreadyMember());

  // Pre-join walkthrough: first show the league details, then the (optional) name form.
  protected readonly stage = signal<'details' | 'form'>('details');

  // Post-join redirect countdown (ticks down each second once success shows).
  protected readonly countdown = signal(REDIRECT_DELAY_MS / 1000);

  // Joining is only allowed while the league is in setup.
  protected readonly joinable = computed(() => {
    const status = this.store.league()?.Status;
    return status === LeagueStatus.SETUP;
  });

  protected readonly full = computed(() => {
    const league = this.store.league();
    return !!league && league.PlayerCount >= league.MaxPlayers;
  });

  protected readonly owner = computed(() => {
    const members = this.store.league()?.Members ?? [];
    const ownerId = this.store.league()?.OwnerUserID;
    return (
      members.find((m) => m.Role === MemberRole.OWNER || m.UserID === ownerId) ??
      members.find((m) => m.UserID === ownerId) ??
      null
    );
  });

  // Staff = every non-MEMBER role, owner on top.
  protected readonly staff = computed(() =>
    (this.store.league()?.Members ?? [])
      .filter((m) => m.Role !== MemberRole.MEMBER)
      .sort((a, b) => ROLE_RANK[a.Role] - ROLE_RANK[b.Role]),
  );

  // Primary display name: in-league name when set, otherwise the user's Discord username.
  protected memberLabel(member: LeagueMember): string {
    return member.InLeagueName || member.User?.DiscordUsername || 'Unknown';
  }

  // Secondary identity (shown when it differs from the in-league name): the Discord username.
  protected memberMeta(member: LeagueMember): string | null {
    if (member.InLeagueName && member.User?.DiscordUsername) return member.User.DiscordUsername;
    return null;
  }

  protected readonly discordUsername = computed(() => this.auth.user()?.DiscordUsername ?? '');

  protected readonly teamPlaceholder = computed(() => `${this.discordUsername()}'s Team`);

  ngOnInit(): void {
    if (this.leagueId) this.store.load(this.leagueId);
  }

  protected continueToForm(): void {
    this.stage.set('form');
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

  private readLeagueId(): UUID | null {
    let snapshot = this.route.snapshot;
    while (snapshot) {
      const id = snapshot.paramMap.get('leagueId');
      if (id) return asUuid(id);
      snapshot = snapshot.parent!;
    }
    return null;
  }
}
