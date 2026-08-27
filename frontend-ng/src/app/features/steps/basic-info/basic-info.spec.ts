import { ComponentFixture, TestBed } from '@angular/core/testing';

import { BasicInfo } from './basic-info';
import { createTestLeagueForm } from '../../../shared/testing/test-league-form';

describe('BasicInfo', () => {
  let component: BasicInfo;
  let fixture: ComponentFixture<BasicInfo>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [BasicInfo],
    }).compileComponents();

    fixture = TestBed.createComponent(BasicInfo);
    fixture.componentRef.setInput('form', createTestLeagueForm());
    fixture.componentRef.setInput('fieldError', () => null);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
