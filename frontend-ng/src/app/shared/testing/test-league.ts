import type { League } from '../../features/league/models/league.model';
import { LeagueStatus } from '../../features/league/models/enums/league-status';
import { LeagueSeasonType } from '../../features/league/models/enums/league-season-type';
import { asIsoDateTime, asUuid } from '../types/branded-strings';
import type { User } from '../models/user.model';
import { UserRole } from '../models/enums/user-role';

const FIXED_TS = '2026-08-30T12:00:00.000Z';

export const TEST_USER: User = {
  ID: asUuid('33333333-3333-4333-8333-333333333333'),
  DiscordID: '1234',
  DiscordUsername: 'Tester',
  DiscordAvatarURL: '',
  ShowdownUsername: 'tester_show',
  Role: UserRole.USER,
};

// Full wire-shape League fixture (PascalCase) for feature specs.
export function makeLeague(overrides: Partial<League> = {}): League {
  return {
    ID: asUuid('11111111-1111-4111-8111-111111111111'),
    Name: 'Test League',
    RulesetDescription: 'Be nice or something lmao',
    PlayerCount: 4,
    MaxPlayers: 8,
    Status: LeagueStatus.SETUP,
    Visibility: 'PRIVATE',
    MaxPokemonPerPlayer: 6,
    MinPokemonPerPlayer: 4,
    CurrentWeekNumber: 0,
    StartDate: asIsoDateTime(FIXED_TS),
    EndDate: asIsoDateTime(FIXED_TS),
    RegularSeasonStartDate: asIsoDateTime(FIXED_TS),
    OwnerUserID: asUuid('55555555-5555-4555-8555-555555555555'),
    DiscordWebhookURL: '',
    StartingDraftPoints: 20,
    NextWeeklyTick: asIsoDateTime(FIXED_TS),
    NewPlayerGroupNumber: 1,
    Format: {
      IsSnakeRoundDraft: true,
      DraftOrderType: 'MANUAL',
      SeasonType: LeagueSeasonType.ROUND_ROBIN_ONLY,
      GroupCount: 1,
      PlayoffSeedingType: 'STANDARD',
      PlayoffType: 'NONE',
      PlayoffParticipantCount: 0,
      PlayoffByesCount: 0,
      AllowTransfers: false,
      TransferUsesCredits: false,
      DropCost: 0,
      PickupCost: 0,
      TransferCreditCap: 0,
      TransferCreditsPerWindow: 0,
      TransferWindowDuration: 3,
      TransferWindowFrequencyDays: 7,
      NextTransferWindowStart: asIsoDateTime(FIXED_TS),
    },
    OwnerUser: {
      ID: asUuid('55555555-5555-4555-8555-555555555555'),
      DiscordID: '1',
      DiscordUsername: 'Owner',
      DiscordAvatarURL: '',
      ShowdownUsername: 'owner',
      Role: UserRole.USER,
    },
    Members: [],
    ...overrides,
  } as League;
}
