import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { LeagueCard } from './league-card';
import { MemberRole } from '../../league/models/enums/member-role';

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
      ID: '11111111-1111-4111-8111-111111111111',
      Name: 'Test League',
      RulesetDescription: 'A test league',
      PlayerCount: 4,
      MaxPlayers: 16,
      Status: 'DRAFTING',
      Format: { SeasonType: 'HYBRID' },
      Members: [
        { ID: '66666666-6666-4666-8666-666666666666', InLeagueName: 'Alice', Role: MemberRole.OWNER, IsActive: true },
        { ID: '77777777-7777-4777-8777-777777777777', InLeagueName: 'Bob', Role: MemberRole.MEMBER, IsActive: true },
      ],
    });
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
