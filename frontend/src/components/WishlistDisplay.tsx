import React from 'react';
import { LeaguePokemon, WishlistDisplayProps } from '../api/data_interfaces';

import { PokemonListItem } from './PokemonListItem';


export const WishlistDisplay: React.FC<WishlistDisplayProps> = ({ allPokemon, wishlist, removePokemonFromWishlist, clearWishlist, isMyTurn, onDraft }) => {

    // console.log("WishlistDisplay: current wishlist state", wishlist);

    const wishlistedPokemon = allPokemon.filter(lp => wishlist.includes(lp.ID));
    const getTotalCostOfWishlistedPokemon = (wishlistedPokemon: LeaguePokemon[]) => {
        return wishlistedPokemon.reduce((sum, p) => sum + (p.Cost ?? 0), 0);
    }

    if (wishlistedPokemon.length === 0) {
        return (
            <div className="p-4 bg-white rounded-lg shadow-md">
                <h2 className="text-lg font-bold mb-2 flex items-center">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6 mr-2">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M11.48 3.499a.562.562 0 0 1 1.04 0l2.125 5.111a.563.563 0 0 0 .475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 0 0-.182.557l1.285 5.385a.562.562 0 0 1-.84.61l-4.725-2.885a.562.562 0 0 0-.586 0L6.982 20.54a.562.562 0 0 1-.84-.61l1.285-5.386a.562.562 0 0 0-.182-.557L3.422 8.99a.562.562 0 0 1 .321-.989l5.518-.442a.563.563 0 0 0 .475-.345L11.48 3.5Z" />
                    </svg>
                    Wishlist ({wishlist.length}/15)
                </h2>
                <p className="text-gray-600">Your wishlist is empty.</p>
            </div>
        );
    }

    return (
        <div className="p-4 bg-white rounded-lg shadow-md">
            <div className="flex items-center justify-between mb-1">
                <h2 className="text-lg font-bold flex items-center">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6 mr-2">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M11.48 3.499a.562.562 0 0 1 1.04 0l2.125 5.111a.563.563 0 0 0 .475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 0 0-.182.557l1.285 5.385a.562.562 0 0 1-.84.61l-4.725-2.885a.562.562 0 0 0-.586 0L6.982 20.54a.562.562 0 0 1-.84-.61l1.285-5.386a.562.562 0 0 0-.182-.557L3.422 8.99a.562.562 0 0 1 .321-.989l5.518-.442a.563.563 0 0 0 .475-.345L11.48 3.5Z" />
                    </svg>
                    Wishlist ({wishlist.length}/15)
                </h2>
                <button
                    onClick={clearWishlist}
                    className="text-error-700 hover:text-white hover:bg-red-500 text-sm py-1 px-2 rounded-md border border-red-500 hover:border-red-700 transition-colors"
                >
                    Clear
                </button>
            </div>
            <div className="text-sm text-gray-600 mb-2">Total Cost: {getTotalCostOfWishlistedPokemon(wishlistedPokemon)}</div>
            <div className="space-y-2">
                {wishlistedPokemon.map(lp => (
                    <PokemonListItem
                        key={lp.ID}
                        pokemon={lp.PokemonSpecies}
                        cost={lp.Cost}
                        leaguePokemonId={lp.ID}
                        onDraft={onDraft}
                        onRemove={removePokemonFromWishlist}
                        showRemoveButton={true}
                        isMyTurn={isMyTurn}
                        isAvailable={lp.IsAvailable}
                        isWishlistItem={true}
                    />
                ))}
            </div>

        </div>
    );
};

