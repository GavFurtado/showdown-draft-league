import { provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideTaiga } from '@taiga-ui/core';

import { ProfileMenu } from './profile-menu';

describe('ProfileMenu', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ProfileMenu],
      providers: [provideTaiga(), provideHttpClientTesting()],
    }).compileComponents();
  });

  it('should create', () => {
    const fixture = TestBed.createComponent(ProfileMenu);
    expect(fixture.componentInstance).toBeTruthy();
  });
});
