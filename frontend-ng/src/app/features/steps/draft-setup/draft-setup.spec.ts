import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DraftSetup } from './draft-setup';
import { createTestLeagueForm } from '../../../shared/testing/test-league-form';

describe('DraftSetup', () => {
  let component: DraftSetup;
  let fixture: ComponentFixture<DraftSetup>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [DraftSetup],
    }).compileComponents();

    fixture = TestBed.createComponent(DraftSetup);
    fixture.componentRef.setInput('form', createTestLeagueForm());
    fixture.componentRef.setInput('fieldError', () => null);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
