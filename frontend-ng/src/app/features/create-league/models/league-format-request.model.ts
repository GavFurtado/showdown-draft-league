import { LeagueDraftOrderType } from '../../league/models/enums/league-draft-order-type';
import { LeaguePlayoffSeedingType } from '../../league/models/enums/league-playoff-seeding-type';
import { LeaguePlayoffType } from '../../league/models/enums/league-playoff-type';
import { LeagueSeasonType } from '../../league/models/enums/league-season-type';

export interface LeagueFormatRequest {
  SeasonType: LeagueSeasonType;
  DraftOrderType: LeagueDraftOrderType;
  GroupCount: number;
  IsSnakeRoundDraft: boolean;
  PlayoffByesCount: number;
  PlayoffParticipantCount: number;
  PlayoffType: LeaguePlayoffType;
  PlayoffSeedingType: LeaguePlayoffSeedingType;
  AllowTransfers: boolean;
  TransfersCostCredits: boolean;
  TransferCreditCap: number;
  DropCost: number;
  PickupCost: number;
  TransferCreditsPerWindow: number;
  TransferWindowDuration: number;
  TransferWindowFrequencyDays: number;
}
