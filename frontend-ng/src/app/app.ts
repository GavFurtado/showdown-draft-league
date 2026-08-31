import { TuiRoot } from '@taiga-ui/core';
import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';

import { ErrorNotifications } from './core/error/error-notifications/error-notifications';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, TuiRoot, ErrorNotifications],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App {}
