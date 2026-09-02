import { ISODateTimeString } from '../../../shared/types/branded-strings';
import { LeagueDraftOrderType } from './enums/league-draft-order-type';
import { LeaguePlayoffSeedingType } from './enums/league-playoff-seeding-type';
import { LeaguePlayoffType } from './enums/league-playoff-type';
import { LeagueSeasonType } from './enums/league-season-type';

export interface LeagueFormat {
  SeasonType: LeagueSeasonType;
  DraftOrderType: LeagueDraftOrderType;
  IsSnakeRoundDraft: boolean;
  GroupCount: number;
  PlayoffByesCount: number;
  PlayoffParticipantCount: number;
  PlayoffSeedingType: LeaguePlayoffSeedingType;
  PlayoffType: LeaguePlayoffType;
  AllowTransfers: boolean;
  TransferUsesCredits: boolean;
  DropCost: number;
  PickupCost: number;
  TransferCreditCap: number;
  TransferCreditsPerWindow: number;
  TransferWindowDuration: number;
  TransferWindowFrequencyDays: number;
  NextTransferWindowStart?: ISODateTimeString;
}
