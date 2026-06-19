import { Draft, League, Player } from '../api/data_interfaces';

// Calculates turns remaining until the player's next pick
export const calculateTurnsUntilNextPick = (
    currentDraft: Draft,
    currentLeague: League,
    currentPlayer: Player,
    allPlayers: Player[]
): number | null => {
    if (!currentDraft || !currentLeague || !currentPlayer || !allPlayers || allPlayers.length === 0) {
        return null;
    }

    const currentOverallPick = currentDraft.CurrentPickOnClock;
    const maxRounds = currentLeague.MaxPokemonPerPlayer;
    const numPlayers = allPlayers.length;
    const playersInDraftOrder = allPlayers; // Assumes sorted by initial draft order

    for (let overallPick = currentOverallPick + 1; overallPick <= maxRounds * numPlayers; overallPick++) {
        const round = Math.ceil(overallPick / numPlayers);
        const pickInRoundIndex = (overallPick - 1) % numPlayers;

        let roundOrder: (Player | undefined)[] = [];
        if (currentLeague.Format.IsSnakeRoundDraft && round % 2 === 0) {
            roundOrder = [...playersInDraftOrder].reverse();
        } else {
            roundOrder = [...playersInDraftOrder];
        }

        const playerPicking = roundOrder[pickInRoundIndex];
        if (playerPicking && playerPicking.ID === currentPlayer.ID) {
            return (overallPick - currentOverallPick) - 1;
        }
    }

    return null;
};
