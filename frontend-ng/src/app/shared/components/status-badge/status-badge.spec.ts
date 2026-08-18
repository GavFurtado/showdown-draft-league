import { ComponentFixture, TestBed } from '@angular/core/testing';

import { StatusBadge } from './status-badge';

describe('StatusBadge', () => {
  let fixture: ComponentFixture<StatusBadge>;

  const badge = (): HTMLElement => fixture.nativeElement.querySelector('[tuiBadge]') as HTMLElement;

  const render = (props: { status?: string; domain?: string; label?: string; size?: string } = {}): void => {
    fixture.componentRef.setInput('status', props.status ?? '');
    if (props.domain) fixture.componentRef.setInput('domain', props.domain);
    if (props.label) fixture.componentRef.setInput('label', props.label);
    if (props.size) fixture.componentRef.setInput('size', props.size);
    fixture.detectChanges();
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [StatusBadge],
    }).compileComponents();

    fixture = TestBed.createComponent(StatusBadge);
  });

  it('renders a prettified label and Tailwind tone classes for a known league status', () => {
    render({ status: 'TRANSFER_WINDOW', domain: 'league' });

    expect(badge().textContent?.trim()).toBe('Transfer Window');
    expect(badge().getAttribute('data-size')).toBe('m');
    expect(badge().classList).toContain('bg-teal-100!');
    expect(badge().classList).toContain('text-teal-800!');
    expect(badge().classList).toContain('[--t-status:#14b8a6]');
  });

  it('colors COMPLETED per domain: gray for league, green for game', () => {
    render({ status: 'COMPLETED', domain: 'league' });
    expect(badge().classList).toContain('bg-gray-100!');

    render({ status: 'COMPLETED', domain: 'game' });
    expect(badge().classList).toContain('bg-green-100!');
  });

  it('falls back to neutral gray with the raw status for unknown values', () => {
    render({ status: 'FOO' });

    expect(badge().textContent?.trim()).toBe('FOO');
    expect(badge().classList).toContain('bg-gray-100!');
  });

  it('allows an explicit label override', () => {
    render({ status: 'COMPLETED', domain: 'league', label: 'Done!' });

    expect(badge().textContent?.trim()).toBe('Done!');
  });

  it('normalizes lowercase claim sources', () => {
    render({ status: 'free_agent', domain: 'claim' });

    expect(badge().textContent?.trim()).toBe('Free Agent');
  });

  it('renders an empty label when status is empty', () => {
    render();

    expect(badge().textContent?.trim()).toBe('');
  });

  it('binds the badge size', () => {
    render({ status: 'SETUP', domain: 'league', size: 's' });

    expect(badge().getAttribute('data-size')).toBe('s');
  });
});
