import { Game } from "./data_interfaces";

export interface ApiSuccessResponse {
    message: string;
}

// This is how the backend does error responses for the timebeing
// Eventually this will be deprecated for better
// API Error Responses that are more verbose
export interface ApiErrorResponse {
    error: string;
}

export interface GetGamesResponse {
    games: Game[];
}

export interface GetGameByIdResponse {
    games: Game;
}