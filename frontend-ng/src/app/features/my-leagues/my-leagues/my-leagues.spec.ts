import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { MyLeagues } from './my-leagues';

describe('MyLeagues', () => {
  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MyLeagues],
      providers: [provideRouter([])],
    }).compileComponents();
  });

  it('should create', () => {
    const fixture = TestBed.createComponent(MyLeagues);
    expect(fixture.componentInstance).toBeTruthy();
  });
});
