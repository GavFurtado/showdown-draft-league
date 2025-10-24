export type PlayerAccumulatedPicks = {
  [playerId: string]: number[];
};

export interface DiscordUser {
  ID: string;
  Username: string;
  Discriminator?: string;
  Avatar: string; // url
}

export interface User {
  ID: string; // uuid.UUID
  DiscordID: string; // uuid.UUID
  DiscordUsername: string; // uuid.UUID
  DiscordAvatarUrl: string;
  ShowdownUsername: string;
  Role: 'user' | 'admin';
  CreatedAt: string; // ISO 8601 string
  UpdatedAt: string; // ISO 8601 string
}

// League enums
export type LeagueStatus = "PENDING" | "SETUP" | "DRAFTING" | "POST_DRAFT" | "TRANSFER_WINDOW" | "REGULARSEASON" | "PLAYOFFS" | "COMPLETED" | "CANCELLED";
export type DraftOrderType = "PENDING" | "RANDOM" | "MANUAL";
export type LeagueSeasonType = "ROUND_ROBIN_ONLY" | "PLAYOFFS_ONLY" | "HYBRID";
export type LeaguePlayoffType = "NONE" | "SINGLE_ELIM" | "DOUBLE_ELIM";
export type LeaguePlayoffSeedingType = "STANDARD" | "SEEDED" | "BYES_ONLY";

export interface LeagueFormat {
  IsSnakeRoundDraft: boolean;
  DraftOrderType: DraftOrderType;
  SeasonType: LeagueSeasonType;
  GroupCount: number;
  GamesPerOpponent: number;
  PlayoffType: LeaguePlayoffType;
  PlayoffParticipantCount: number;
  PlayoffByesCount: number;
  PlayoffSeedingType: LeaguePlayoffSeedingType;
  AllowTrading: boolean;
  AllowTransferCredits: boolean;
  TransferCreditsPerWindow: number;
  TransferCreditCap: number;
  TransferWindowFrequencyDays: number;
  TransferWindowDuration: number;
  DropCost: number;
  PickupCost: number;
  NextTransferWindowStart?: string; // ISO 8601 string 
}

export interface League {
  ID: string; // uuid.UUID
  Name: string;
  StartDate: string; // ISO 8601 string
  EndDate: string | null; // ISO 8601 string
  RulesetDescription: string;
  Status: LeagueStatus;
  MaxPokemonPerPlayer: number;
  MinPokemonPerPlayer: number;
  StartingDraftPoints: number;
  Format: LeagueFormat;
  DiscordWebhookURL: string | null;
  // NewPlayerGroupNumber: number; // has no real use in the frontend; still in json because yes
  // Relationships
  Players?: Player[];
  // LeaguePokemon?: LeaguePokemon[];
  // DraftedPokemon?: DraftedPokemon[];
}

// Draft related
export type DraftStatus = "PENDING" | "ONGOING" | "PAUSED" | "COMPLETED";
export interface Draft {
  ID: string,
  LeagueID: string,
  Status: DraftStatus,
  CurrentTurnPlayerID: string | null,
  CurrentRound: number,
  CurrentPickInRound: number,
  CurrentPickOnClock: number,
  PlayersWithAccumulatedPicks: PlayerAccumulatedPicks,
  CurrentTurnStartTime: string // ISO 8601 string
  TurnTimeLimit: number // in minutes
  StartTime: string // ISO 8601 string
  EndTime: string // ISO 8601 string
  CreatedAt?: string // ISO 8601 string
  UpdatedAt?: string // ISO 8601 string
  // Relationships
  League?: League
  CurrentTurnPlayer?: Player // pretty sure this is preloaded by backend
}

export type PlayerRole = "member" | "moderator" | "owner";
export interface Player {
  ID: string;
  UserID: string;
  LeagueID: string;
  InLeagueName: string;
  TeamName: string;
  Wins: number;
  Losses: number;
  DraftPoints: number;
  GroupNumber: number;
  SkipsLeft: number;
  TransferCredits: number;
  Role: PlayerRole; // league-specific
  IsParticipating: boolean;
  CreatedAt: string; // ISO 8601 string
  UpdatedAt: string; // ISO 8601 string
}

export interface LeaguePokemon {
  ID: string; // uuid.UUID
  LeagueId: string; // uuid.UUID
  PokemonSpeciesId: number;
  Cost: number;
  IsAvailable: boolean;
  CreatedAt?: string; // ISO 8601 string
  UpdatedAt?: string; // ISO 8601 string
  DeletedAt?: string; // ISO 8601 string
  // Relationships
  League?: League;
  PokemonSpecies: PokemonSpecies;
}

// Pokemon Species stuff
export interface PokemonSpecies {
  ID: number;
  DexID: number;
  Name: string;
  Types: string[];
  Stats: PokemonStat;
  Abilities: PokemonAbility[];
  Sprites: PokemonSprites;
  CreatedAt?: string; // ISO 8601 string
  UpdatedAt?: string; // ISO 8601 string
}

export interface PokemonStat {
  [key: string]: number; // e.g., hp, attack, defense, special-attack, special-defense, speed
}

export interface PokemonAbility {
  Name: string;
  IsHidden: boolean;
}
export interface PokemonSprites {
  FrontDefault?: string; // url
  OfficialArtwork?: string; //url
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
  viewMode?: 'draftboard' | 'teamsheet';
  cardSize?: 'default' | 'small';
  key: string;
  leaguePokemonId: string;
  pokemon: PokemonSpecies;
  cost: number;
  onImageError?: (e: React.SyntheticEvent<HTMLImageElement, Event>) => void;
  addPokemonToWishlist: (pokemonId: string) => void;
  removePokemonFromWishlist: (pokemonId: string) => void;
  isPokemonInWishlist: (pokemonId: string) => boolean;
  isFlipped: boolean;
  onFlip: (pokemonId: string) => void;
  isDraftable: boolean;
  onDraft: (leaguePokemonId: string) => void;
  isAvailable: boolean;
  isMyTurn: boolean;
}

export interface WishlistDisplayProps {
  allPokemon: LeaguePokemon[];
  wishlist: string[];
  removePokemonFromWishlist: (pokemonId: string) => void;
  clearWishlist: () => void;
  isMyTurn: boolean; // New prop
  onDraft: (leaguePokemonId: string) => void; // New prop
}

export interface DraftedPokemon {
  ID: string;
  LeagueID: string;
  PlayerID: string;
  PokemonSpeciesID: number;
  LeaguePokemonID: string;
  DraftRoundNumber: number;
  DraftPickNumber: number;
  IsReleased: boolean;
  CreatedAt: string;
  UpdatedAt: string;
  League: League;
  Player: Player;
  PokemonSpecies: PokemonSpecies;
  LeaguePokemon: LeaguePokemon;
}

export type PlayerPick = {
  pickNumber: number;
  pokemon: DraftedPokemon | null;
};
