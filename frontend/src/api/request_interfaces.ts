import { LeagueFormat } from "./data_interfaces";
// User Related
export interface UpdateUserProfileRequest {
    showdownName?: string,
}


// League Related
export interface LeagueCreateRequest {
    name: string,
    rulesetDescription: string,
    maxPokemonPerPlayer: number,
    startingDraftPoints: number,
    startDate: string // ISO8601 string
    format: LeagueFormat
}

// Player Related
export interface PlayerCreateRequest {
    userId: string,
    leagueId: string,
    inLeagueName?: string,
    teamName?: string
}

export interface UpdatePlayerInfoRequest {
    InLeagueName?: string,
    TeamName?: string,
    Wins?: number,
    Losses?: number,
    DraftPoints?: string,
    DraftPosition?: string
}

// PokemonSpecies related
export interface PokemonSpecies {
    id: number,
    name: string,
    types: string[],
    FrontDefault: string // url
}

// LeaguePokemon Related
export interface LeaguePokemonCreateRequest {
    leagueID: string
    pokemonSpeciesId: number,
    cost?: number
}
export type LeaguePokemonBatchCreateRequest = LeaguePokemonCreateRequest[];

export interface LeaguePokemonUpdateRequest {
    leaguePokemonId: string,
    cost?: number,
    isAvailable?: boolean
}

// DraftedPokemon Related
export interface DraftedPokemonCreateRequest {
    leagueId: string,
    playerId: string,
    pokemonSpeciesId: string,
    draftRoundNumber: number
    draftPickNumber: number,
    isReleased?: boolean
}

export interface DraftedPokemonUpdateRequest {
    leagueId?: string,
    playerId?: string,
    pokemonSpeciesId?: string,
    draftPickNumber?: number,
    draftRoundNumber?: number,
    isReleased?: boolean
}

export interface MakePickRequest {
    RequestedPickCount: number,
    RequestedPicks: RequestedPick[]
}

export interface RequestedPick {
    LeaguePokemonID: string,
    DraftPickNumber: number
}

export interface PickupFreeAgentRequest {
    playerId: string;
}
