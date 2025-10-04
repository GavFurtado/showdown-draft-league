import { useState, useEffect } from 'react';

const WISHLIST_STORAGE_KEY_PREFIX = 'wishlist_';
const MAX_WISHLIST_SIZE = 15;

interface UseWishlistHook {
    wishlist: string[];
    addPokemonToWishlist: (pokemonId: string) => string[];
    removePokemonFromWishlist: (pokemonId: string) => string[];
    isPokemonInWishlist: (pokemonId: string) => boolean;
    clearWishlist: () => void;
}

export const useWishlist = (leagueId: string): UseWishlistHook => {
    console.log("useWishlist: Initializing for leagueId", leagueId);
    const storageKey = `${WISHLIST_STORAGE_KEY_PREFIX}${leagueId}`;
    const [wishlist, setWishlist] = useState<string[]>(() => {
        try {
            const storedWishlist = localStorage.getItem(storageKey);
            const initialWishlist = storedWishlist ? JSON.parse(storedWishlist) : [];
            console.log("useWishlist: Initial wishlist from storage", storageKey, initialWishlist);
            console.log("useWishlist: Current leagueId for storageKey", leagueId, storageKey);
            return initialWishlist;
        } catch (error) {
            console.error("Error parsing wishlist from localStorage", error);
            return [];
        }
    });

    useEffect(() => {
        console.log("useWishlist: Saving wishlist to storage", storageKey, wishlist);
        try {
            localStorage.setItem(storageKey, JSON.stringify(wishlist));
        } catch (error) {
            console.error("Error saving wishlist to localStorage", error);
        }
    }, [wishlist, storageKey]);

    const addPokemonToWishlist = (pokemonId: string) => {
        setWishlist(prevWishlist => {
            console.log("addPokemonToWishlist: Adding", pokemonId, "to", prevWishlist);
            if (prevWishlist.length >= MAX_WISHLIST_SIZE) {
                console.warn("Wishlist is full. Cannot add more Pokemon.");
                return prevWishlist; // Do not add if wishlist is full
            }
            if (!prevWishlist.includes(pokemonId)) {
                return [...prevWishlist, pokemonId];
            }
            return prevWishlist;
        });
    };

    const removePokemonFromWishlist = (pokemonId: string) => {
        setWishlist(prevWishlist => {
            console.log("removePokemonFromWishlist: Removing", pokemonId, "from", prevWishlist);
            const newWishlist = prevWishlist.filter(id => id !== pokemonId);
            console.log("removePokemonFromWishlist: New wishlist after removing", newWishlist);
            return newWishlist;
        });
    };

    const isPokemonInWishlist = (pokemonId: string): boolean => {
        return wishlist.includes(pokemonId);
    };

    const clearWishlist = (): void => {
        console.log("clearWishlist: Clearing wishlist");
        setWishlist([]);
    };

    return {
        wishlist,
        addPokemonToWishlist,
        removePokemonFromWishlist,
        isPokemonInWishlist,
        clearWishlist,
    };
};
