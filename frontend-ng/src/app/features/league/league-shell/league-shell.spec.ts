import { ComponentFixture, TestBed } from '@angular/core/testing';

import { LeagueShell } from './league-shell';
import { LeagueContextStore } from '../league-context-store';

describe('LeagueShell', () => {
  let component: LeagueShell;
  let fixture: ComponentFixture<LeagueShell>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LeagueShell],
      providers: [LeagueContextStore],
    }).compileComponents();

    fixture = TestBed.createComponent(LeagueShell);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
