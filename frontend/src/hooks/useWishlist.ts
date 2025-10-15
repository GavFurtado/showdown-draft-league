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
    const storageKey = `${WISHLIST_STORAGE_KEY_PREFIX}${leagueId}`;

    const [wishlist, setWishlist] = useState<string[]>([]);

    // Effect to load wishlist from localStorage when leagueId becomes available
    useEffect(() => {
        if (!leagueId) {
            return;
        }
        try {
            const storedWishlist = localStorage.getItem(storageKey);
            const initialWishlist = storedWishlist ? JSON.parse(storedWishlist) : [];
            setWishlist(initialWishlist);
        } catch (error) {
            console.error("Error parsing wishlist from localStorage (in useEffect)", error);
            setWishlist([]);
        }
    }, [leagueId, storageKey]);

    // Effect to save wishlist to localStorage whenever it changes and leagueId is available
    useEffect(() => {
        if (!leagueId) {
            return;
        }
        try {
            localStorage.setItem(storageKey, JSON.stringify(wishlist));
        } catch (error) {
            console.error("Error saving wishlist to localStorage", error);
        }
    }, [wishlist, storageKey, leagueId]);

    const addPokemonToWishlist = (pokemonId: string) => {
        if (!leagueId) return; // Prevent actions if leagueId is not set
        setWishlist(prevWishlist => {
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
        if (!leagueId) return; // Prevent actions if leagueId is not set
        setWishlist(prevWishlist => {
            const newWishlist = prevWishlist.filter(id => id !== pokemonId);
            return newWishlist;
        });
    };

    const isPokemonInWishlist = (pokemonId: string): boolean => {
        return leagueId ? wishlist.includes(pokemonId) : false;
    };

    const clearWishlist = (): void => {
        if (!leagueId) return; // Prevent actions if leagueId is not set
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
