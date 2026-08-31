import { ComponentFixture, TestBed } from '@angular/core/testing';

import { LeagueDashboard } from './dashboard';
import { LeagueContextStore } from '../league-context-store';

describe('LeagueDashboard', () => {
  let component: LeagueDashboard;
  let fixture: ComponentFixture<LeagueDashboard>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LeagueDashboard],
      providers: [LeagueContextStore],
    }).compileComponents();

    fixture = TestBed.createComponent(LeagueDashboard);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
