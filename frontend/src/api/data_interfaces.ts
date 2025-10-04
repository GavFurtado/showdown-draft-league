import { NoSubstitutionTemplateLiteral } from "typescript";

export interface User {
  id: string; // uuid.UUID
  discordId: string; // uuid.UUID
  discordUsername: string; // uuid.UUID
  discordAvatarUrl: string;
  showdownUsername: string;
  role: 'user' | 'admin';
  createdAt: string;
  updatedAt: string;
}

// Assuming these enums are also defined in your Go backend and sent as strings
export type LeagueStatus = "pending" | "active" | "completed" | "cancelled";
export type LeagueSeasonType = "ROUND_ROBIN_ONLY" | "PLAYOFFS_ONLY" | "HYBRID";
export type LeaguePlayoffType = "NONE" | "SINGLE_ELIM" | "DOUBLE_ELIM";
export type LeaguePlayoffSeedingType = "STANDARD" | "SEEDED" | "BYES_ONLY";

export interface LeagueFormat {
  seasonType: LeagueSeasonType;
  groupCount: number;
  gamesPerOpponent: number;
  playoffType: LeaguePlayoffType;
  playoffParticipantCount: number;
  playoffByesCount: number;
  playoffSeedingType: LeaguePlayoffSeedingType;
  isSnakeRoundDraft: boolean;
  allowTrading: boolean;
  allowTransferCredits: boolean;
  transferCreditsPerWindow: number;
}

export interface League {
  id: string; // uuid.UUID
  name: string;
  startDate: string; // time.Time
  endDate: string | null; // *time.Time
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
  createdAt: string;
  updatedAt: string;
}

export interface LeaguePokemon {
  id: string; // uuid.UUID
  leagueId: string; // uuid.UUID
  pokemonSpeciesId: number;
  cost: number;
  isAvailable: boolean;
  createdAt?: string; // time.Time
  updatedAt?: string; // time.Time
  deletedAt?: string; // time.time;
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
  created_at?: string;
  updated_at?: string;
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
}

export interface WishlistDisplayProps {
  allPokemon: LeaguePokemon[];
  wishlist: string[];
  removePokemonFromWishlist: (pokemonId: string) => void;
  clearWishlist: () => void;
}
