export type PlayerAccumulatedPicks = {
  [playerId: string]: number[];
};

export interface DiscordUser {
  id: string;
  discordId: string;
  discordUsername: string;
  discordAvatarUrl: string;
}

export interface User {
  id: string; // uuid.UUID
  discordId: string; // uuid.UUID
  discordUsername: string; // uuid.UUID
  discordAvatarUrl: string;
  showdownUsername: string;
  role: 'user' | 'admin';
  createdAt: string; // ISO 8601 string
  updatedAt: string; // ISO 8601 string
}

// League enums
export type LeagueStatus = "pending" | "active" | "completed" | "cancelled";
export type DraftOrderType = "PENDING" | "RANDOM" | "MANUAL";
export type LeagueSeasonType = "ROUND_ROBIN_ONLY" | "PLAYOFFS_ONLY" | "HYBRID";
export type LeaguePlayoffType = "NONE" | "SINGLE_ELIM" | "DOUBLE_ELIM";
export type LeaguePlayoffSeedingType = "STANDARD" | "SEEDED" | "BYES_ONLY";

export interface LeagueFormat {
  is_snake_round_draft: boolean;
  draft_order_type: DraftOrderType;
  season_type: LeagueSeasonType;
  group_count: number;
  games_per_opponent: number;
  playoff_type: LeaguePlayoffType;
  playoff_participant_count: number;
  playoff_byes_count: number;
  playoff_seeding_type: LeaguePlayoffSeedingType;
  allow_trading: boolean;
  allow_transfer_credits: boolean;
  transfer_credits_per_window: number;
  transfer_credit_cap: number;
  transfer_window_frequency_days: number;
  transfer_window_duration: number;
  drop_cost: number;
  pickup_cost: number;
  next_transfer_window_start?: string; // ISO 8601 string 
}

export interface League {
  id: string; // uuid.UUID
  name: string;
  startDate: string; // ISO 8601 string
  endDate: string | null; // ISO 8601 string
  rulesetDescription: string;
  status: LeagueStatus;
  maxPokemonPerPlayer: number;
  startingDraftPoints: number;
  format: LeagueFormat;
  discordWebhookURL: string | null;

  // Relationships
  // Players?: Player[];
  // LeaguePokemon?: LeaguePokemon[];
  // DraftedPokemon?: DraftedPokemon[];
}

// Draft related
export type DraftStatus = "PENDING" | "ONGOING" | "PAUSED" | "COMPLETED";
export interface Draft {
  id: string,
  leagueId: string,
  status: DraftStatus,
  currentTurnPlayerID: string | null,
  currentRound: number,
  currentPickInRound: number,
  currentPickOnClock: number,
  playersWithAccumulatedPicks: PlayerAccumulatedPicks,
  currentTurnStartTime: string // ISO 8601 string
  turnTimeLimit: number // in minutes
  startTime: string // ISO 8601 string
  endTime: string // ISO 8601 string
  createdAt?: string // ISO 8601 string
  updatedAt?: string // ISO 8601 string
  // Relationships
  league?: League
  CurrentTurnPlayer?: Player // pretty sure this is preloaded by backend
}

export type PlayerRole = "member" | "moderator" | "owner";
export interface Player {
  id: string;
  userId: string;
  leagueId: string;
  inLeagueName: string;
  teamName: string;
  wins: number;
  losses: number;
  role: PlayerRole; // league-specific
  isParticipating: boolean;
  createdAt: string; // ISO 8601 string
  updatedAt: string; // ISO 8601 string
}

export interface LeaguePokemon {
  id: string; // uuid.UUID
  leagueId: string; // uuid.UUID
  pokemonSpeciesId: number;
  cost: number;
  isAvailable: boolean;
  createdAt?: string; // ISO 8601 string
  updatedAt?: string; // ISO 8601 string
  deletedAt?: string; // ISO 8601 string
  // Relationships
  League?: League;
  PokemonSpecies: Pokemon;
}

// Pokemon Species stuff
export interface Pokemon {
  id: number;
  dex_id: number;
  name: string;
  types: string[];
  stats: PokemonStat;
  abilities: PokemonAbility[];
  sprites: PokemonSprites;
  created_at?: string; // ISO 8601 string
  updated_at?: string; // ISO 8601 string
}

export interface PokemonStat {
  [key: string]: number; // e.g., hp, attack, defense, special-attack, special-defense, speed
}

export interface PokemonAbility {
  name: string;
  is_hidden: boolean;
}
export interface PokemonSprites {
  front_default?: string; // url
  official_artwork?: string; //url
}



export interface FilterState {
  selectedTypes: string[];
  minCost: string;
  maxCost: string;
  costSortOrder: 'asc' | 'desc'
  sortByStat: string;
  sortOrder: 'asc' | 'desc';
}

export interface DraftCardProps {
  key: string;
  leaguePokemonId: string;
  pokemon: Pokemon;
  cost: number;
  onImageError?: (e: React.SyntheticEvent<HTMLImageElement, Event>) => void;
  addPokemonToWishlist: (pokemonId: string) => void;
  removePokemonFromWishlist: (pokemonId: string) => void;
  isPokemonInWishlist: (pokemonId: string) => boolean;
  isFlipped: boolean;
  onFlip: (pokemonId: string) => void;
  isDraftable: boolean;
  onDraft: (leaguePokemonId: string) => void;
}

export interface WishlistDisplayProps {
  allPokemon: LeaguePokemon[];
  wishlist: string[];
  removePokemonFromWishlist: (pokemonId: string) => void;
  clearWishlist: () => void;
  isMyTurn: boolean; // New prop
  onDraft: (leaguePokemonId: string) => void; // New prop
}
