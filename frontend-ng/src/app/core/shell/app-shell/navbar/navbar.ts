import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { TuiButton, TuiIcon } from '@taiga-ui/core';
import { TuiNavigation } from '@taiga-ui/layout';

import { LeagueDropdown } from '../league-dropdown/league-dropdown';
import { ProfileMenu } from '../profile-menu/profile-menu';

@Component({
  selector: 'app-navbar',
  imports: [
    RouterLink,
    RouterLinkActive,
    TuiButton,
    TuiIcon,
    ...TuiNavigation,
    LeagueDropdown,
    ProfileMenu,
  ],
  templateUrl: './navbar.html',
  styleUrl: './navbar.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class Navbar {}