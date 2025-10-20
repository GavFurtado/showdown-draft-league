import React, { useState, useEffect, useMemo } from 'react';
import { Player, Draft, DraftedPokemon, League, PlayerPick } from "../api/data_interfaces";
import { formatPokemonName } from '../utils/nameFormatter';

// -- Co-located Components for Draft History Modal --
interface PlayerDraftPickItemProps {
    pick: PlayerPick;
    currentPickNumbersOnClock: number[];
    isOnClock: boolean;
}

const PlayerDraftPickItem: React.FC<PlayerDraftPickItemProps> = ({ pick, currentPickNumbersOnClock, isOnClock }) => {
    const isCurrentPick = currentPickNumbersOnClock.includes(pick.pickNumber);
    const highlightClass = (isCurrentPick && isOnClock) ? 'border-2 border-green-400' : '';
    return (
        <div className={`flex items-center justify-between p-2 rounded-md bg-gray-800 shadow-md ${highlightClass}`}>
            <div className="flex items-center flex-1 gap-x-4">
                <span className="text-sm font-semibold text-gray-400 w-12">#{pick.pickNumber}</span>
                {pick.pokemon ? (
                    <>
                        <img
                            src={pick.pokemon.PokemonSpecies.Sprites.FrontDefault}
                            alt={pick.pokemon.PokemonSpecies.Name}
                            className="w-10 h-10 object-contain bg-gray-700 rounded-full"
                        />
                        <span className="font-medium text-white">{formatPokemonName(pick.pokemon.PokemonSpecies.Name)}</span>
                    </>
                ) : (
                    <div className="flex items-center gap-x-4">
                        <div className="w-10 h-10 bg-gray-700 rounded-full" />
                        <span className="text-gray-500">Empty</span>
                    </div>
                )}
            </div>
        </div>
    );
};

interface PlayerDraftPicksListProps {
    picks: PlayerPick[];
    isOnClock: boolean;
    currentPickNumbersOnClock: number[];
}

const PlayerDraftPicksList: React.FC<PlayerDraftPicksListProps> = ({ picks, isOnClock, currentPickNumbersOnClock }) => {
    const highlightClass = isOnClock ? 'border-4 border-yellow-500' : '';
    return (
        <div className={`flex flex-col gap-y-2 p-4 bg-gray-900 max-w-xl w-full rounded-lg overflow-y-auto max-h-[60vh] ${highlightClass}`}>
            {picks.map((pick) => (
                <PlayerDraftPickItem key={pick.pickNumber} pick={pick} currentPickNumbersOnClock={currentPickNumbersOnClock} isOnClock={isOnClock} />
            ))}
        </div>
    );
};

interface PlayerPaneProps {
    player: Player | null;
    picks: PlayerPick[];
    position: 'left' | 'center' | 'right';
    authenticatedPlayerId: string | null;
    onClockPlayerId: string | null;
    currentPickNumbersOnClock: number[];
}

const PlayerPane: React.FC<PlayerPaneProps> = ({ player, picks, position, authenticatedPlayerId, onClockPlayerId, currentPickNumbersOnClock }: PlayerPaneProps) => {
    const isOnClock = player?.ID === onClockPlayerId;
    console.log(`PlayerPane: player ID: ${player?.ID}, onClockPlayerId: ${onClockPlayerId}, isOnClock: ${isOnClock}`);
    const getTransform = () => {
        switch (position) {
            case 'left':
                return 'translateX(30%) translateZ(100px) scale(0.7) rotateY(15deg)';
            case 'right':
                return 'translateX(-30%) translateZ(100px) scale(0.7) rotateY(-15deg)';
            case 'center':
            default:
                return 'translateX(0) translateZ(-100px) scale(1)';
        }
    };

    const zIndex = position === 'center' ? 20 : 10;
    const opacity = player ? 1 : 0;

    return (
        <div
            className="absolute w-full h-full transition-transform duration-500 ease-in-out flex flex-col items-center justify-center"
            style={{
                transform: getTransform(),
                zIndex,
                opacity,
                backfaceVisibility: 'hidden'
            }}
        >
            <div className="bg-white shadow-xl rounded-lg max-w-xl w-full p-6 flex flex-col items-center">
                <h3 className="text-2xl font-bold text-center text-gray-800 mb-4">
                    {player?.TeamName || player?.InLeagueName || ''}
                    {player && player.ID === authenticatedPlayerId && player.ID === onClockPlayerId && <span className="text-lg text-blue-500 ml-2">(you, on clock)</span>}
                    {player && player.ID === authenticatedPlayerId && player.ID !== onClockPlayerId && <span className="text-lg text-green-500 ml-2">(you)</span>}
                    {player && player.ID !== authenticatedPlayerId && player.ID === onClockPlayerId && <span className="text-lg text-red-500 ml-2">(on clock)</span>}
                </h3>
                <PlayerDraftPicksList picks={picks} isOnClock={isOnClock} currentPickNumbersOnClock={currentPickNumbersOnClock} />
            </div>
        </div>
    );
};


