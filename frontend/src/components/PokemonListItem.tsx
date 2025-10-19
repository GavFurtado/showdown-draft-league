import React from 'react';
import { PokemonSpecies } from '../api/data_interfaces'; // Reverted to Pokemon
import pokeballIcon from '../assets/pokeball-icon.png'
import { formatPokemonName } from "../utils/nameFormatter";

interface PokemonListItemProps {
    pokemon: PokemonSpecies;
    cost: number | null;
    leaguePokemonId: string;
    onDraft?: (leaguePokemonId: string) => void;
    onRemove?: (leaguePokemonId: string) => void;
    isMyTurn?: boolean;
    isAvailable?: boolean;
    bgColor?: string;
    pickNumber?: number;
    isWishlistItem?: boolean;
    showCost?: boolean;
    showRemoveButton?: boolean;
}

export const PokemonListItem: React.FC<PokemonListItemProps> = ({
    pokemon,
    cost,
    leaguePokemonId,
    onDraft,
    onRemove,
    isMyTurn,
    isAvailable,
    pickNumber,
    bgColor = 'bg-background-tertiary',
    isWishlistItem,
    showCost = true,
    showRemoveButton = false
}) => {
    // console.log("PokemonListItem:: pokemon prop: ", pokemon)
    return (
        <div className={`flex items-center justify-between p-2 rounded-md ${bgColor}`}>
            <div className="flex items-center flex-1 gap-x-2">
                <img
                    src={pokemon.Sprites.FrontDefault}
                    alt={pokemon.Name}
                    className="w-10 h-10 object-contain"
                />
                <div className="flex-1">
                    <p className={`font-medium text-gray-800 ${isWishlistItem && !isAvailable ? 'line-through text-gray-500' : ''}`}>{formatPokemonName(pokemon.Name)}</p>
                    {showCost && <p className="text-sm text-gray-600">Cost: {cost ?? 'N/A'}</p>}
                </div>
            </div>
            <div className="flex items-center gap-4">
                {pickNumber && <p className="text-sm font-semibold text-gray-400">#{pickNumber}</p>}
                {isMyTurn && isAvailable && onDraft && (
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            onDraft(leaguePokemonId);
                        }}
                        className="relative flex items-center align-center justify-center rounded-full p-0 transition-all duration-150 hover:bg-gray-200 hover:shadow-g"
                    >
                        <img src={pokeballIcon} alt="draft mon" className="w-5 h-5 rounded-full transition-transform duration-150 transform hover:scale-125" />
                    </button>
                )}
                {showRemoveButton && onRemove && (
                    <button
                        onClick={() => onRemove(leaguePokemonId)}
                        className="text-red-500 hover:text-red-700 text-sm p-1 rounded-full hover:bg-red-100"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
                            <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                )}
            </div>
        </div>
    );
};