import { ComponentFixture, TestBed } from '@angular/core/testing';

import { PlayoffSetup } from './playoff-setup';
import { createTestLeagueForm } from '../../../shared/testing/test-league-form';

describe('PlayoffSetup', () => {
  let component: PlayoffSetup;
  let fixture: ComponentFixture<PlayoffSetup>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [PlayoffSetup],
    }).compileComponents();

    fixture = TestBed.createComponent(PlayoffSetup);
    fixture.componentRef.setInput('form', createTestLeagueForm());
    fixture.componentRef.setInput('fieldError', () => null);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