// -- Main Modal Component --

interface FullScreenDraftModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: string;
    leaguePlayers: Player[];
    draftHistory: DraftedPokemon[];
    currentDraft: Draft | null;
    currentPlayer: Player | null;
    currentLeague: League | null;
}

const FullScreenDraftModal: React.FC<FullScreenDraftModalProps> = ({ isOpen, onClose, title, leaguePlayers, draftHistory, currentDraft, currentPlayer, currentLeague }) => {
    const [currentViewedPlayerIndex, setCurrentViewedPlayerIndex] = useState(0);
    const handleNextPlayer = () => {
        if (leaguePlayers.length > 0) {
            setCurrentViewedPlayerIndex(prevIndex => (prevIndex + 1) % leaguePlayers.length);
        }
    };

    const handlePreviousPlayer = () => {
        if (leaguePlayers.length > 0) {
            setCurrentViewedPlayerIndex(prevIndex => (prevIndex - 1 + leaguePlayers.length) % leaguePlayers.length);
        }
    };

    useEffect(() => {
        if (isOpen && currentPlayer && leaguePlayers.length > 0) {
            const initialIndex = leaguePlayers.findIndex(p => p.ID === currentPlayer.ID);
            if (initialIndex !== -1) {
                setCurrentViewedPlayerIndex(initialIndex);
            }
        }
    }, [isOpen, currentPlayer, leaguePlayers]);

    useEffect(() => {
        if (!isOpen) return;

        const handleKeyDown = (event: KeyboardEvent) => {
            switch (event.key) {
                case 'ArrowLeft':
                    handlePreviousPlayer();
                    break;
                case 'ArrowRight':
                    handleNextPlayer();
                    break;
                case 'Escape':
                    onClose();
                    break;
                default:
                    break;
            }
        };

        window.addEventListener('keydown', handleKeyDown);

        return () => {
            window.removeEventListener('keydown', handleKeyDown);
        };
    }, [isOpen, handlePreviousPlayer, handleNextPlayer, onClose]);



    const getPlayerPicks = (player: Player | null): PlayerPick[] => {
        if (!currentLeague || !player) return [];
        const totalRounds = currentLeague.MaxPokemonPerPlayer;
        const numPlayers = leaguePlayers.length;
        const playerIndex = leaguePlayers.findIndex(p => p.ID === player.ID);
        if (playerIndex === -1) return [];

        const picks: PlayerPick[] = [];
        for (let round = 0; round < totalRounds; round++) {
            let pickNumber: number;
            if (currentLeague.Format.IsSnakeRoundDraft && round % 2 !== 0) {
                pickNumber = (round * numPlayers) + (numPlayers - 1 - playerIndex) + 1;
            } else {
                pickNumber = (round * numPlayers) + playerIndex + 1;
            }
            const draftedPokemon = draftHistory.find(p => p.DraftPickNumber === pickNumber && p.PlayerID === player.ID && p.DraftPickNumber > 0) || null;
            picks.push({ pickNumber, pokemon: draftedPokemon });
        }
        return picks;
    };

    const getPlayerByIndex = (index: number) => leaguePlayers[index] || null;

    const prevPlayerIndex = leaguePlayers.length > 1 ? (currentViewedPlayerIndex - 1 + leaguePlayers.length) % leaguePlayers.length : null;
    const nextPlayerIndex = leaguePlayers.length > 1 ? (currentViewedPlayerIndex + 1) % leaguePlayers.length : null;

    const currentViewedPlayer = getPlayerByIndex(currentViewedPlayerIndex);
    const prevPlayer = prevPlayerIndex !== null ? getPlayerByIndex(prevPlayerIndex) : null;
    const nextPlayer = nextPlayerIndex !== null ? getPlayerByIndex(nextPlayerIndex) : null;

    const currentPicks = getPlayerPicks(currentViewedPlayer);
    const prevPicks = getPlayerPicks(prevPlayer);
    const nextPicks = getPlayerPicks(nextPlayer);

    const currentPickNumbersOnClock = useMemo(() => {
        if (!currentDraft || !currentDraft.CurrentTurnPlayerID || !currentDraft.PlayersWithAccumulatedPicks) {
            return [];
        }
        return currentDraft.PlayersWithAccumulatedPicks[currentDraft.CurrentTurnPlayerID] || [];
    }, [currentDraft]);

    if (!isOpen) return null;

    return (
        <>
            {/* Backdrop */}
            <div
                className="fixed inset-0 bg-gray-900/90 z-40"
                onClick={onClose}
            />

            {/* Modal Content Wrapper (for centering) */}
            <div className="fixed inset-0 z-50 flex items-center justify-center p-4 pointer-events-none">
                {/* Actual Modal Content Card */}
                <div
                    className="relative bg-gray-900 shadow-xl rounded-lg w-full h-[85vh] flex flex-col p-6 pointer-events-auto"
                    onClick={(e) => e.stopPropagation()}
                >
                    {/* Title and Close Button */}
                    <div className="flex justify-between items-center w-full mb-4 flex-shrink-0">
                        <h3 className="text-2xl font-bold text-white">{title}</h3>
                        <button className="text-gray-400 hover:text-gray-200 z-50" onClick={onClose}>
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                            </svg>
                        </button>
                    </div>

                    {/* Carousel Stage */}
                    <div className="relative flex-1 w-full flex items-center justify-center overflow-hidden" style={{ transformStyle: 'preserve-3d', perspective: '1000px' }}>
                        <PlayerPane player={prevPlayer} picks={prevPicks} position="left" authenticatedPlayerId={currentPlayer?.ID || null} onClockPlayerId={currentDraft?.CurrentTurnPlayerID || null} currentPickNumbersOnClock={currentPickNumbersOnClock} />
                        <PlayerPane player={currentViewedPlayer} picks={currentPicks} position="center" authenticatedPlayerId={currentPlayer?.ID || null} onClockPlayerId={currentDraft?.CurrentTurnPlayerID || null} currentPickNumbersOnClock={currentPickNumbersOnClock} />
                        <PlayerPane player={nextPlayer} picks={nextPicks} position="right" authenticatedPlayerId={currentPlayer?.ID || null} onClockPlayerId={currentDraft?.CurrentTurnPlayerID || null} currentPickNumbersOnClock={currentPickNumbersOnClock} />
                    </div>

                    {/* Navigation Buttons */}
                    <button
                        onClick={handlePreviousPlayer}
                        className="absolute left-4 top-1/2 -translate-y-1/2 p-3 rounded-full bg-gray-700 text-white hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed z-50 shadow-lg"
                        disabled={leaguePlayers.length <= 1}
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                        </svg>
                    </button>
                    <button
                        onClick={handleNextPlayer}
                        className="absolute right-4 top-1/2 -translate-y-1/2 p-3 rounded-full bg-gray-700 text-white hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed z-50 shadow-lg"
                        disabled={leaguePlayers.length <= 1}
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                    </button>
                </div>
            </div>
        </>
    );
};

export default FullScreenDraftModal;
