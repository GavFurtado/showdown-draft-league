import React, { useState, useEffect } from 'react';
import FullScreenModal from './FullScreenModal';
import { Player, Draft, DraftedPokemon, League } from "../api/data_interfaces";
import { PokemonRosterList } from "./PokemonRosterList";

interface FullScreenDraftModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: string;
    leaguePlayers: Player[];
    allPlayersDraftedPokemon: { [key: string]: DraftedPokemon[] };
    currentDraft: Draft | null;
    currentPlayer: Player | null;
    currentLeague: League | null;

}

const FullScreenDraftModal: React.FC<FullScreenDraftModalProps> = ({ isOpen, onClose, title, leaguePlayers, allPlayersDraftedPokemon, currentDraft, currentPlayer, currentLeague }) => {
    const [currentViewedPlayerIndex, setCurrentViewedPlayerIndex] = useState(0);
    const [draftOrder, setDraftOrder] = useState<string[]>([]);

    useEffect(() => {
        if (currentDraft && leaguePlayers.length > 0 && currentLeague) {
            const order: string[] = [];
            const playersInOrder = [...leaguePlayers].sort((a, b) => a.InLeagueName.localeCompare(b.InLeagueName)); // Assuming leaguePlayers is sorted by draft position or similar

            const totalRounds = currentLeague.MaxPokemonPerPlayer;

            for (let round = 0; round < totalRounds; round++) {
                if (currentLeague.Format.IsSnakeRoundDraft && round % 2 !== 0) {
                    // Odd rounds in snake draft: reverse order
                    order.push(...[...playersInOrder].reverse().map(p => p.ID));
                } else {
                    // Even rounds or linear draft: normal order
                    order.push(...playersInOrder.map(p => p.ID));
                }
            }
            setDraftOrder(order);
        }
    }, [currentDraft, leaguePlayers, currentLeague]);

    useEffect(() => {
        if (isOpen && currentDraft?.CurrentTurnPlayerID && draftOrder.length > 0) {
            const initialIndex = draftOrder.findIndex(playerId => playerId === currentDraft.CurrentTurnPlayerID);
            if (initialIndex !== -1) {
                setCurrentViewedPlayerIndex(initialIndex);
            }
        }
    }, [isOpen, currentDraft?.CurrentTurnPlayerID, draftOrder]);

    const currentViewedPlayerId = draftOrder[currentViewedPlayerIndex];
    const currentViewedPlayer = leaguePlayers.find(player => player.ID === currentViewedPlayerId);
    const currentViewedPlayerRoster = allPlayersDraftedPokemon[currentViewedPlayerId || ''] || [];

    const handleNextPlayer = () => {
        setCurrentViewedPlayerIndex(prevIndex => (prevIndex + 1) % draftOrder.length);
    };

    const handlePreviousPlayer = () => {
        setCurrentViewedPlayerIndex(prevIndex => (prevIndex - 1 + draftOrder.length) % draftOrder.length);
    };



    return (
        <div className="fixed inset-0 bg-black bg-opacity-75 backdrop-blur-sm overflow-y-auto h-full w-full text-text-primary">
            <FullScreenModal isOpen={isOpen} onClose={onClose} title={title}>
                <div className="flex flex-col items-center justify-center h-full">
                    <div className="flex items-center justify-between w-full max-w-4xl px-4">
                        <button onClick={handlePreviousPlayer} className="p-2 rounded-full bg-gray-700 text-white hover:bg-gray-600">
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                            </svg>
                        </button>
                        <div className="text-center">
                            <h3 className="text-2xl font-bold mb-2">{currentViewedPlayer?.TeamName || currentViewedPlayer?.InLeagueName}</h3>
                            <PokemonRosterList
                                roster={currentViewedPlayerRoster}
                                rosterType="drafted"
                                bgColor="bg-gray-800"
                            />
                        </div>
                        <button onClick={handleNextPlayer} className="p-2 rounded-full bg-gray-700 text-white hover:bg-gray-600">
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                            </svg>
                        </button>
                    </div>

                </div>
            </FullScreenModal>
        </div>
    );
};

export default FullScreenDraftModal;
