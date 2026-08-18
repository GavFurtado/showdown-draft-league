import { ChangeDetectionStrategy, Component, inject } from '@angular/core';
import { RouterLink } from '@angular/router';
import { TuiButton } from '@taiga-ui/core/components/button';
import { AuthService } from '../../core/auth/auth-service';

@Component({
  selector: 'app-landing',
  imports: [RouterLink, TuiButton],
  templateUrl: './landing.html',
  styleUrl: './landing.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class Landing {
  protected readonly auth = inject(AuthService);

  protected ctaButtonText(): string {
    return this.auth.isLoggedIn() ? 'Proceed to Leagues' : 'Get Started';
  }
}
