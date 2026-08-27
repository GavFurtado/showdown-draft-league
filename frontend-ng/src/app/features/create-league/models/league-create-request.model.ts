import { LeagueVisibility } from '../../league/models/enums/league-visibility';
import { LeagueFormatRequest } from './league-format-request.model';

export interface LeagueCreateRequest {
  MaxPlayers: number;
  MaxPokemonPerPlayer: number;
  MinPokemonPerPlayer: number;
  Name: string;
  RulesetDescription: string;
  StartingDraftPoints: number;
  Visibility: LeagueVisibility;
  Format: LeagueFormatRequest;
}
