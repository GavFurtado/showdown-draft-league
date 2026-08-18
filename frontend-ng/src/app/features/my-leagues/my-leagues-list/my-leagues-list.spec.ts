import { ComponentFixture, TestBed } from '@angular/core/testing';

import { MyLeaguesList } from './my-leagues-list';

describe('MyLeaguesList', () => {
  let component: MyLeaguesList;
  let fixture: ComponentFixture<MyLeaguesList>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MyLeaguesList],
    }).compileComponents();

    fixture = TestBed.createComponent(MyLeaguesList);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
