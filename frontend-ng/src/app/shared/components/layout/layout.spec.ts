import { TestBed } from '@angular/core/testing';
import { Layout } from './layout';

describe('Layout', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Layout],
    }).compileComponents();
  });

  it('should create', () => {
    const fixture = TestBed.createComponent(Layout);
    expect(fixture.componentInstance).toBeTruthy();
  });

  it('should render container variant', () => {
    const fixture = TestBed.createComponent(Layout);
    fixture.componentRef.setInput('variant', 'container');
    fixture.detectChanges();
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('div.mx-auto')).toBeTruthy();
  });

  it('should render full variant without wrapper', () => {
    const fixture = TestBed.createComponent(Layout);
    fixture.componentRef.setInput('variant', 'full');
    fixture.detectChanges();
    const el = fixture.nativeElement as HTMLElement;
    expect(el.querySelector('div.mx-auto')).toBeFalsy();
  });
});
