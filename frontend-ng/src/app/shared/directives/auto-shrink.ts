import {
  Directive,
  ElementRef,
  afterNextRender,
  inject,
  input,
} from '@angular/core';

@Directive({
  selector: '[autoShrink]',
})
export class AutoShrink {
  readonly startSize = input(10);
  readonly minSize = input(7);

  private readonly element = inject(ElementRef<HTMLElement>);

  constructor() {
    afterNextRender(() => {
      const el = this.element.nativeElement;

      const resizeObserver = new ResizeObserver(() => {
        this.shrinkToFit(el);
      });

      resizeObserver.observe(el.parentElement ?? el);

      this.shrinkToFit(el);
    });
  }

  private shrinkToFit(el: HTMLElement) {
    const start = this.startSize();
    const min = this.minSize();

    el.style.fontSize = `${start}px`;

    while (el.scrollWidth > el.clientWidth && parseFloat(el.style.fontSize) > min) {
      el.style.fontSize = `${parseFloat(el.style.fontSize) - 0.5}px`;
    }
  }
}
