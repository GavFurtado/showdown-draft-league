import type { League } from '../../features/league/models/league.model';
import { LeagueStatus } from '../../features/league/models/enums/league-status';
import { LeagueSeasonType } from '../../features/league/models/enums/league-season-type';
import type { User } from '../models/user.model';

export const TEST_USER: User = {
  ID: 'user-1',
  DiscordID: '1234',
  DiscordUsername: 'Gavin',
  DiscordAvatarURL: '',
  ShowdownUsername: 'GavinTest',
  Role: 'user',
};

// Full wire-shape League fixture (PascalCase) for feature specs.
export function makeLeague(overrides: Partial<League> = {}): League {
  return {
    ID: 'league-1',
    Name: 'Test League',
    RulesetDescription: 'Be nice',
    PlayerCount: 4,
    MaxPlayers: 8,
    Status: LeagueStatus.SETUP,
    Visibility: 'PRIVATE',
    MaxPokemonPerPlayer: 6,
    MinPokemonPerPlayer: 4,
    CurrentWeekNumber: 0,
    StartDate: '',
    EndDate: '',
    RegularSeasonStartDate: '',
    OwnerUserID: 'owner-1',
    DiscordWebhookURL: '',
    StartingDraftPoints: 20,
    NextWeeklyTick: '',
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
      NextTransferWindowStart: '',
    },
    OwnerUser: {
      ID: 'owner-1',
      DiscordID: '1',
      DiscordUsername: 'Owner',
      DiscordAvatarURL: '',
      ShowdownUsername: 'owner',
      Role: 'user',
    },
    Members: [],
    ...overrides,
  } as League;
}
