import { ComponentFixture, TestBed } from '@angular/core/testing';

import { LeagueSubnav } from './league-subnav';

describe('LeagueSubnav', () => {
  let component: LeagueSubnav;
  let fixture: ComponentFixture<LeagueSubnav>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LeagueSubnav],
    }).compileComponents();

    fixture = TestBed.createComponent(LeagueSubnav);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
