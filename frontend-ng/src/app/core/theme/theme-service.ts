import { Injectable, signal } from '@angular/core';

const STORAGE_KEY = 'tuiDark';

@Injectable({ providedIn: 'root' })
export class ThemeService {
  readonly isDark = signal(this.getInitial());

  constructor() {
    this.apply(this.isDark());
  }

  toggle(): void {
    this.isDark.update((d) => !d);
    this.apply(this.isDark());
    localStorage.setItem(STORAGE_KEY, String(this.isDark()));
  }

  private getInitial(): boolean {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored !== null) return stored === 'true';
    return matchMedia('(prefers-color-scheme: dark)').matches;
  }

  private apply(dark: boolean): void {
    document.body.setAttribute('tuiTheme', dark ? 'dark' : 'light');
  }
}
