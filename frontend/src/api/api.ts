import axios from 'axios';
import { LeagueCreateRequest, UpdatePlayerInfoRequest, UpdateUserProfileRequest, LeaguePokemonCreateRequest, LeaguePokemonBatchCreateRequest, LeaguePokemonUpdateRequest, MakePickRequest, PickupFreeAgentRequest } from './request_interfaces'
import { DiscordUser, Draft, League, LeaguePokemon } from './data_interfaces';

export const API_BASE_URL = 'http://localhost:8080'; // temp; make this an env var

const api = axios.create({
    baseURL: API_BASE_URL,
    withCredentials: true
});

// --- Public Routes -- -
export const getDiscordLoginUrl = () => `${API_BASE_URL}/auth/discord/login`;
export const discordCallback = (code: string) => api.get(`/auth/discord/callback?code=${code}`);
export const logout = () => api.post('/auth/logout')


// --- Protected Routes ---
// User calls
export const getUserMe = () => api.get('/api/users/me');
export const getMyDiscordDetails = () => api.get<DiscordUser>('/api/users/me/discord');
export const getMyLeagues = () => api.get('/api/users/me/leagues');
export const updateUserProfile = (data: UpdateUserProfileRequest) => api.put('/api/users/profile', data);
export const getPlayersByUserId = (id: string) => api.get(`/api/users/${id}/players`);

// League calls
export const createLeague = (data: LeagueCreateRequest) => api.post('/api/leagues/', data);
export const getLeague = (leagueId: string) => api.get<League>(`/api/leagues/${leagueId}`);
export const getPlayersByLeague = (leagueId: string) => api.get(`/api/leagues/${leagueId}/players`);
export const joinLeague = (leagueId: string) => api.post(`/api/leagues/${leagueId}/join`);
export const leaveLeague = (leagueId: string) => api.delete(`/api/leagues/${leagueId}/leave`);

// Player calls
export const getPlayerById = (leagueId: string, id: string) => api.get(`/api/leagues/${leagueId}/player/${id}`);
export const getPlayerByUserIdAndLeagueId = (leagueId: string, userId: string) =>
    api.get(`/api/leagues/${leagueId}/player?userId=${userId}`);
export const getPlayerWithFullRoster =
    (leagueId: string, id: string) =>
        api.get(`/api/leagues/${leagueId}/player/${id}/roster`);
export const updatePlayerProfile =
    (leagueId: string, id: string, data: UpdatePlayerInfoRequest) =>
        api.put(`/api/leagueId/${leagueId}/player/${id}/profile`, data);

// LeaguePokemon calls
export const getAllLeaguePokmeon = (leagueId: string) =>
    api.get<LeaguePokemon[]>(`/api/leagues/${leagueId}/pokemon`);
export const createLeaguePokemonSingle = (leagueId: string, data: LeaguePokemonCreateRequest) =>
    api.post(`/api/leagues/${leagueId}/pokemon/single`, data);
export const createLeaguePokemonBatch = (leagueId: string, data: LeaguePokemonBatchCreateRequest) =>
    api.post(`/api/leagues/${leagueId}/pokemon/batch`, data);
export const updateLeaguePokemon = (leagueId: string, data: LeaguePokemonUpdateRequest) =>
    api.put(`/api/leagues/${leagueId}/pokemon`, data);
export const pickupFreeAgent = (leagueId: string, leaguePokemonId: string, data: PickupFreeAgentRequest) =>
    api.post(`/api/leagues/${leagueId}/pokemon/${leaguePokemonId}/pickup`, data);

// DraftedPokemon calls
export const getDraftedPokemonByID = (leagueId: string, id: string) =>
    api.get(`/api/leagues/${leagueId}/drafted_pokemon/${id}`);
export const getDraftedPokemonByPlayer = (leagueId: string, playerId: string) =>
    api.get(`/api/leagues/${leagueId}/drafted_pokemon/player/${playerId}`);
export const getDraftedPokemonByLeague = (leagueId: string) =>
    api.get(`/api/leagues/${leagueId}/drafted_pokemon/`);
export const getActiveDraftedPokemonByLeague = (leagueId: string) =>
    api.get(`/api/leagues/${leagueId}/drafted_pokemon/active`);
export const getReleasedPokemonByLeague = (leagueId: string) =>
    api.get(`/api/leagues/${leagueId}/drafted_pokemon/released`);
export const isPokemonDrafted = (leagueId: string, speciesId: string) =>
    api.get(`/api/leagues/${leagueId}/drafted_pokemon/is_drafted/${speciesId}`);
export const getNextDraftPickNumber = (leagueId: string) =>
    api.get(`/api/leagues/${leagueId}/drafted_pokemon/next_pick_number`);
export const releasePokemon = (leagueId: string, id: string) =>
    api.patch(`/api/leagues/${leagueId}/drafted_pokemon/${id}/release`);
export const getDraftedPokemonCountByPlayer = (leagueId: string, playerId: string) =>
    api.get(`/api/leagues/${leagueId}/drafted_pokemon/count/${playerId}`);
export const getDraftHistory = (leagueId: string) =>
    api.get(`/api/leagues/${leagueId}/drafted_pokemon/history`);
export const dropPokemon = (leagueId: string, id: string) =>
    api.post(`/api/leagues/${leagueId}/drafted_pokemon/${id}/drop`);

// Draft Management Endpoints
export const getDraftByID = (leagueId: string, draftId: string) =>
    api.get<Draft>(`/api/leagues/${leagueId}/draft/${draftId}`);
export const getDraftByLeagueID = (leagueId: string) =>
    api.get<Draft>(`/api/leagues/${leagueId}/draft/`);
export const startDraft = (leagueId: string) =>
    api.post<Draft>(`/api/leagues/${leagueId}/draft/start`);
export const makePick = (leagueId: string, data: MakePickRequest) =>
    api.post(`/api/leagues/${leagueId}/draft/pick`, data);
export const skipPick = (leagueId: string) =>
    api.post(`/api/leagues/${leagueId}/draft/skip`);
export const startTransferPeriod = (leagueId: string) =>
    api.post(`/api/leagues/${leagueId}/transfers/start`);
export const endTransferPeriod = (leagueId: string) =>
    api.post(`/api/leagues/${leagueId}/transfers/end`);


