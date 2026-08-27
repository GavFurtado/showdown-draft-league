import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { TuiButton, TuiIcon } from '@taiga-ui/core';
import { TuiNavigation } from '@taiga-ui/layout';

import { environment } from '../../../environments/environment';
import { AuthService } from '../auth/auth-service';
import { ThemeService } from '../theme/theme-service';
import { DevUserSwitcher } from './dev-user-switcher';

@Component({
  selector: 'app-shell',
  imports: [RouterLink, RouterLinkActive, RouterOutlet, TuiButton, TuiIcon, ...TuiNavigation, DevUserSwitcher],
  templateUrl: './shell.html',
  styleUrl: './shell.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class Shell {
  private readonly auth = inject(AuthService);
  private readonly router = inject(Router);
  protected readonly theme = inject(ThemeService);
  protected readonly devTools = environment.devTools;

  protected readonly user = this.auth.user;

  protected logout(): void {
    this.auth.logout();
    void this.router.navigateByUrl('/');
  }
}
