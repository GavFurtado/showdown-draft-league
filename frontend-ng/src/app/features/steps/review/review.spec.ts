import { ComponentFixture, TestBed } from '@angular/core/testing';

import { Review } from './review';
import { createTestLeagueForm } from '../../../shared/testing/test-league-form';

describe('Review', () => {
  let component: Review;
  let fixture: ComponentFixture<Review>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [Review],
    }).compileComponents();

    fixture = TestBed.createComponent(Review);
    fixture.componentRef.setInput('form', createTestLeagueForm());
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
