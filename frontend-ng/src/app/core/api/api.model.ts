import { Game } from '../../features/league/models/game.model';

export interface ApiErrorResponse {
  Timestamp: string;
  Status: number;
  Error: string;
  Message: string;
  Path: string;
}

// ClientError idea came from pdz-client
// https://github.com/LumarisX/pokemon-draftzone-client/blob/main/src/pdz/layout/error/error.service.ts
//
// If this ever becomes DB-backed, add an id (but then again maybe the server should create the id)
export interface ClientError {
  message: string;
  status?: number;
  statusText?: string;
  url?: string;
  code?: string | null;
  details?: unknown;
  stack?: string;
  meta?: {
    requestId: string | null;
    timestamp: string;
  };
}

// Shared object wrapper envelopes
// NOTE: Highly likely to change on the server to something more systematic
// It also isn't PascalCase keys...
export interface MessageResponse {
  message: string;
}

export interface StatusResponse {
  status: string;
}

export interface TokenResponse {
  Token: string;
}

// yeah idk what i was thinking
export interface NextPickNumberResponse {
  // eslint-disable-next-line @typescript-eslint/naming-convention -- wire key is snake_case
  next_pick_number: number;
}

export interface GamesResponse {
  games: Game[];
}

// Need to normalize to games and only have one game maybe
// note sure
export interface GameResponse {
  game: Game;
}
