import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { TuiIcon } from '@taiga-ui/core';
import { TuiNavigation } from '@taiga-ui/layout';

import { environment } from '../../../environments/environment';
import { DevUserSwitcher } from './dev-user-switcher/dev-user-switcher';
import { Navbar } from './navbar/navbar';

@Component({
  selector: 'app-shell',
  imports: [RouterOutlet, TuiIcon, ...TuiNavigation, DevUserSwitcher, Navbar],
  templateUrl: './shell.html',
  styleUrl: './shell.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class Shell {
  protected readonly devTools = environment.devTools;
}