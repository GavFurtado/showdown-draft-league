import { LeagueFormat } from "./data_interfaces";

export interface JoinLeagueRequest {
    UserID: string,
    LeagueID: string,
    InLeagueName?: string,
    TeamName?: string,
}

// User Related
export interface UpdateUserProfileRequest {
    ShowdownName?: string,
}

// League Related
export interface LeagueCreateRequest {
    Name: string,
    RulesetDescription: string,
    MaxPokemonPerPlayer: number,
    StartingDraftPoints: number,
    StartDate: string
    Format: LeagueFormat
}

// Player Related
export interface PlayerCreateRequest {
    UserId: string,
    LeagueId: string,
    InLeagueName?: string,
    TeamName?: string
}

export interface UpdatePlayerInfoRequest {
    InLeagueName?: string,
    TeamName?: string,
    Wins?: number,
    Losses?: number,
    DraftPoints?: string,
    DraftPosition?: string
}

// LeaguePokemon Related
export interface LeaguePokemonCreateRequest {
    LeagueID: string
    PokemonSpeciesId: number,
    Cost?: number
}
export type LeaguePokemonBatchCreateRequest = LeaguePokemonCreateRequest[];

export interface LeaguePokemonUpdateRequest {
    LeaguePokemonId: string,
    Cost?: number,
    IsAvailable?: boolean
}

// DraftedPokemon Related
export interface DraftedPokemonCreateRequest {
    LeagueId: string,
    PlayerId: string,
    PokemonSpeciesId: string,
    DraftRoundNumber: number
    DraftPickNumber: number,
    IsReleased?: boolean
}

export interface DraftedPokemonUpdateRequest {
    LeagueId?: string,
    PlayerId?: string,
    PokemonSpeciesId?: string,
    DraftPickNumber?: number,
    DraftRoundNumber?: number,
    IsReleased?: boolean
}

export interface MakePickRequest {
    RequestedPickCount: number,
    RequestedPicks: RequestedPick[]
}

export interface RequestedPick {
    LeaguePokemonID: string,
    DraftPickNumber: number
}
