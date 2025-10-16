import { useState, useEffect } from 'react';

const WISHLIST_STORAGE_KEY_PREFIX = 'wishlist_';
const MAX_WISHLIST_SIZE = 15;

interface UseWishlistHook {
    wishlist: string[];
    addPokemonToWishlist: (pokemonId: string) => string[] | null;
    removePokemonFromWishlist: (pokemonId: string) => string[] | null;
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

    const addPokemonToWishlist = (pokemonId: string): string[] | null => {
        if (!leagueId) return null;
        let updatedWishlist: string[] | null = null;
        setWishlist(prevWishlist => {
            if (prevWishlist.length >= MAX_WISHLIST_SIZE) {
                console.warn("Wishlist is full. Cannot add more Pokemon.");
                updatedWishlist = prevWishlist;
                return prevWishlist;
            }
            if (!prevWishlist.includes(pokemonId)) {
                updatedWishlist = [...prevWishlist, pokemonId];
                return updatedWishlist;
            }
            updatedWishlist = prevWishlist;
            return prevWishlist;
        });
        return updatedWishlist;
    };

    const removePokemonFromWishlist = (pokemonId: string): string[] | null => {
        if (!leagueId) return null;
        let updatedWishlist: string[] | null = null;
        setWishlist(prevWishlist => {
            updatedWishlist = prevWishlist.filter(id => id !== pokemonId);
            return updatedWishlist;
        });
        return updatedWishlist;
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
