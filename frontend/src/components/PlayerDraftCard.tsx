import React from 'react';
import { Player, PlayerPick } from '../api/data_interfaces';
import { PokemonListItem } from './PokemonListItem';

interface PlayerDraftCardProps {
    player: Player;
    draftPicks: PlayerPick[];
    isCurrentUserOnClock: boolean;
    remainingPoints: number;
}

const PlayerDraftCard: React.FC<PlayerDraftCardProps> = ({ player, draftPicks, isCurrentUserOnClock, remainingPoints }) => {
    const cardStyle = isCurrentUserOnClock
        ? 'border-2 border-green-500 shadow-md'
        : 'border border-gray-300';

    return (
        <div className={`bg-background-surface p-2 rounded-lg ${cardStyle} backdrop-blur-2xl`}>
            <div className="flex justify-between items-center mb-2 px-2">
                <h2 className="text-lg font-semibold truncate text-text-primary">{player.TeamName || player.InLeagueName}</h2>
                <div className="text-md font-bold text-text-secondary">{remainingPoints}pts</div>
            </div>
            <div className="border-t border-gray-600 pt-2 space-y-1">
                {draftPicks.length > 0 ? (
                    draftPicks.map(pick => (
                        pick.pokemon ? (
                            <PokemonListItem
                                key={pick.pokemon.ID}
                                pokemon={pick.pokemon.PokemonSpecies}
                                cost={pick.pokemon.LeaguePokemon.Cost}
                                leaguePokemonId={pick.pokemon.LeaguePokemonID}
                                pickNumber={pick.pickNumber}
                                // bgColor='bg-background-main'
                                showCost={true}
                            />
                        ) : (
                            <div key={pick.pickNumber} className="flex items-center justify-between p-1 text-sm text-text-secondary rounded-md bg-background-surface-alt h-[48px]">
                                <span className='pl-2'>Pick #{pick.pickNumber}</span>
                                <span>Empty</span>
                            </div>)
                    ))
                ) : (
                    <p className="text-gray-400 italic px-2">No picks made yet.</p>
                )}
            </div>
        </div>
    );
};

export default PlayerDraftCard;
