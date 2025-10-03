import axios from 'axios';

// Define base URL for your backend API
export const API_BASE_URL = 'http://localhost:8080';

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

export const getMyProfile = () => api.get('/api/profile');

// Example: interface CreateLeagueRequest { name: string; description: string; }
// Example: interface League { id: number; name: string; description: string; }

export const createLeague = (data: LeagueCreateRequest) => api.post('/api/leagues/', data);
export const getLeague = (id: string) => api.get(`/api/leagues/${id}`);
export const getPlayersByLeague = (id: string) => api.get(`/api/leagues/${id}/players`);
export const joinLeague = (id: string) => api.post(`/api/leagues/${id}/join`);
export const leaveLeague = (id: string) => api.delete(`/api/leagues/${id}/leave`);

export const getUserMe = () => api.get('/api/users/me');
export const getMyDiscordDetails = () => api.get('/api/users/me/discord');
export const getMyLeagues = () => api.get('/api/users/me/leagues');
export const updateUserProfile = (data: any) => api.put('/api/users/profile', data); // Replace 'any' with a proper interface
export const getPlayersByUserId = (id: string) => api.get(`/api/users/${id}/players`);

export const getPlayerByID = (id: string) => api.get(`/api/players/${id}`);
export const getPlayerWithFullRoster = (id: string) => api.get(`/api/players/${id}/roster`);
export const updatePlayerProfile = (id: string, data: PlayerProfileUpdateRequest) => api.put(`/api/players/${id}/profile`, data);

// Note: Add more specific interfaces for request and response data
// based on your backend Go structs for better type safety.
