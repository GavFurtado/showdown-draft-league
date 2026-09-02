import { UUID, ISODateTimeString } from '../../../shared/types/branded-strings';
import { DraftStatus } from './enums/draft-status';
import { LeagueMember } from './league-member.model';
import { League } from './league.model';

export interface Draft {
  ID: UUID;
  LeagueID: UUID;

  CurrentPickOnClock: number;
  CurrentPickInRound: number;
  CurrentRound: number;

  CurrentTurnPlayerID: UUID | null;
  CurrentTurnStartTime: ISODateTimeString;
  CurrentTurnMember?: LeagueMember;
  League?: League;
  PlayersWithAccumulatedPicks: PlayerAccumulatedPicks;

  StartTime: ISODateTimeString;
  EndTime: ISODateTimeString;

  Status: DraftStatus;
  TurnTimeLimit: number;
  // CreatedAt: string;
  // UpdatedAt: string;
}

export type PlayerAccumulatedPicks = Record<UUID, number[]>;
