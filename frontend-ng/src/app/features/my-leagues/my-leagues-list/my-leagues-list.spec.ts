import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';

import { MyLeaguesList } from './my-leagues-list';

describe('MyLeaguesList', () => {
  let component: MyLeaguesList;
  let fixture: ComponentFixture<MyLeaguesList>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MyLeaguesList],
      providers: [provideRouter([])],
    }).compileComponents();

    fixture = TestBed.createComponent(MyLeaguesList);
    fixture.componentRef.setInput('leagues', []);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
