import { TestBed } from '@angular/core/testing';

import { MyLeagueStore } from './my-league-store';

describe('MyLeagueStore', () => {
  let service: MyLeagueStore;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(MyLeagueStore);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
