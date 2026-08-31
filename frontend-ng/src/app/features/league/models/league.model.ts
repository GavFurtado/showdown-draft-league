import { UUID, ISODateTimeString } from '../../../shared/types/branded-strings';
import { User } from '../../../shared/models/user.model';
import { LeagueStatus } from './enums/league-status';
import { LeagueVisibility } from './enums/league-visibility';
import { LeagueFormat } from './league-format.model';
import { LeagueMember } from './league-member.model';

export interface League {
  ID: UUID;
  Name: string;
  RulesetDescription: string;
  PlayerCount: number;
  MaxPlayers: number;
  Status: LeagueStatus;
  Visibility: LeagueVisibility;
  MaxPokemonPerPlayer: number;
  MinPokemonPerPlayer: number;
  CurrentWeekNumber: number;
  StartDate: ISODateTimeString | null;
  EndDate: ISODateTimeString | null;
  RegularSeasonStartDate: ISODateTimeString | null;
  OwnerUserID: UUID;
  DiscordWebhookURL: string;
  StartingDraftPoints: number;
  NextWeeklyTick: ISODateTimeString | null;
  NewPlayerGroupNumber: number;
  Format: LeagueFormat;
  // CreatedAt: string;
  // UpdatedAt: string;

  OwnerUser: User;
  Members: LeagueMember[];
}
