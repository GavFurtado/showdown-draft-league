import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PublicLeaguesList } from './public-leagues-list';

describe('PublicLeaguesList', () => {
  let component: PublicLeaguesList;
  let fixture: ComponentFixture<PublicLeaguesList>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PublicLeaguesList],
    }).compileComponents();

    fixture = TestBed.createComponent(PublicLeaguesList);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
