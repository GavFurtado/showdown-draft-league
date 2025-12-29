import React, { useMemo, useEffect } from 'react';
import FullScreenModal from './FullScreenModal';
import { Draft, DraftedPokemon, League, Player, PlayerPick } from '../api/data_interfaces';
import PlayerDraftCard from './PlayerDraftCard';

interface FullDraftViewProps {
  isOpen: boolean;
  onClose: () => void;
  draft: Draft | null;
  draftHistory: DraftedPokemon[];
  league: League | null;
  leaguePlayers: Player[];
}

const FullDraftView: React.FC<FullDraftViewProps> = ({
  isOpen,
  onClose,
  draft,
  draftHistory,
  league,
  leaguePlayers,
}) => {
  const playersWithPicks = useMemo(() => {
    if (!league || leaguePlayers.length === 0) {
      return [];
    }

    return leaguePlayers.map(player => {
      const totalRounds = league.MaxPokemonPerPlayer;
      const numPlayers = leaguePlayers.length;
      const playerIndex = leaguePlayers.findIndex(p => p.ID === player.ID);

      const picks: PlayerPick[] = [];
      if (playerIndex !== -1) {
        for (let round = 0; round < totalRounds; round++) {
          let pickNumber: number;
          if (league.Format.IsSnakeRoundDraft && round % 2 !== 0) {
            pickNumber = (round * numPlayers) + (numPlayers - 1 - playerIndex) + 1;
          } else {
            pickNumber = (round * numPlayers) + playerIndex + 1;
          }
          const draftedPokemon = draftHistory.find(p => p.DraftPickNumber === pickNumber && p.PlayerID === player.ID && p.DraftPickNumber > 0) || null;
          picks.push({ pickNumber, pokemon: draftedPokemon });
        }
      }

      return {
        ...player,
        draftPicks: picks,
      };
    });
  }, [leaguePlayers, draftHistory, league]);

  const currentPlayerIdOnClock = draft?.CurrentTurnPlayerID;

  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden';
    } else {
      document.body.style.overflow = 'unset';
    }
    return () => {
      document.body.style.overflow = 'unset';
    };
  }, [isOpen]);

  if (!isOpen) {
    return null;
  }

  return (
    <FullScreenModal isOpen={isOpen} onClose={onClose}>
      <div className="p-4 rounded-2xl bg-background-primary text-text-primary h-full overflow-y-auto">
        <h1 className="text-white text-3xl font-bold text-center mb-6">Draft</h1>
        <div className="grid gap-4" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(350px, 1fr))' }}>
          {playersWithPicks.map(player => (
            <PlayerDraftCard
              key={player.ID}
              player={player}
              draftPicks={player.draftPicks}
              remainingPoints={player.DraftPoints}
              isCurrentUserOnClock={draft?.Status !== 'COMPLETED' && player.ID === currentPlayerIdOnClock}
            />
          ))}
        </div>
      </div>
    </FullScreenModal>
  );
};

export default FullDraftView;
