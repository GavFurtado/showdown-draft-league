import axios from 'axios';
import { DiscordUser, LeagueCreateRequest, UpdatePlayerInfoRequest, UpdateUserProfileRequest } from './request_interfaces'
import { Pokemon, League } from './data_interfaces'; // <--- ADD THIS IMPORT

export const API_BASE_URL = 'http://localhost:8080'; // temp. make this an env var

const api = axios.create({
    baseURL: API_BASE_URL,
});

// --- Public Routes ---
export const discordLogin = () => api.get('/auth/discord/login');
export const discordCallback = (code: string) => api.get(`/auth/discord/callback?code=${code}`);


// --- Protected Routes ---

// Add a request interceptor to include the JWT token
api.interceptors.request.use(
    (config) => {
        const token = localStorage.getItem('jwt_token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    (error) => {
        return Promise.reject(error);
    }
);

export const getUserMe = () => api.get('/api/users/me');
export const getMyDiscordDetails = () => api.get<DiscordUser>('/api/users/me/discord');
export const getMyLeagues = () => api.get('/api/users/me/leagues');
export const updateUserProfile = (data: UpdateUserProfileRequest) => api.put('/api/users/profile', data);
export const getPlayersByUserId = (id: string) => api.get(`/api/users/${id}/players`);

export const createLeague = (data: LeagueCreateRequest) => api.post('/api/leagues/', data);
export const getLeague = (leagueId: string) => api.get<League>(`/api/leagues/${leagueId}`);
export const getPlayersByLeague = (leagueId: string) => api.get(`/api/leagues/${leagueId}/players`);
export const joinLeague = (leagueId: string) => api.post(`/api/leagues/${leagueId}/join`);
export const leaveLeague = (leagueId: string) => api.delete(`/api/leagues/${leagueId}/leave`);

export const getPlayerByID = (leagueId: string, id: string) => api.get(`/api/leagues/${leagueId}/players/${id}`);
export const getPlayerWithFullRoster =
    (leagueId: string, id: string) =>
        api.get(`/api/leagues/${leagueId}/players/${id}/roster`);
export const updatePlayerProfile =
    (leagueId: string, id: string, data: UpdatePlayerInfoRequest) =>
        api.put(`/api/leagueId/${leagueId}/players/${id}/profile`, data);

// --- New API Call for Pokemon Data ---
export const getAvailablePokemon = () => api.get<Pokemon[]>('/api/pokemon/available');
