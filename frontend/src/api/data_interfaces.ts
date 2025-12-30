import React from 'react';

export interface DiscordUser {
    ID: string;
    Username: string;
    Discriminator?: string;
    Avatar: string; // url
}

export interface User {
    ID: string; // UUID
    DiscordID: string; // UUID
    DiscordUsername: string; // UUID
    DiscordAvatarUrl: string;
    ShowdownUsername: string;
    Role: 'user' | 'admin';
    CreatedAt: string; // ISO 8601 string
    UpdatedAt: string; // ISO 8601 string
}

// League Enums
export type LeagueStatus = "PENDING" | "SETUP" | "DRAFTING" | "POST_DRAFT" | "TRANSFER_WINDOW" | "REGULAR_SEASON" | "PLAYOFFS" | "COMPLETED" | "CANCELLED";
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
    AllowTransfers: boolean;
    TransfersCostCredits: boolean;
    TransferCreditsPerWindow: number;
    TransferCreditCap: number;
    TransferWindowFrequencyDays: number;
    TransferWindowDuration: number;
    DropCost: number;
    PickupCost: number;
    NextTransferWindowStart?: string; // ISO 8601 string
}

export interface League {
    ID: string; // UUID
    Name: string;

    StartDate: string; // ISO 8601 string
    EndDate: string | null; // ISO 8601 string
    NextWeeklyTick: string; // ISO 8601 string
    RegularSeasonStartDate: string; // ISO 8601 string

    CurrentWeekNumber: number;
    Status: LeagueStatus;
    MaxPokemonPerPlayer: number;
    MinPokemonPerPlayer: number;
    StartingDraftPoints: number;
    Format: LeagueFormat;
    PlayerCount: number;

    RulesetDescription: string;
    DiscordWebhookURL: string | null;

    CreatedAt?: string; // ISO 8601 string
    UpdatedAt?: string; // ISO 8601 string
    // NewPlayerGroupNumber: number; // has no real use in the frontend

    // Relationships
    Players?: Player[];
    // These next two are never populated by the backend
    // Kept here to show relationships
    DefinedPokemon?: LeaguePokemon[];
    AllDraftedPokemon?: DraftedPokemon[];
}

// Draft related
export type DraftStatus = "PENDING" | "ONGOING" | "PAUSED" | "COMPLETED";
export interface Draft {
    ID: string;
    LeagueID: string;
    Status: DraftStatus;
    CurrentTurnPlayerID: string | null;
    CurrentRound: number;
    CurrentPickInRound: number;
    CurrentPickOnClock: number;
    PlayersWithAccumulatedPicks: PlayerAccumulatedPicks;
    CurrentTurnStartTime: string; // ISO 8601 string
    TurnTimeLimit: number; // in minutes
    StartTime: string; // ISO 8601 string
    EndTime: string; // ISO 8601 string
    CreatedAt?: string; // ISO 8601 string
    UpdatedAt?: string; // ISO 8601 string
    // Relationships
    League?: League;
    CurrentTurnPlayer?: Player;
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
    ID: string; // UUID
    LeagueId: string; // UUID
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
    costSortOrder: 'asc' | 'desc';
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
    onDraft: (leaguePokemonId: string) => void;
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
    AcquiredWeek: number;
    ReleasedWeek: number;
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

export type PlayerAccumulatedPicks = {
    [playerId: string]: number[];
};

export type GameType = "GRAND_FINAL" | "REGULAR_SEASON" | "PLAYOFF_UPPER" | "PLAYOFF_LOWER" | "PLAYOFF_SINGLEELIM" | "TOURNAMENT_SINGLEELIM" | "TOURNAMENT_UPPER" | "TOURNAMENT_LOWER";
export type GameStatus = "SCHEDULED" | "APPROVAL_PENDING" | "COMPLETED" | "DISPUTED";
export interface Game {
    ID: string; // UUID
    LeagueID: string; // UUID

    Player1ID: string; // UUID
    Player2ID: string; // UUID

    WinnerID?: string; // UUID
    LoserID?: string; // UUID

    Player1Wins: number;
    Player2Wins: number;

    RoundNumber: number;
    GroupNumber?: number;

    GameType: GameType;
    Status: GameStatus;

    BracketPosition?: string;
    ShowdownReplayLinks: string[];

    ReportingPlayerID?: string; // UUID
    ApproverID?: string; // UUID

    WinnerToGameID: string; // UUID
    LoserToGameID: string; // UUID

    CreatedAt: string; // ISO 8601 string
    UpdatedAt: string; // ISO 8601 string

    // Relationships
    ReportingPlayer?: Player;
    ApproverPlayer?: Player;
    League?: League; // never populated by the backend
    Player1?: Player;
    Player2?: Player;
    Winner?: Player;
    Loser?: Player;
}
