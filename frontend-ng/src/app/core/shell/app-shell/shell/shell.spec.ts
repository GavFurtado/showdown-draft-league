import { provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { provideTaiga } from '@taiga-ui/core';

import { UserLeaguesStore } from './user-leagues-store';
import { Shell } from './shell';

describe('Shell', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Shell],
      providers: [
        provideRouter([]),
        provideTaiga(),
        provideHttpClientTesting(),
        { provide: UserLeaguesStore, useValue: { leagues: [], loading: false, error: null, refetch: () => undefined } },
      ],
    }).compileComponents();
  });

  it('should create', () => {
    const fixture = TestBed.createComponent(Shell);
    expect(fixture.componentInstance).toBeTruthy();
  });
});