import DraftCard from "../components/draftCards"
import Filter from "../components/filter"
import { useState, useEffect, useCallback } from "react"
import { FilterState, DraftCardProps, LeaguePokemon, DraftedPokemon } from "../api/data_interfaces"
import { getAllLeaguePokmeon, makePick, getDraftedPokemonByPlayer } from "../api/api"
import { useLeague } from "../context/LeagueContext"
import axios from 'axios';
import { WishlistDisplay } from "../components/WishlistDisplay"
import { useWishlist } from '../hooks/useWishlist';
import { useDraftTimer } from '../hooks/useDraftTimer';
import Modal from "../components/Modal";
import { PokemonListItem } from "../components/PokemonListItem";

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
    const { wishlist, addPokemonToWishlist, removePokemonFromWishlist, clearWishlist, isPokemonInWishlist } = useWishlist(currentLeague?.ID || '');
    const { timeRemaining, shouldShowDraftStatus, nextTurnInXTurns } = useDraftTimer(currentDraft, currentLeague, currentPlayer, currentLeague?.Players);

    const [allPokemon, setAllPokemon] = useState<LeaguePokemon[]>([]);
    const [cards, setCards] = useState<LeaguePokemon[]>([]);
    const [filters, setFilters] = useState<FilterState>(defaultFilters);
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [pokemonLoading, setPokemonLoading] = useState<boolean>(true);
    const [pokemonError, setPokemonError] = useState<string | null>(null);
    const [currentlyFlippedCardId, setCurrentlyFlippedCardId] = useState<string | null>(null);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [selectedPokemon, setSelectedPokemon] = useState<LeaguePokemon | null>(null);
    const [draftedPokemon, setDraftedPokemon] = useState<DraftedPokemon[]>([]);
    const [draftedPokemonError, setDraftedPokemonError] = useState<string | null>(null);
    const [isInfoOpen, setIsInfoOpen] = useState(false);

    const isMyTurn = currentDraft?.CurrentTurnPlayerID === currentPlayer?.ID;

    const fetchPokemon = useCallback(async () => {
        if (!currentLeague?.ID) {
            setAllPokemon([]);
            setCards([]);
            setPokemonLoading(false);
            return;
        }
        try {
            setPokemonLoading(true);
            setPokemonError(null);
            const response = await getAllLeaguePokmeon(currentLeague.ID);
            // Prevent unnecessary re-renders if data is the same
            if (JSON.stringify(response.data) === JSON.stringify(allPokemon)) {
                setPokemonLoading(false);
                return;
            }
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
    }, [currentLeague?.ID]);

    useEffect(() => {
        if (currentLeague && currentPlayer) {
            getDraftedPokemonByPlayer(currentLeague.ID, currentPlayer.ID)
                .then(response => {
                    setDraftedPokemon(response.data);
                    setDraftedPokemonError(null); // Clear error on success
                })
                .catch(error => {
                    console.error("Failed to fetch drafted pokemon:", error);
                    if (axios.isAxiosError(error) && error.response) {
                        setDraftedPokemonError(error.response.data.error || "Failed to fetch drafted Pokemon.");
                    } else {
                        setDraftedPokemonError("A network or unknown error occurred while fetching drafted Pokemon.");
                    }
                });
        }
    }, [currentLeague, currentPlayer]);


    const handleCardFlip = useCallback((pokemonId: string) => {
        setCurrentlyFlippedCardId(prevId => (prevId === pokemonId ? null : pokemonId));
    }, [setCurrentlyFlippedCardId]);

    const onDraft = useCallback((leaguePokemonId: string) => {
        const pokemonToDraft = allPokemon.find(p => p.ID === leaguePokemonId);
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
                        LeaguePokemonID: selectedPokemon.ID,
                        DraftPickNumber: currentDraft.CurrentPickOnClock
                    }]
                };
                await makePick(currentLeague.ID, pickRequest);
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
                card.PokemonSpecies.Name.toLowerCase().includes(searchTerm.toLowerCase())
            );
        }

        if (filters.selectedTypes.length > 0) {
            updatedCards = updatedCards.filter(card =>
                filters.selectedTypes.some(type =>
                    card.PokemonSpecies.Types.includes(type)
                )
            );
        }

        if (filters.minCost !== '') {
            const min = parseInt(filters.minCost);
            updatedCards = updatedCards.filter(card => card.Cost >= min);
        }

        if (filters.maxCost !== '') {
            const max = parseInt(filters.maxCost);
            updatedCards = updatedCards.filter(card => card.Cost <= max);
        }

        // Always sort with cost as secondary sort
        updatedCards = updatedCards.sort((a, b) => {
            // If sortByStat is selected, use it as primary sort
            if (filters.sortByStat) {
                const statA = a.PokemonSpecies.Stats[filters.sortByStat];
                const statB = b.PokemonSpecies.Stats[filters.sortByStat];

                if (statA !== undefined && statB !== undefined) {
                    const statDiff = filters.sortOrder === 'desc' ? statB - statA : statA - statB;

                    // If stats are equal, sort by cost as tiebreaker
                    if (statDiff !== 0) {
                        return statDiff;
                    }
                }
            }

            // Primary sort by cost (if no stat selected), or secondary sort (if stats are equal)
            return filters.costSortOrder === 'desc' ? b.Cost - a.Cost : a.Cost - b.Cost;
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

    const remainingPoints = currentPlayer?.DraftPoints ?? 0;

    const isDraftInfoNotFoundError = leagueError?.includes("drafted information not found for league");

    if ((leagueError && !isDraftInfoNotFoundError) || pokemonError) {
        return (
            <div className="min-h-screen bg-[#BFC0C0] flex items-center justify-center">
                <p className="text-xl text-red-600">Error: {leagueError || pokemonError}</p>
            </div>
        );
    }

    const cardsToDisplay = cards.map((leaguePokemon: LeaguePokemon) => {
        // console.log("Draftboard::cardsToDisplay: leaguePokemon id, cost, pokemonSpecies", leaguePokemon.id, leaguePokemon.cost, leaguePokemon.PokemonSpecies);

        if (!leaguePokemon.PokemonSpecies || !currentLeague?.ID) {
            console.warn("Draftboard: Skipping card due to missing pokemonSpecies:", leaguePokemon);
            return null;
        }

        const pokemon = leaguePokemon.PokemonSpecies;
        const name = pokemon.Name.charAt(0).toUpperCase() + pokemon.Name.slice(1);
        const draftCardProps: DraftCardProps = {
            key: leaguePokemon.ID,
            leaguePokemonId: leaguePokemon.ID,
            pokemon: {
                ...pokemon,
                Name: name,
            },
            cost: leaguePokemon.Cost,
            onImageError: handleImageError,
            addPokemonToWishlist: addPokemonToWishlist,
            isPokemonInWishlist: isPokemonInWishlist,
            removePokemonFromWishlist: removePokemonFromWishlist,
            isFlipped: currentlyFlippedCardId === leaguePokemon.ID,
            onFlip: handleCardFlip,
            isDraftable: isMyTurn && leaguePokemon.IsAvailable,
            onDraft: onDraft,
            isAvailable: leaguePokemon.IsAvailable,
            isMyTurn: isMyTurn ?? false,
        };
        const { key, ...rest } = draftCardProps;
        return <DraftCard key={key} {...rest} />;
    });

    return (
        <>
            {isDraftInfoNotFoundError && (
                <div className="bg-yellow-100 border-l-4 border-yellow-500 text-yellow-700 p-4 mb-4" role="alert">
                    <p className="font-bold">Warning</p>
                    <p>{leagueError}</p>
                </div>
            )}
            <div className="min-h-screen bg-[#BFC0C0] ">
                <div className="flex flex-col md:flex-row">
                    <div className="flex flex-col w-full md:w-[75%] order-2 md:order-1 p-4 sm:p-6"> {/* Main content area */}
                        <div className="flex flex-row pb-0 mb-2 justify-between">
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
                        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6 gap-2 h-auto rounded-2xl">
                            {cardsToDisplay}
                        </div>
                    </div>

                    <div className="flex flex-col w-full md:w-[20%] md:mt-16 h-auto gap-4 p-4 md:p-0 order-1 md:order-2"> {/* Right-hand content */}
                        <button
                            className="md:hidden bg-gray-200 p-2 rounded-md"
                            onClick={() => setIsInfoOpen(!isInfoOpen)}
                        >
                            {isInfoOpen ? 'Hide' : 'Show'} Draft Info
                        </button>
                        <div className={`${isInfoOpen ? 'block' : 'hidden'} md:block space-y-4`}>
                            {shouldShowDraftStatus && currentDraft && (
                                <div className="p-4 bg-white rounded-lg shadow-md">
                                    <h2 className="text-lg font-bold mb-2">Draft Status</h2>

                                    {currentDraft.Status === 'PENDING' && (
                                        <p className="text-2s font-mono">Draft has not started yet</p>
                                    )}

                                    {currentDraft.Status !== 'PENDING' && currentDraft.Status !== 'COMPLETED' && (
                                        <>
                                            <p className="flex justify-between items-center">
                                                <span>Player Turn: {currentDraft.CurrentTurnPlayer?.TeamName || currentDraft.CurrentTurnPlayer?.InLeagueName || currentDraft.CurrentTurnPlayerID}</span>
                                                {isMyTurn && <span className="text-green-500">(You)</span>}
                                            </p>
                                            <p className="text-2m font-mono">
                                                Time Left: {currentDraft.Status === 'ONGOING' ? timeRemaining : "Draft Paused"}
                                            </p>
                                            {nextTurnInXTurns !== null && (
                                                <p className="text-sm">Your next turn is in {nextTurnInXTurns} turns</p>
                                            )}
                                        </>
                                    )}

                                    {currentDraft.Status === 'COMPLETED' && (
                                        <p className="text-xl">Draft Completed</p>
                                    )}
                                </div>
                            )}
                            {currentLeague?.ID && (
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
                                <div className="bg-gray-100 py-2 px-4 flex justify-between items-center">
                                    <h2 className="text-l font-semibold text-gray-800">Your Team</h2>
                                    {currentLeague && <span className="text-sm font-medium text-gray-600">Points Left: {remainingPoints}</span>}
                                </div>
                                <div className="p-2 space-y-2">
                                    {draftedPokemonError ? (
                                        <p className="text-red-600">Error: {draftedPokemonError}</p>
                                    ) : draftedPokemon.length > 0 ? (
                                        draftedPokemon.map(p => {
                                            const fullLeaguePokemon = allPokemon.find(lp => lp.ID === p.LeaguePokemonID);
                                            const cost = fullLeaguePokemon ? fullLeaguePokemon.Cost : null;
                                            return (
                                                <PokemonListItem
                                                    key={p.ID}
                                                    pokemon={p.PokemonSpecies}
                                                    cost={cost}
                                                    leaguePokemonId={p.LeaguePokemonID}
                                                    pickNumber={p.DraftPickNumber}
                                                    bgColor="bg-gray-200"
                                                />
                                            );
                                        })
                                    ) : (
                                        <p className="text-gray-600 p-2">Your team is empty.</p>
                                    )}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            {selectedPokemon && (
                <Modal isOpen={isModalOpen} onClose={handleCancelDraft} title="Confirm Draft">
                    <h2 className="text-xl font-bold mb-4">Confirm Draft</h2>
                    <p>Are you sure you want to draft {selectedPokemon.PokemonSpecies.Name} for {selectedPokemon.Cost} points?</p>
                    <div className="mt-6 flex justify-end gap-4">
                        <button onClick={handleCancelDraft} className="px-4 py-2 bg-gray-300 rounded hover:bg-gray-400">Cancel</button>
                        <button onClick={handleConfirmDraft} className="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600">Confirm</button>
                    </div>
                </Modal>
            )}
        </>
    );
}

