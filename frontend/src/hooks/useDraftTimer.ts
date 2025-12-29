import { useState, useEffect } from 'react';
import { Draft, League, Player } from '../api/data_interfaces';
import { calculateTurnsUntilNextPick } from '../utils/draftUtils';

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
        nextTurnInXTurns = calculateTurnsUntilNextPick(currentDraft, currentLeague, currentPlayer, allPlayers);
    }

    return { timeRemaining, shouldShowDraftStatus, nextTurnInXTurns };
};
