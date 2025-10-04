import DraftCard from "../components/draftCards"
import Filter from "../components/filter"
import { useState, useEffect } from "react"
import { FilterState, DraftCardProps, LeaguePokemon } from "../api/data_interfaces"
import { getAvailablePokemon } from "../api/api"
import { useLeague } from "../context/LeagueContext"
import axios from 'axios'; // Import axios for error handling
import { WishlistDisplay } from "../components/WishlistDisplay"
import { useWishlist } from '../hooks/useWishlist';

const defaultFilters: FilterState = {
    selectedTypes: [],
    minCost: '',
    maxCost: '',
    costSortOrder: 'desc',
    sortByStat: '',
    sortOrder: 'desc',
};

export default function Draftboard() {
    const { currentLeague, loading: leagueLoading, error: leagueError } = useLeague();
    const { wishlist, addPokemonToWishlist, removePokemonFromWishlist, clearWishlist, isPokemonInWishlist } = useWishlist(currentLeague?.id || '');
    const [allPokemon, setAllPokemon] = useState<LeaguePokemon[]>([]);
    const [cards, setCards] = useState<LeaguePokemon[]>([]);
    const [filters, setFilters] = useState<FilterState>(defaultFilters);
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [pokemonLoading, setPokemonLoading] = useState<boolean>(true);
    const [pokemonError, setPokemonError] = useState<string | null>(null);
    const [currentlyFlippedCardId, setCurrentlyFlippedCardId] = useState<string | null>(null);

    const handleCardFlip = (pokemonId: string) => {
        setCurrentlyFlippedCardId(prevId => (prevId === pokemonId ? null : pokemonId));
    };

    // handleImageError function
    const handleImageError = (e: React.SyntheticEvent<HTMLImageElement, Event>) => {
        e.currentTarget.onerror = null;
        e.currentTarget.src = `https://placehold.co/150x150/cccccc/333333?text=No+Image`;
    };

    // Effect to fetch Pokemon data when the league changes
    useEffect(() => {
        // console.log("Draftboard: useEffect for fetchPokemon running. currentLeague?.id:", currentLeague?.id);
        const fetchPokemon = async () => {
            // console.log("Draftboard: fetchPokemon called.");
            if (!currentLeague?.id) {
                // console.log("Draftboard: No currentLeague ID, skipping fetchPokemon.");
                setAllPokemon([]);
                setCards([]);
                setPokemonLoading(false);
                return;
            }

            try {
                setPokemonLoading(true);
                setPokemonError(null);
                // console.log(`Draftboard: Attempting to fetch available Pokemon for league ID: ${currentLeague.id}`);
                const response = await getAvailablePokemon(currentLeague.id);
                // console.log("Draftboard: getAvailablePokemon response:", response.data);
                setAllPokemon(response.data);
                setCards(response.data); // Initialize cards with all fetched pokemon
            } catch (err) {
                if (axios.isAxiosError(err) && err.response) {
                    setPokemonError(err.response.data.error || "Failed to fetch Pokemon data.");
                } else {
                    setPokemonError("A network or unknown error occurred while fetching Pokemon data.");
                }
                console.error("Draftboard: Error fetching Pokemon data:", err);
            } finally {
                setPokemonLoading(false);
                console.log("Draftboard: fetchPokemon finished. Pokemon loading:", pokemonLoading);
            }
        };
        fetchPokemon();
    }, [currentLeague?.id]); // Re-fetch when currentLeague.id changes

    // Effect to apply filters when filters, search term, or allPokemon changes
    useEffect(() => {
        console.log("Draftboard: useEffect for applyFilter running.");
        applyFilter();
    }, [filters, searchTerm, allPokemon]);

    function applyFilter() {
        let updatedCards: LeaguePokemon[] = [...allPokemon];

        if (searchTerm.trim() !== '') {
            updatedCards = updatedCards.filter(card =>
                card.PokemonSpecies.name.toLowerCase().includes(searchTerm.toLowerCase())
            );
        }

        if (filters.selectedTypes.length > 0) {
            updatedCards = updatedCards.filter(card =>
                filters.selectedTypes.some(type =>
                    card.PokemonSpecies.types.includes(type)
                )
            );
        }

        if (filters.minCost !== '') {
            const min = parseInt(filters.minCost);
            updatedCards = updatedCards.filter(card => card.cost >= min);
        }

        if (filters.maxCost !== '') {
            const max = parseInt(filters.maxCost);
            updatedCards = updatedCards.filter(card => card.cost <= max);
        }

        // Always sort with cost as secondary sort
        updatedCards = updatedCards.sort((a, b) => {
            // If sortByStat is selected, use it as primary sort
            if (filters.sortByStat) {
                const statA = a.PokemonSpecies.stats[filters.sortByStat];
                const statB = b.PokemonSpecies.stats[filters.sortByStat];

                if (statA !== undefined && statB !== undefined) {
                    const statDiff = filters.sortOrder === 'desc' ? statB - statA : statA - statB;

                    // If stats are equal, sort by cost as tiebreaker
                    if (statDiff !== 0) {
                        return statDiff;
                    }
                }
            }

            // Primary sort by cost (if no stat selected), or secondary sort (if stats are equal)
            return filters.costSortOrder === 'desc' ? b.cost - a.cost : a.cost - b.cost;
        });

        setCards(updatedCards);
    }

    function updateFilter(key: keyof FilterState, value: any) {
        setFilters(prev => ({ ...prev, [key]: value }));
    }

    const resetAllFilters = () => {
        setFilters(defaultFilters);
        setSearchTerm('');
    };

    if (leagueLoading || pokemonLoading) {
        return (
            <div className="min-h-screen bg-[#BFC0C0] flex items-center justify-center">
                <p className="text-xl text-gray-800">Loading data...</p>
            </div>
        );
    }

    if (leagueError || pokemonError) {
        return (
            <div className="min-h-screen bg-[#BFC0C0] flex items-center justify-center">
                <p className="text-xl text-red-600">Error: {leagueError || pokemonError}</p>
            </div>
        );
    }

    if (!currentLeague) {
        return (
            <div className="min-h-screen bg-[#BFC0C0] flex items-center justify-center">
                <p className="text-xl text-gray-800">No league selected. Please select a league.</p>
            </div>
        );
    }

    const cardsToDisplay = cards.map((leaguePokemon: LeaguePokemon) => {
        // console.log("Draftboard::cardsToDisplay: leaguePokemon id, cost, pokemonSpecies", leaguePokemon.id, leaguePokemon.cost, leaguePokemon.PokemonSpecies);

        if (!leaguePokemon.PokemonSpecies || !currentLeague?.id) {
            console.warn("Draftboard: Skipping card due to missing pokemonSpecies:", leaguePokemon);
            return null;
        }

        const pokemon = leaguePokemon.PokemonSpecies;
        const name = pokemon.name.charAt(0).toUpperCase() + pokemon.name.slice(1);
        const draftCardProps: DraftCardProps = {
            key: leaguePokemon.id,
            leaguePokemonId: leaguePokemon.id,
            pokemon: {
                ...pokemon,
                name: name,
            },
            cost: leaguePokemon.cost,
            onImageError: handleImageError,
            addPokemonToWishlist: addPokemonToWishlist,
            isPokemonInWishlist: isPokemonInWishlist,
            removePokemonFromWishlist: removePokemonFromWishlist,
            isFlipped: currentlyFlippedCardId === leaguePokemon.id,
            onFlip: handleCardFlip
        };
        return <DraftCard {...draftCardProps} />;

    });

    return (
        <>
            <div className="min-h-screen bg-[#BFC0C0] ">
                <div className="flex flex-row">
                    <div className="flex flex-col w-[70%]">
                        <div className="flex flex-row m-4 p-8 pb-0 mb-2 justify-between">
                            {/* search button */}
                            <div className="relative flex text-black">
                                <input
                                    type="search"
                                    className="placeholder:text-black relative m-0 block flex-auto rounded border border-solid border-black bg-transparent bg-clip-padding px-3 py-[0.25rem] text-base font-normal leading-[1.6] text-surface outline-none transition duration-200 ease-in-out focus:z-[3] focus:border-primary focus:shadow-inset focus:outline-none motion-reduce:transition-none  dark:autofill:shadow-autofill dark:focus:border-primary"
                                    placeholder="Search"
                                    aria-label="Search"
                                    id="exampleFormControlInput2"
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                />
                                <span
                                    className="flex items-center whitespace-nowrap px-3 py-[0.25rem] text-surface dark:border-neutral-400 dark:text-white [&>svg]:h-5 [&>svg]:w-5"
                                    id="button-addon2">
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        fill="none"
                                        viewBox="0 0 24 24"
                                        strokeWidth="2"
                                        stroke="black">
                                        <path
                                            strokeLinecap="round"
                                            strokeLinejoin="round"
                                            d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
                                    </svg>
                                </span>
                            </div>
                            <Filter updateFilter={updateFilter} filters={filters} resetAllFilters={resetAllFilters} />
                        </div>
                        <div className="grid grid-cols-5 gap-4 m-4 p-6 mt-0 pt-0 pr-8 h-auto rounded-2xl">
                            {cardsToDisplay}
                        </div>
                    </div>

                    <div className="flex flex-col w-[25%] mx-auto mt-16 ml-2 h-[100%] gap-4">
                        {currentLeague?.id && <WishlistDisplay allPokemon={allPokemon} wishlist={wishlist} removePokemonFromWishlist={removePokemonFromWishlist} clearWishlist={clearWishlist} />}
                        <div className="bg-white shadow-md rounded-md overflow-hidden">
                            <div className="bg-gray-100 py-2 px-4">
                                <h2 className="text-l font-semibold text-gray-800">Your Team</h2>
                            </div>
                            <ul className="divide-y divide-gray-200">
                                <li className="flex items-center py-4 px-6">
                                    <img className="w-12 h-12 object-cover mr-4" src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/9.png" alt="User avatar"></img>
                                    <div className="flex-1">
                                        <h3 className="text-lg font-medium text-gray-800">BIG MAN BLASTOISE</h3>
                                        {/* <p className="text-gray-600 text-base">1234 points</p> */}
                                    </div>
                                </li>
                            </ul>
                        </div>
                    </div>
                </div>
            </div>
        </>
    );
}
