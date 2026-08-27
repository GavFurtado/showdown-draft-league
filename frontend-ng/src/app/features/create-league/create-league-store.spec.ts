import { TestBed } from '@angular/core/testing';

import { CreateLeagueStore } from './create-league-store';

describe('CreateLeagueStore', () => {
  let service: CreateLeagueStore;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(CreateLeagueStore);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
