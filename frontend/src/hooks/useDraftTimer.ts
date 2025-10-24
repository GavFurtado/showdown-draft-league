import { useState, useEffect } from 'react';
import { Draft, League, Player } from '../api/data_interfaces';

export const useDraftTimer = (currentDraft: Draft | null | undefined, currentLeague: League | null | undefined, currentPlayer: Player | null | undefined, allPlayers: Player[] | undefined) => {
    const [timeRemaining, setTimeRemaining] = useState<string>('');

    useEffect(() => {
        if (currentDraft?.Status === 'ONGOING' && currentDraft.CurrentTurnStartTime && currentDraft.TurnTimeLimit) {
            const interval = setInterval(() => {
                const startTime = new Date(currentDraft.CurrentTurnStartTime);
                const timeLimitMinutes = currentDraft.TurnTimeLimit;
                const endTime = new Date(startTime.getTime() + timeLimitMinutes * 60000);
                const now = new Date();
                const diff = endTime.getTime() - now.getTime();

                if (diff <= 0) {
                    setTimeRemaining('00:00:00');
                    clearInterval(interval);
                    window.location.reload(); // Trigger page refresh
                    return;
                }

                const hours = Math.floor(diff / (1000 * 60 * 60));
                const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
                const seconds = Math.floor((diff % (1000 * 60)) / 1000);

                setTimeRemaining(
                    `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
                );
            }, 1000);

            return () => clearInterval(interval);
        }
    }, [currentDraft]);

    const shouldShowDraftStatus = currentLeague?.Status === 'DRAFTING' || currentLeague?.Status === 'POST_DRAFT';

    let nextTurnInXTurns: number | null = null;

    if (currentDraft && currentLeague && currentPlayer && allPlayers && allPlayers.length > 0) {
        const currentOverallPick = currentDraft.CurrentPickOnClock;
        const maxRounds = currentLeague.MaxPokemonPerPlayer; // Assuming MaxPokemonPerPlayer is total rounds
        const numPlayers = allPlayers.length;

        // IMPORTANT: This assumes allPlayers is already sorted by initial draft order.
        // If not, this calculation will be incorrect. The backend League model has Players []Player,
        // but no explicit order. For now, we assume the order in allPlayers is the initial draft order.
        const playersInDraftOrder = allPlayers;

        let foundNextPick = false;
        // Start searching from the current overall pick + 1
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
            if (!playerPicking) continue;

            if (playerPicking && playerPicking.ID === currentPlayer.ID) {
                nextTurnInXTurns = (overallPick - currentOverallPick) - 1;
                foundNextPick = true;
                break;
            }
        }
    }

    return { timeRemaining, shouldShowDraftStatus, nextTurnInXTurns };
};
