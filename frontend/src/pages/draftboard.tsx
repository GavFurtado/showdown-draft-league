import DraftCard from "../components/draftCards"
import Filter from "../components/filter"
import { useState, useEffect, useCallback, useMemo } from "react"
import { FilterState, DraftCardProps, LeaguePokemon } from "../api/data_interfaces"
import { getAllLeaguePokmeon, makePick, getDraftedPokemonByPlayer } from "../api/api"
import { useLeague } from "../context/LeagueContext"
import axios from 'axios'; // Import axios for error handling
import { WishlistDisplay } from "../components/WishlistDisplay"
import { useWishlist } from '../hooks/useWishlist';
import Modal from "../components/Modal";

const defaultFilters: FilterState = {
    selectedTypes: [],
    minCost: '',
    maxCost: '',
    costSortOrder: 'desc',
    sortByStat: '',
    sortOrder: 'desc',
};

export default function Draftboard() {
    const { currentLeague, currentDraft, currentPlayer, refetch: refetchLeague, loading: leagueLoading, error: leagueError } = useLeague();
    const { wishlist, addPokemonToWishlist, removePokemonFromWishlist, clearWishlist, isPokemonInWishlist } = useWishlist(currentLeague?.id || '');

    const [allPokemon, setAllPokemon] = useState<LeaguePokemon[]>([]);
    const [cards, setCards] = useState<LeaguePokemon[]>([]);
    const [filters, setFilters] = useState<FilterState>(defaultFilters);
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [pokemonLoading, setPokemonLoading] = useState<boolean>(true);
    const [pokemonError, setPokemonError] = useState<string | null>(null);
    const [currentlyFlippedCardId, setCurrentlyFlippedCardId] = useState<string | null>(null);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedPokemon, setSelectedPokemon] = useState<LeaguePokemon | null>(null);
    const [draftedPokemon, setDraftedPokemon] = useState<LeaguePokemon[]>([]);

    const isMyTurn = currentDraft?.currentTurnPlayerID === currentPlayer?.id;

    const fetchPokemon = useCallback(async () => {
        if (!currentLeague?.id) {
            setAllPokemon([]);
            setCards([]);
            setPokemonLoading(false);
            return;
        }
        try {
            setPokemonLoading(true);
            setPokemonError(null);
            const response = await getAllLeaguePokmeon(currentLeague.id);
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
        }
    }, [currentLeague?.id]);

    useEffect(() => {
        if (currentLeague && currentPlayer) {
            getDraftedPokemonByPlayer(currentLeague.id, currentPlayer.id)
                .then(response => setDraftedPokemon(response.data))
                .catch(error => console.error("Failed to fetch drafted pokemon:", error));
        }
    }, [currentLeague, currentPlayer, fetchPokemon]);


    const handleCardFlip = useCallback((pokemonId: string) => {
        setCurrentlyFlippedCardId(prevId => (prevId === pokemonId ? null : pokemonId));
    }, [setCurrentlyFlippedCardId]);

    const onDraft = useCallback((leaguePokemonId: string) => {
        const pokemonToDraft = allPokemon.find(p => p.id === leaguePokemonId);
        if (pokemonToDraft) {
            setSelectedPokemon(pokemonToDraft);
            setIsModalOpen(true);
        }
    }, [allPokemon, setSelectedPokemon, setIsModalOpen]);

    const handleConfirmDraft = useCallback(async () => {
        if (selectedPokemon && currentLeague && currentPlayer) {
            try {
                if (!currentDraft) {
                    console.error("Draft data is not available.");
                    return;
                }
                const pickRequest = {
                    RequestedPickCount: 1,
                    RequestedPicks: [{
                        LeaguePokemonID: selectedPokemon.id,
                        DraftPickNumber: currentDraft.currentPickOnClock
                    }]
                };
                await makePick(currentLeague.id, pickRequest);
                fetchPokemon();
                refetchLeague();
            } catch (error) {
                console.error("Failed to make a pick:", error);
            } finally {
                setIsModalOpen(false);
                setSelectedPokemon(null);
            }
        }
    }, [selectedPokemon, currentLeague, currentPlayer, currentDraft, fetchPokemon, refetchLeague, setIsModalOpen, setSelectedPokemon]);

    const handleCancelDraft = useCallback(() => {
        setIsModalOpen(false);
        setSelectedPokemon(null);
    }, [setIsModalOpen, setSelectedPokemon]);


    // handleImageError function
    const handleImageError = useCallback((e: React.SyntheticEvent<HTMLImageElement, Event>) => {
        e.currentTarget.onerror = null;
        e.currentTarget.src = `https://placehold.co/150x150/cccccc/333333?text=No+Image`;
    }, []);

    useEffect(() => {
        fetchPokemon();
    }, [fetchPokemon]); // Re-fetch when fetchPokemon changes

    // Effect to apply filters when filters, search term, or allPokemon changes
    useEffect(() => {
        console.log("Draftboard: useEffect for applyFilter running.");
        applyFilter();
    }, [filters, searchTerm, allPokemon]);

    const applyFilter = useCallback(() => {
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
    }, [allPokemon, filters, searchTerm, setCards]);

    const updateFilter = useCallback((key: keyof FilterState, value: any) => {
        setFilters(prev => ({ ...prev, [key]: value }));
    }, [setFilters]);

    const resetAllFilters = useCallback(() => {
        setFilters(defaultFilters);
        setSearchTerm('');
    }, [setFilters, setSearchTerm]);

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
            onFlip: handleCardFlip,
            isDraftable: isMyTurn && leaguePokemon.isAvailable,
            onDraft: onDraft,
        };
        const { key, ...rest } = draftCardProps;
        return <DraftCard key={key} {...rest} />;
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
                        {currentDraft && (
                            <div className="p-4 bg-white rounded-lg shadow-md">
                                <h2 className="text-lg font-bold mb-2">Draft Status</h2>
                                <p>Round: {currentDraft?.currentRound}</p>
                                <p>Pick: {currentDraft?.currentPickInRound}</p>
                                <p>Player on clock: {currentDraft?.CurrentTurnPlayer?.inLeagueName}</p>                                </div>
                        )}
                        {currentLeague?.id && (
                            <WishlistDisplay
                                allPokemon={allPokemon}
                                wishlist={wishlist}
                                removePokemonFromWishlist={removePokemonFromWishlist}
                                clearWishlist={clearWishlist}
                                isMyTurn={isMyTurn ?? false}
                                onDraft={onDraft}
                            />
                        )}
                        <div className="bg-white shadow-md rounded-md overflow-hidden">
                            <div className="bg-gray-100 py-2 px-4">
                                <h2 className="text-l font-semibold text-gray-800">Your Team</h2>
                            </div>
                            <ul className="divide-y divide-gray-200">
                                {draftedPokemon.map(p => (
                                    <li key={p.id} className="flex items-center py-4 px-6">
                                        <img className="w-12 h-12 object-cover mr-4" src={p.PokemonSpecies.sprites.front_default} alt={p.PokemonSpecies.name}></img>
                                        <div className="flex-1">
                                            <h3 className="text-lg font-medium text-gray-800">{p.PokemonSpecies.name}</h3>
                                        </div>
                                    </li>
                                ))}
                            </ul>
                        </div>
                    </div>
                </div>
            </div>
            {selectedPokemon && (
                <Modal isOpen={isModalOpen} onClose={handleCancelDraft} title="Confirm Draft">
                    <h2 className="text-xl font-bold mb-4">Confirm Draft</h2>
                    <p>Are you sure you want to draft {selectedPokemon.PokemonSpecies.name} for {selectedPokemon.cost} points?</p>
                    <div className="mt-6 flex justify-end gap-4">
                        <button onClick={handleCancelDraft} className="px-4 py-2 bg-gray-300 rounded hover:bg-gray-400">Cancel</button>
                        <button onClick={handleConfirmDraft} className="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600">Confirm</button>
                    </div>
                </Modal>
            )}
        </>
    );
}