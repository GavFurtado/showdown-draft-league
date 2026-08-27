import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { LeagueCard } from './league-card';

describe('LeagueCard', () => {
  let component: LeagueCard;
  let fixture: ComponentFixture<LeagueCard>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [LeagueCard],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(LeagueCard);
    component = fixture.componentInstance;
    fixture.componentRef.setInput('league', {
      ID: 'test-id',
      Name: 'Test League',
      RulesetDescription: 'A test league',
      PlayerCount: 4,
      MaxPlayers: 16,
      Status: 'DRAFTING',
      Format: { SeasonType: 'HYBRID' },
      Members: [
        { ID: 'm1', DisplayName: 'Alice', Role: 'OWNER', IsActive: true },
        { ID: 'm2', DisplayName: 'Bob', Role: 'MEMBER', IsActive: true },
      ],
    });
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
