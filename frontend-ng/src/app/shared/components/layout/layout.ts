import { ChangeDetectionStrategy, Component, input } from '@angular/core';

@Component({
  selector: 'app-layout',
  imports: [],
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './layout.html',
  styleUrl: './layout.css',
})
export class Layout {
  readonly variant = input<'container' | 'full'>('container');
}
