// User Related
export interface DiscordUser {
    id: string,
    username: string,
    discriminator: string,
    avatar: string // url
}
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

// LeaguePokemon Related
export interface LeaguePokemonCreateRequest {
    leagueId: string,
    pokemonSpeciesId: number,
    cost?: string
}

export interface LeaguePokemonUpdateRequest {
    leaguePokemonId: string,
    cost?: number,
    isAvailable?: boolean
}

// DraftedPokemon Related
// not fully implemented in backend yet
export interface DraftedPokemonCreateRequest {
    leagueId: string,
    playerId: string,
    pokemonSpeciesId: string,
    draftRoundNumber: number,
    draftPickNumber: number,
    isReleased?: boolean
}
// this will most likely need to be changed after backend updates
export interface DraftedPokemonUpdateRequest {
    leagueId?: string,
    playerId?: string,
    pokemonSpeciesId?: string,
    draftRoundNumber?: number,
    draftPickNumber?: number,
    isReleased?: boolean
}

