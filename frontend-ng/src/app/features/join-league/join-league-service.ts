import { Service, inject } from '@angular/core';
import type { Observable } from 'rxjs';

import { ApiService } from '../../core/api/api-service';
import { routePaths } from '../../core/api/route-paths';
import { League } from '../league/models/league.model';
import { LeagueMember } from '../league/models/league-member.model';

// Wire DTO for POST /api/leagues/:leagueId/members/join. InLeagueName/TeamName are
// optional on the wire (Go *string, omitempty validation 3–20 chars when present):
// the server defaults them to DiscordUsername / "<username>'s Team" when omitted.
export interface LeagueJoinRequest {
  UserID: string;
  LeagueID: string;
  InLeagueName?: string;
  TeamName?: string;
}

@Service()
export class JoinLeagueService {
  private readonly api = inject(ApiService);

  getLeague(leagueId: string): Observable<League> {
    return this.api.get<League>(routePaths.leagues.byId(leagueId));
  }

  join(request: LeagueJoinRequest): Observable<LeagueMember> {
    return this.api.post<LeagueMember>(routePaths.leagues.members.join(request.LeagueID), request, {
      // The join mutates league membership; stale GETs must refetch.
      invalidatePending: [routePaths.leagues.byId(request.LeagueID), routePaths.users.myLeagues],
    });
  }
}
