import { Component, effect, inject } from '@angular/core';
import { TuiNotificationService } from '@taiga-ui/core';

import { ClientError } from '../../api/api.model';
import { ErrorService } from '../error-service';

@Component({
  selector: 'app-error-notifications',
  imports: [],
  templateUrl: './error-notifications.html',
  styleUrl: './error-notifications.css',
})
export class ErrorNotifications {
  private readonly errors = inject(ErrorService);
  private readonly notifications = inject(TuiNotificationService);

  // Track how many errors we've already surfaced so we only toast the newly-appended
  // tail of the (deduped, capped) errors signal rather than re-toasting on every change.
  private lastReported = 0;

  constructor() {
    effect(() => {
      const list = this.errors.errors();
      for (let i = this.lastReported; i < list.length; i++) {
        this.show(list[i]);
      }
      this.lastReported = list.length;
    });
  }

  private show(error: ClientError): void {
    this.notifications
      .open(error.message, {
        label: 'Something went wrong',
        appearance: 'negative',
        autoClose: 10000,
        closable: true,
      })
      .subscribe();
  }
}
