import { ComponentFixture, TestBed } from '@angular/core/testing';

import { SeasonSetup } from './season-setup';
import { createTestLeagueForm } from '../../../shared/testing/test-league-form';

describe('SeasonSetup', () => {
  let component: SeasonSetup;
  let fixture: ComponentFixture<SeasonSetup>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [SeasonSetup],
    }).compileComponents();

    fixture = TestBed.createComponent(SeasonSetup);
    fixture.componentRef.setInput('form', createTestLeagueForm());
    fixture.componentRef.setInput('fieldError', () => null);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
