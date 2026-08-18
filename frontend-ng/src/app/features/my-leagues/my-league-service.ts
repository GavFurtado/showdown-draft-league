import { inject, Injectable } from '@angular/core';
import { ApiService } from '../../core/api/api-service';
import { Observable } from 'rxjs';
import { League } from '../league/models/league.model';
import { routePaths } from '../../core/api/route-paths';

@Injectable({ providedIn: 'root' })
export class MyLeagueService {
  private readonly api = inject(ApiService);

  getMyLeagues(): Observable<League[]> {
    return this.api.get<League[]>(routePaths.users.myLeagues);
  }
}
