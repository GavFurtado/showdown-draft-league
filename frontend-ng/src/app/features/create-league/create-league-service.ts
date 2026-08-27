import { inject, Service } from '@angular/core';
import { ApiService } from '../../core/api/api-service';
import { LeagueCreateRequest } from './models/league-create-request.model';
import { Observable } from 'rxjs';
import { League } from '../league/models/league.model';
import { routePaths } from '../../core/api/route-paths';

@Service()
export class CreateLeagueService {
  private readonly api = inject(ApiService);

  createLeague(request: LeagueCreateRequest): Observable<League> {
    return this.api.post<League>(routePaths.leagues.base, request);
  }
}
