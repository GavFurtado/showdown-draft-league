import { Player } from '../api/data_interfaces';

// Sorts players by ranking (Wins desc, Losses asc)
export const sortPlayersByStanding = (players: Player[]): Player[] => {
    return [...players].sort((a, b) => {
        if (b.Wins !== a.Wins) {
            return b.Wins - a.Wins;
        }
        return a.Losses - b.Losses;
    });
};

// Gets top N players
export const getTopPlayers = (players: Player[], n: number = 5): Player[] => {
    const sorted = sortPlayersByStanding(players);
    return sorted.slice(0, n);
};
