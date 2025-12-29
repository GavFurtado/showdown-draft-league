import { Game } from '../api/data_interfaces';

// Sorts games: Non-completed first, then by RoundNumber, then Completed games
export const sortGamesBySchedule = (games: Game[]): Game[] => {
    return [...games].sort((a, b) => {
        // Priority to non-completed games
        if (a.GameStatus !== 'COMPLETED' && b.GameStatus === 'COMPLETED') return -1;
        if (a.GameStatus === 'COMPLETED' && b.GameStatus !== 'COMPLETED') return 1;
        
        // Then by week (round number)
        return a.RoundNumber - b.RoundNumber;
    });
};

export const getUpcomingGames = (games: Game[], limit: number = 3): Game[] => {
    const sorted = sortGamesBySchedule(games);
    return sorted.slice(0, limit);
};
