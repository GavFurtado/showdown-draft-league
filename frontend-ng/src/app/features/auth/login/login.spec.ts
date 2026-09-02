import { provideHttpClientTesting } from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { provideTaiga } from '@taiga-ui/core';

import { AuthService, LoginError } from '../../../core/auth/auth-service';
import { Login } from './login';

describe('Login', () => {
  let fixture: ComponentFixture<Login>;
  let auth: AuthService;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Login],
      providers: [provideRouter([]), provideTaiga(), provideHttpClientTesting()],
    }).compileComponents();
    auth = TestBed.inject(AuthService);
    fixture = TestBed.createComponent(Login);
  });

  it('renders the sign-in page', () => {
    expect(fixture.nativeElement.querySelector('h1').textContent).toContain('Sign in');
  });

  it('logs in via the popup and navigates to /my-leagues', async () => {
    vi.spyOn(auth, 'login').mockResolvedValue();
    const router = TestBed.inject(Router);
    const nav = vi.spyOn(router, 'navigateByUrl').mockResolvedValue(true);

    fixture.nativeElement.querySelector('button').click();
    await fixture.whenStable();

    expect(auth.login).toHaveBeenCalled();
    expect(nav).toHaveBeenCalledWith('/my-leagues');
  });

  it('falls back to full-page navigation when the popup is blocked', async () => {
    vi.spyOn(auth, 'login').mockRejectedValue(new LoginError('popup-blocked'));
    const button = fixture.nativeElement.querySelector('button');

    button.click();
    await fixture.whenStable();

    // jsdom cannot actually navigate (location.href is non-configurable);
    // assert the fallback path was taken: state stays 'logging-in', no error surfaced.
    expect(auth.login).toHaveBeenCalled();
    expect(button.disabled).toBe(true);
    expect(fixture.nativeElement.textContent).not.toContain('Login window was closed');
  });

  it('shows an error when the popup is closed without completing', async () => {
    vi.spyOn(auth, 'login').mockRejectedValue(new LoginError('popup-closed'));
    const button = fixture.nativeElement.querySelector('button');

    button.click();
    await fixture.whenStable();

    expect(fixture.nativeElement.textContent).toContain('Login window was closed');
    expect(button.disabled).toBe(false);
  });
});
