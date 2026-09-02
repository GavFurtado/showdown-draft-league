import { TestBed } from '@angular/core/testing';

import { CreateLeagueService } from './create-league-service';

describe('CreateLeagueService', () => {
  let service: CreateLeagueService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(CreateLeagueService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });
});
