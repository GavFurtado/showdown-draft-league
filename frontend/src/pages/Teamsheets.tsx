import { useState, useEffect, useMemo } from 'react';
import Layout from '../components/Layout';
import { useLeague } from '../context/LeagueContext';
import { getPlayersByLeague } from '../api/api';
import { Player } from '../api/data_interfaces';
import axios from 'axios';
import { PokemonRosterList } from '../components/PokemonRosterList';
import { DefensiveTypeChart } from '../components/DefensiveTypeChart';
import TeamPokemonView from '../components/TeamPokemonView';
import SpeedTable from '../components/SpeedTable';
import { useDraftHistory } from '../context/DraftHistoryContext';

// Main Teamsheets Page Component
const Teamsheets: React.FC = () => {
    const { currentLeague, currentDraft, loading: leagueLoading } = useLeague();
    const { draftHistory, loading: historyLoading, error: historyError } = useDraftHistory();

    const [players, setPlayers] = useState<Player[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [selectedPlayer, setSelectedPlayer] = useState<Player | null>(null);
    const [viewWeek, setViewWeek] = useState<number>(1);

    useEffect(() => {
        if (currentLeague) {
            setViewWeek(Math.max(1, currentLeague.CurrentWeekNumber));
        }
    }, [currentLeague]);

    useEffect(() => {
        if (!currentLeague?.ID) return;

        const fetchPlayers = async () => {
            try {
                setLoading(true);
                setError(null);
                const response = await getPlayersByLeague(currentLeague.ID);
                const sortedPlayers = response.data.sort((a: Player, b: Player) => a.InLeagueName.localeCompare(b.InLeagueName));
                setPlayers(sortedPlayers);

                if (sortedPlayers.length > 0) {
                    setSelectedPlayer(sortedPlayers[0]);
                }
            } catch (err) {
                if (axios.isAxiosError(err) && err.response) {
                    setError(err.response.data.error || "Failed to load players.");
                } else {
                    setError("A network or unknown error occurred while fetching players.");
                }
                console.error("Error fetching players:", err);
            } finally {
                setLoading(false);
            }
        };

        fetchPlayers();
    }, [currentLeague?.ID]);

    const selectedPlayerRoster = useMemo(() => {
        if (!selectedPlayer || !draftHistory) {
            return [];
        }
        return draftHistory.filter(pick => {
            const isPlayer = pick.PlayerID === selectedPlayer.ID;
            const isAcquired = pick.AcquiredWeek <= viewWeek;
            const isNotReleased = !pick.IsReleased || pick.ReleasedWeek > viewWeek;
            return isPlayer && isAcquired && isNotReleased;
        });
    }, [selectedPlayer, draftHistory, viewWeek]);

    if (!leagueLoading && !currentDraft) {
        return (
            <Layout variant="full">
                <div className="grow flex items-center justify-center">
                    <div className="bg-white p-8 rounded-lg shadow-md text-center">
                        <h1 className="text-2xl font-bold text-gray-800 mb-4">Draft Not Started</h1>
                        <p className="text-gray-600">The league has not yet begun and is being set up. This page is unavailable.</p>
                    </div>
                </div>
            </Layout>
        );
    }

    return (
        <Layout variant="container">
            <div className="flex flex-col md:flex-row bg-background-surface rounded-lg shadow-md p-4">
                {/* Left Pane: Player List */}
                <div className="w-full md:w-1/3 bg-background-surface p-6 rounded-lg shadow-md">
                    <h2 className="text-2xl font-bold text-text-primary mb-4">Players</h2>
                    {loading && <p className="text-text-secondary">Loading...</p>}
                    {error && <p className="text-red-500">Error: {error}</p>}
                    <ul className="space-y-2">
                        {players.map(player => (
                            <li key={player.ID}>
                                <button
                                    onClick={() => setSelectedPlayer(player)}
                                    className={`w-full text-left p-3 rounded-lg transition-all duration-150 border shadow-sm ${selectedPlayer?.ID === player.ID
                                        ? 'border-accent-primary bg-accent-primary text-text-on-accent font-semibold shadow-inner'
                                        : 'border-gray-200 bg-white text-text-primary hover:text-white hover:bg-accent-primary-hover'}`}>
                                    {player.InLeagueName}
                                </button>
                            </li>
                        ))}
                    </ul>
                </div>

                {/* Vertical Divider */}
                <div className="hidden md:flex items-stretch px-2">
                    <div className="w-px bg-gray-900/20 rounded-full" />
                </div>

                {/* Right Pane: Roster Display */}
                {/* TODO: Make week based roster buttons not render for tournament only leagues */}
                {/* These do not use transfers or CurrentWeekNumber, so they shouldn't need those week buttons */}
                <div className="w-full md:w-2/3 bg-background-surface p-6 rounded-lg shadow-md">
                    {historyLoading && <p className="text-text-secondary">Loading roster...</p>}
                    {historyError && <p className="text-red-500">Error: {historyError}</p>}
                    {!historyLoading && !historyError && selectedPlayer ? (
                        <div>
                            <h3 className="text-2xl font-bold text-text-primary mb-4">{selectedPlayer.TeamName || selectedPlayer.InLeagueName}'s Roster</h3>

                            {/* Week Selection Buttons */}
                            <div className="flex flex-wrap gap-2 mb-4">
                                {currentLeague && Array.from({ length: Math.max(1, currentLeague.CurrentWeekNumber) }, (_, i) => i + 1).map(week => (
                                    <button
                                        key={week}
                                        onClick={() => setViewWeek(week)}
                                        className={`px-3 py-1 rounded-md text-sm transition-all duration-150 border shadow-sm ${viewWeek === week
                                            ? 'border-accent-primary bg-accent-primary text-text-on-accent font-bold shadow-inner'
                                            : 'border-gray-200 bg-white text-text-primary hover:text-white hover:bg-accent-primary-hover'
                                        }`}
                                    >
                                        RS W{week}
                                    </button>
                                ))}
                            </div>

                            {selectedPlayerRoster.length > 0 ? (
                                <PokemonRosterList roster={selectedPlayerRoster} rosterType='drafted' bgColor="bg-gray-50" />
                            ) : (
                                <p className="text-text-secondary">This player has no Pokémon on their roster.</p>
                            )}
                        </div>
                    ) : (
                        <div className="flex items-center justify-center h-full">
                            <p className="text-text-secondary text-lg">Select a player to view their roster.</p>
                        </div>
                    )}
                </div>
            </div>
            <div className="flex flex-col md:flex-col gap-5 py-5">
                <div className="w-full bg-background-surface p-4 rounded-lg shadow-md">
                    <TeamPokemonView roster={selectedPlayerRoster} />
                </div>
                <div className="w-full bg-background-surface p-4 rounded-lg shadow-md">
                    {historyLoading && <p className="text-text-secondary">Loading defensive chart...</p>}
                    {historyError && <p className="text-red-500">Error: {historyError}</p>}
                    {!historyLoading && !historyError && selectedPlayerRoster.length > 0 && (
                        <DefensiveTypeChart roster={selectedPlayerRoster} />
                    )}
                    {!historyLoading && !historyError && selectedPlayerRoster.length === 0 && selectedPlayer && (
                        <p className="p-4 text-text-secondary">{selectedPlayer.InLeagueName} has no Pokémon on their roster to display defensive types.</p>
                    )}
                    {!selectedPlayer && (
                        <p className="p-4 text-text-secondary">Select a player to view their defensive type chart.</p>
                    )}
                </div>
                <div className="w-full bg-background-surface p-4 rounded-lg shadow-md">
                    <SpeedTable roster={selectedPlayerRoster} />
                </div>
            </div>
        </Layout>
    );
};
export default Teamsheets;
