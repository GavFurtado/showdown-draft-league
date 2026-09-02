import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TransferMarket } from './transfer-market';
import { createTestLeagueForm } from '../../../shared/testing/test-league-form';

describe('TransferMarket', () => {
  let component: TransferMarket;
  let fixture: ComponentFixture<TransferMarket>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [TransferMarket],
    }).compileComponents();

    fixture = TestBed.createComponent(TransferMarket);
    fixture.componentRef.setInput('form', createTestLeagueForm());
    fixture.componentRef.setInput('fieldError', () => null);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
