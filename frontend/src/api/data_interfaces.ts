
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
  draft_points: number;
  draftPosition: number;
  role: PlayerRole; // league-specific
  isParticipating: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface PokemonStat {
  [key: string]: number; // e.g., hp, attack, defense, special-attack, special-defense, speed
}

export interface PokemonAbility {
  ability: {
    name: string;
    url: string;
  };
  is_hidden: boolean;
  slot: number;
}

export interface PokemonSprites {
  front_default: string;
}

export interface Pokemon {
  id: number;
  name: string;
  types: string[];
  stats: PokemonStat;
  abilities: PokemonAbility[];
  cost: number;
}

export interface FilterState {
  selectedTypes: string[];
  selectedCost: string;
  sortByStat: string;
  sortOrder: 'asc' | 'desc';
}

export interface DraftCardProps {
  name: string;
  pic: string;
  type: string[];
  hp: number;
  ability: PokemonAbility[];
  attack: number;
  defense: number;
  specialAtk: number;
  specialDef: number;
  speed: number;
  cost: number;
}
