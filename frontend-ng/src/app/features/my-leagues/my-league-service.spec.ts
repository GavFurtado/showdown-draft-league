import { TestBed } from '@angular/core/testing';

import { MyLeagueService } from './my-league-service';

describe('MyLeagueService', () => {
  let service: MyLeagueService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(MyLeagueService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
