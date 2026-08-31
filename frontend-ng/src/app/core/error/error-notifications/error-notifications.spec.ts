import { ComponentFixture, TestBed } from '@angular/core/testing';
import { provideTaiga } from '@taiga-ui/core';

import { ErrorNotifications } from './error-notifications';

describe('ErrorNotifications', () => {
  let component: ErrorNotifications;
  let fixture: ComponentFixture<ErrorNotifications>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ErrorNotifications],
      providers: [provideTaiga()],
    }).compileComponents();

    fixture = TestBed.createComponent(ErrorNotifications);
    component = fixture.componentInstance;
    await fixture.whenStable();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
