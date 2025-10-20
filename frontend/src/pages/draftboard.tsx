import DraftCard from "../components/draftCards"
import Filter from "../components/filter"
import { useState, useEffect, useCallback } from "react"
import { formatPokemonName } from "../utils/nameFormatter";
import { FilterState, DraftCardProps, LeaguePokemon, DraftedPokemon, Player } from "../api/data_interfaces"
import { getAllLeaguePokmeon, makePick, getDraftedPokemonByPlayer, skipPick, getPlayersByLeague } from "../api/api"
import { useLeague } from "../context/LeagueContext"
import { WishlistDisplay } from "../components/WishlistDisplay"
import { useWishlist } from '../hooks/useWishlist';
import { useDraftTimer } from '../hooks/useDraftTimer';
import Modal from "../components/Modal";
import { PokemonRosterList } from "../components/PokemonRosterList";
import Layout from "../components/Layout";
import FullScreenModal from "../components/FullScreenModal";
import FullScreenDraftModal from "../components/FullScreenDraftModal";

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
    const [draftedPokemon, setDraftedPokemon] = useState<DraftedPokemon[]>([]);
    const [draftedPokemonError, setDraftedPokemonError] = useState<string | null>(null);
    const [isInfoOpen, setIsInfoOpen] = useState(false);
    const [pendingPicks, setPendingPicks] = useState<LeaguePokemon[]>([]);
    const [isFullScreenModalOpen, setIsFullScreenModalOpen] = useState(false);
    const [allPlayersDraftedPokemon, setAllPlayersDraftedPokemon] = useState<{ [key: string]: DraftedPokemon[] }>({});
    const [leaguePlayers, setLeaguePlayers] = useState<Player[]>([]);

    const isMyTurn = currentDraft?.CurrentTurnPlayerID === currentPlayer?.ID;
    const accumulatedPicks = currentDraft?.PlayersWithAccumulatedPicks?.[currentPlayer?.ID || ''] || [];
    const pickNumbersToUse = [currentDraft?.CurrentPickOnClock, ...accumulatedPicks].filter(n => n !== undefined) as number[];
    const numberOfPicksAvailable = pickNumbersToUse.length;

    const fetchPokemon = useCallback(async () => {
        if (!currentLeague?.ID) return;
        try {
            setPokemonLoading(true);
            setPokemonError(null);
            const response = await getAllLeaguePokmeon(currentLeague.ID);
            setAllPokemon(response.data);
            setCards(response.data);
        } catch (err) {
            setPokemonError("Failed to fetch Pokemon data.");
            console.error("Draftboard: Error fetching Pokemon data:", err);
        } finally {
            setPokemonLoading(false);
        }
    }, [currentLeague?.ID]);

    const fetchDraftedPokemon = useCallback(async () => {
        if (currentLeague && currentPlayer) {
            try {
                const response = await getDraftedPokemonByPlayer(currentLeague.ID, currentPlayer.ID);
                setDraftedPokemon(response.data);
                setDraftedPokemonError(null);
            } catch (error) {
                console.error("Failed to fetch drafted pokemon:", error);
                setDraftedPokemonError("Failed to fetch drafted Pokemon.");
            }
        }
    }, [currentLeague, currentPlayer]);

    const fetchAllPlayersDraftedPokemon = useCallback(async () => {
        if (currentLeague && leaguePlayers.length > 0) {
            try {
                const promises = leaguePlayers.map(player => getDraftedPokemonByPlayer(currentLeague.ID, player.ID));
                const results = await Promise.all(promises);
                const newAllPlayersDraftedPokemon = results.reduce((acc, result, index) => {
                    const player = leaguePlayers[index];
                    acc[player.ID] = result.data;
                    return acc;
                }, {} as { [key: string]: DraftedPokemon[] });
                setAllPlayersDraftedPokemon(newAllPlayersDraftedPokemon);
            } catch (error) {
                console.error("Failed to fetch all players drafted pokemon:", error);
            }
        }
    }, [currentLeague, leaguePlayers]);

    useEffect(() => {
        fetchPokemon();
        fetchDraftedPokemon();
    }, [fetchPokemon, fetchDraftedPokemon]);

    useEffect(() => {
        if (currentLeague?.ID) {
            getPlayersByLeague(currentLeague.ID)
                .then(response => setLeaguePlayers(response.data))
                .catch(error => console.error("Failed to fetch league players:", error));
        }
    }, [currentLeague?.ID]);

    useEffect(() => {
        if (isFullScreenModalOpen) {
            fetchAllPlayersDraftedPokemon();
        }
    }, [isFullScreenModalOpen, fetchAllPlayersDraftedPokemon]);

    const handleCardFlip = useCallback((pokemonId: string) => {
        setCurrentlyFlippedCardId(prevId => (prevId === pokemonId ? null : pokemonId));
    }, []);

    const onDraft = useCallback((leaguePokemonId: string) => {
        const pokemonToDraft = allPokemon.find(p => p.ID === leaguePokemonId);
        if (!pokemonToDraft || !isMyTurn) return;

        setPendingPicks(prevPicks => {
            const isAlreadyPending = prevPicks.some(p => p.ID === leaguePokemonId);
            if (isAlreadyPending) {
                return prevPicks.filter(p => p.ID !== leaguePokemonId);
            }
            if (prevPicks.length < numberOfPicksAvailable) {
                return [...prevPicks, pokemonToDraft];
            }
            return prevPicks;
        });
    }, [allPokemon, isMyTurn, numberOfPicksAvailable]);

    const handleConfirmDraft = useCallback(async () => {
        if (pendingPicks.length > 0 && currentLeague && currentPlayer && currentDraft) {
            try {
                const pickRequest = {
                    RequestedPickCount: pendingPicks.length,
                    RequestedPicks: pendingPicks.map((pick, index) => ({
                        LeaguePokemonID: pick.ID,
                        DraftPickNumber: pickNumbersToUse[index]
                    }))
                };
                await makePick(currentLeague.ID, pickRequest);
                refetchLeague();
                fetchDraftedPokemon();
                setPendingPicks([]);
            } catch (error) {
                console.error("Failed to make a pick:", error);
            } finally {
                setIsModalOpen(false);
            }
        }
    }, [pendingPicks, currentLeague, currentPlayer, currentDraft, pickNumbersToUse, refetchLeague, fetchDraftedPokemon]);

    const handleCancelDraft = useCallback(() => setIsModalOpen(false), []);

    const handleSkipTurn = useCallback(async () => {
        if (currentLeague && isMyTurn) {
            try {
                await skipPick(currentLeague.ID);
                refetchLeague();
            } catch (error) {
                console.error("Failed to skip turn:", error);
            }
        }
    }, [currentLeague, isMyTurn, refetchLeague]);

    const handleImageError = useCallback((e: React.SyntheticEvent<HTMLImageElement, Event>) => {
        e.currentTarget.onerror = null;
        e.currentTarget.src = `https://placehold.co/150x150/cccccc/333333?text=No+Image`;
    }, []);

    useEffect(() => {
        applyFilter();
    }, [filters, searchTerm, allPokemon]);

    const applyFilter = useCallback(() => {
        let updatedCards: LeaguePokemon[] = [...allPokemon];
        if (searchTerm.trim() !== '') {
            updatedCards = updatedCards.filter(card => card.PokemonSpecies.Name.toLowerCase().includes(searchTerm.toLowerCase()));
        }
        if (filters.selectedTypes.length > 0) {
            updatedCards = updatedCards.filter(card => filters.selectedTypes.some(type => card.PokemonSpecies.Types.includes(type)));
        }
        if (filters.minCost !== '') {
            const min = parseInt(filters.minCost);
            updatedCards = updatedCards.filter(card => card.Cost >= min);
        }
        if (filters.maxCost !== '') {
            const max = parseInt(filters.maxCost);
            updatedCards = updatedCards.filter(card => card.Cost <= max);
        }
        updatedCards.sort((a, b) => {
            if (filters.sortByStat) {
                const statA = a.PokemonSpecies.Stats[filters.sortByStat];
                const statB = b.PokemonSpecies.Stats[filters.sortByStat];
                if (statA !== undefined && statB !== undefined) {
                    const statDiff = filters.sortOrder === 'desc' ? statB - statA : statA - statB;
                    if (statDiff !== 0) return statDiff;
                }
            }
            return filters.costSortOrder === 'desc' ? b.Cost - a.Cost : a.Cost - b.Cost;
        });
        setCards(updatedCards);
    }, [allPokemon, filters, searchTerm]);

    const updateFilter = useCallback((key: keyof FilterState, value: any) => {
        setFilters(prev => ({ ...prev, [key]: value }));
    }, []);

    const resetAllFilters = useCallback(() => {
        setFilters(defaultFilters);
        setSearchTerm('');
    }, []);

    const remainingPoints = currentPlayer?.DraftPoints ?? 0;
    const skipsAllowed = (currentLeague?.MaxPokemonPerPlayer ?? 0) - (currentLeague?.MinPokemonPerPlayer ?? 0);
    const skipsUsed = accumulatedPicks.length;
    const skipsLeft = skipsAllowed - skipsUsed;

    if (leagueError || pokemonError) {
        return (
            <div className="min-h-screen bg-background-main flex items-center justify-center">
                <p className="text-xl text-red-600">Error: {leagueError || pokemonError}</p>
            </div>
        );
    }

    const cardsToDisplay = cards.map(leaguePokemon => {
        if (!leaguePokemon.PokemonSpecies || !currentLeague?.ID) return null;
        const pokemon = leaguePokemon.PokemonSpecies;
        const name = formatPokemonName(pokemon.Name);
        const isPending = pendingPicks.some(p => p.ID === leaguePokemon.ID);

        const draftCardProps: DraftCardProps = {
            key: leaguePokemon.ID,
            leaguePokemonId: leaguePokemon.ID,
            pokemon: { ...pokemon, Name: name },
            cost: leaguePokemon.Cost,
            onImageError: handleImageError,
            addPokemonToWishlist: addPokemonToWishlist,
            isPokemonInWishlist: isPokemonInWishlist,
            removePokemonFromWishlist: removePokemonFromWishlist,
            isFlipped: currentlyFlippedCardId === leaguePokemon.ID,
            onFlip: handleCardFlip,
            isDraftable: isMyTurn && leaguePokemon.IsAvailable && !isPending,
            onDraft: onDraft,
            isAvailable: leaguePokemon.IsAvailable,
            isMyTurn: isMyTurn ?? false,
        };
        const { key, ...rest } = draftCardProps;
        return <DraftCard key={key} {...rest} />;
    });

    return (
        <Layout variant="full">
            <div className="flex flex-col md:flex-row">
                {/* Main content area */}
                <div className="flex flex-col w-full md:w-[75%] order-2 md:order-1 p-4 sm:p-6">
                    <div className="flex flex-row pb-0 mb-2 justify-between">
                        <div className="relative flex text-black">
                            <input
                                type="search"
                                className="placeholder:text-black relative m-0 block flex-auto rounded border border-solid 
                                border-black bg-transparent bg-clip-padding px-3 py-[0.25rem] text-base font-normal leading-[1.6]
                                text-surface outline-none transition duration-200 ease-in-out focus:z-[3] focus:border-primary 
                                focus:shadow-inset focus:outline-none motion-reduce:transition-none"
                                placeholder="Search" value={searchTerm} onChange={(e) => setSearchTerm(e.target.value)} />
                            <span className="flex items-center whitespace-nowrap px-3 py-[0.25rem] text-surface [&>svg]:h-5 [&>svg]:w-5" id="button-addon2">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="2" stroke="black"><path strokeLinecap="round" strokeLinejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
                            </span>
                        </div>
                        <Filter updateFilter={updateFilter} filters={filters} resetAllFilters={resetAllFilters} />
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6 gap-2 h-auto rounded-2xl">
                        {cardsToDisplay}
                    </div>
                </div>

                {/* Right-hand content */}
                <div className="flex flex-col w-full md:w-[20%] md:mt-16 h-auto gap-4 p-4 md:p-0 order-1 md:order-2">
                    {/* only shows for small viewport */}
                    <button className="md:hidden bg-gray-200 p-2 rounded-md" onClick={() => setIsInfoOpen(!isInfoOpen)}>
                        {isInfoOpen ? 'Hide' : 'Show'} Draft Info
                    </button>
                    <div className={`${isInfoOpen ? 'block' : 'hidden'} md:block space-y-4`}>
                        {shouldShowDraftStatus && currentDraft && (
                            <div className="p-4 bg-background-surface rounded-lg shadow-md">
                                <div className="flex justify-between items-center mb-2">
                                    <h2 className="text-lg font-bold">Draft Status</h2>
                                    <button className="bg-gray-200 p-2 rounded-md" onClick={() => setIsFullScreenModalOpen(!isFullScreenModalOpen)}>
                                        Show Full Draft
                                    </button>
                                </div>
                                <div className="space-y-2 text-sm">
                                    <div className="flex justify-between items-center">
                                        <span className="font-semibold">Player Turn:</span>
                                        <span className="truncate">{currentDraft.CurrentTurnPlayer?.TeamName || currentDraft.CurrentTurnPlayer?.InLeagueName}{isMyTurn && <span className="text-green-500 font-bold ml-2">(You)</span>}</span>
                                    </div>
                                    <div className="flex justify-between items-center">
                                        <span className="font-semibold">Time Left:</span>
                                        <span className="font-mono">{currentDraft.Status === 'ONGOING' ? timeRemaining : "Draft Paused"}</span>
                                    </div>
                                    {nextTurnInXTurns !== null && (
                                        <div className="flex justify-between items-center">
                                            <span className="font-semibold">Your Next Turn:</span>
                                            <span>in {nextTurnInXTurns} turns</span>
                                        </div>
                                    )}
                                    {isMyTurn && (
                                        <div className="flex justify-between items-center">
                                            <span className="font-semibold">Action:</span>
                                            <button onClick={handleSkipTurn} disabled={skipsLeft <= 0} className="flex flex-col items-center px-3 py-1.5 rounded-md transition-colors duration-150 disabled:bg-gray-300 disabled:cursor-not-allowed bg-yellow-500 text-white hover:bg-yellow-600">
                                                <span className="font-medium">Skip Turn</span>
                                                <span className="text-xs opacity-80">({skipsLeft} left)</span>
                                            </button>
                                        </div>
                                    )}
                                </div>
                            </div>
                        )}
                        <WishlistDisplay {...{ allPokemon, wishlist, removePokemonFromWishlist, clearWishlist, isMyTurn: isMyTurn ?? false, onDraft }} />
                        {/* Your Team / Pending Picks */}
                        <div className="bg-background-surface shadow-md rounded-md overflow-hidden">
                            <div className="bg-gray-100 py-2 px-4 flex justify-between items-center">
                                <h2 className="text-l font-semibold text-gray-800">Your Team ({draftedPokemon.length})</h2>
                                <span className="text-sm font-medium text-gray-600">Points Left: {remainingPoints}</span>
                            </div>
                            {/* Pending Picks Display */}
                            <div className="p-2 space-y-2">
                                {isMyTurn && pendingPicks.length > 0 && (
                                    <>
                                        <h3 className="text-sm font-bold px-1 pt-1">Pending Picks ({pendingPicks.length}/{numberOfPicksAvailable})</h3>
                                        <PokemonRosterList
                                            roster={pendingPicks}
                                            rosterType="pendingPick"
                                            onRemove={onDraft}
                                            showRemoveButton={true}
                                            bgColor="bg-yellow-100 border border-yellow-400"
                                        />
                                        <hr className="my-2 border-gray-300" />
                                    </>
                                )}
                                {draftedPokemonError && <p className="text-red-600">Error: {draftedPokemonError}</p>}
                                {draftedPokemon.length > 0 ? (
                                    <PokemonRosterList
                                        roster={draftedPokemon}
                                        rosterType="drafted"
                                        bgColor="bg-gray-200"
                                    />
                                ) : (
                                    !isMyTurn || pendingPicks.length === 0 && <p className="text-gray-600 p-2">Your team is empty.</p>
                                )}
                                {isMyTurn && pendingPicks.length > 0 && (
                                    <button onClick={() => setIsModalOpen(true)} className="w-full mt-2 px-4 py-2 text-sm font-medium rounded-md transition-colors duration-150 bg-accent-primary text-text-on-accent hover:bg-accent-primary-hover">
                                        Review & Confirm Picks
                                    </button>
                                )}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            {isModalOpen && (
                <Modal isOpen={isModalOpen} onClose={handleCancelDraft} title={`Confirm Your Picks (${pendingPicks.length})`}>
                    <div className="space-y-4">
                        <table className="min-w-full text-left text-sm">
                            <thead className="border-b border-gray-200">
                                <tr>
                                    <th className="p-2 font-semibold">Pick #</th>
                                    <th className="p-2 font-semibold">Pokémon</th>
                                    <th className="p-2 font-semibold">Cost</th>
                                </tr>
                            </thead>
                            <tbody>
                                {pendingPicks.map((p, index) => (
                                    <tr key={p.ID} className="border-t border-gray-200">
                                        <td className="p-2">{pickNumbersToUse[index]}</td>
                                        <td className="p-2">{formatPokemonName(p.PokemonSpecies.Name)}</td>
                                        <td className="p-2">{p.Cost}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                        {/* New section for cost summary */}
                        <div className="mt-4 pt-4 border-t border-gray-200 space-y-2 text-base">
                            <div className="flex justify-between">
                                <span>Your Current Points:</span>
                                <span className="font-semibold">{remainingPoints}</span>
                            </div>
                            <div className="flex justify-between">
                                <span>Total Cost of Pending Picks:</span>
                                <span className="font-semibold">{pendingPicks.reduce((sum, p) => sum + p.Cost, 0)}</span>
                            </div>
                            <div className="flex justify-between font-bold">
                                <span>Points After Draft:</span>
                                <span className={remainingPoints - pendingPicks.reduce((sum, p) => sum + p.Cost, 0) < 0 ? 'text-red-500' : 'text-green-500'}>
                                    {remainingPoints - pendingPicks.reduce((sum, p) => sum + p.Cost, 0)}
                                </span>
                            </div>
                        </div>
                        <div className="mt-6 flex justify-end gap-4">
                            <button onClick={handleCancelDraft} className="px-4 py-2 bg-gray-400 rounded hover:bg-gray-400">Cancel</button>
                            <button
                                onClick={handleConfirmDraft}
                                disabled={remainingPoints - pendingPicks.reduce((sum, p) => sum + p.Cost, 0) < 0}
                                className={`px-4 py-2 rounded text-text-on-accent border-accent-primary ${remainingPoints - pendingPicks.reduce((sum, p) => sum + p.Cost, 0) < 0 ? 'bg-gray-400 cursor-not-allowed' : 'bg-accent-primary hover:bg-accent-primary-hover'}`}
                            >
                                Confirm
                            </button>
                        </div>
                    </div>
                </Modal>
            )}
            {isFullScreenModalOpen && (
                <FullScreenDraftModal
                    isOpen={isFullScreenModalOpen}
                    onClose={() => setIsFullScreenModalOpen(false)}
                    title="Full Draft View"
                    leaguePlayers={leaguePlayers}
                    allPlayersDraftedPokemon={allPlayersDraftedPokemon}
                    currentDraft={currentDraft}
                    currentPlayer={currentPlayer}
                    currentLeague={currentLeague}
                />
            )}
        </Layout>
    );
}

