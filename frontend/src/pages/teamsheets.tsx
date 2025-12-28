import { useState, useEffect, useCallback } from 'react'; // Dummy change
import Layout from '../components/Layout';
import { useLeague } from '../context/LeagueContext';
import { getPlayersByLeague, getDraftedPokemonByPlayer } from '../api/api';
import { Player, DraftedPokemon } from '../api/data_interfaces';
import axios from 'axios';
import { PokemonRosterList } from '../components/PokemonRosterList';
import { DefensiveTypeChart } from '../components/DefensiveTypeChart';
import TeamPokemonView from '../components/TeamPokemonView';
import SpeedTable from '../components/SpeedTable';

// Roster Display Component
const RosterDisplay = ({ player }: { player: Player }) => {
    const [roster, setRoster] = useState<DraftedPokemon[]>([]);
    const [rosterLoading, setRosterLoading] = useState(true);
    const [rosterError, setRosterError] = useState<string | null>(null);

    useEffect(() => {
        if (!player.ID) return;

        const fetchRoster = async () => {
            try {
                setRosterLoading(true);
                setRosterError(null);
                const response = await getDraftedPokemonByPlayer(player.LeagueID, player.ID);
                setRoster(response.data || []);
            } catch (err) {
                if (axios.isAxiosError(err) && err.response) {
                    setRosterError(err.response.data.error || "Failed to load roster. Server maybe offline.");
                } else {
                    setRosterError("A network or unknown error occurred while fetching the roster.");
                }
                console.error(`Error fetching roster for player ${player.ID}:`, err);
            } finally {
                setRosterLoading(false);
            }
        };

        fetchRoster();
    }, [player.ID, player.LeagueID]);

    if (rosterLoading) return <p className="text-text-secondary">Loading roster...</p>;
    if (rosterError) return <p className="text-red-500">Error: {rosterError}</p>;

    return (
        <div>
            <h3 className="text-2xl font-bold text-text-primary mb-4">{player.TeamName || player.InLeagueName}'s Roster</h3>
            {roster.length > 0 ? (
                <PokemonRosterList roster={roster} rosterType='drafted' bgColor="bg-gray-50" />
            ) : (
                <p className="text-text-secondary">This player has no Pokémon on their roster.</p>
            )}
        </div>
    );
};

// Main Teamsheets Page Component
const Teamsheets: React.FC = () => {
    const { currentLeague, currentDraft, loading: leagueLoading } = useLeague();
    const [players, setPlayers] = useState<Player[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [selectedPlayer, setSelectedPlayer] = useState<Player | null>(null);
    const [selectedPlayerRoster, setSelectedPlayerRoster] = useState<DraftedPokemon[]>([]);
    const [rosterLoading, setRosterLoading] = useState(false);
    const [rosterError, setRosterError] = useState<string | null>(null);

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

    const fetchSelectedPlayerRoster = useCallback(async () => {
        if (!selectedPlayer?.ID || !currentLeague?.ID) {
            setSelectedPlayerRoster([]);
            return;
        }

        try {
            setRosterLoading(true);
            setRosterError(null);
            const response = await getDraftedPokemonByPlayer(currentLeague.ID, selectedPlayer.ID);
            setSelectedPlayerRoster(response.data || []);
        } catch (err) {
            if (axios.isAxiosError(err) && err.response) {
                setRosterError(err.response.data.error || "Failed to load roster. Server maybe offline.");
            } else {
                setRosterError("A network or unknown error occurred while fetching the roster.");
            }
            console.error(`Error fetching roster for player ${selectedPlayer.ID}:`, err);
        } finally {
            setRosterLoading(false);
        }
    }, [selectedPlayer?.ID, currentLeague?.ID]);

    useEffect(() => {
        fetchSelectedPlayerRoster();
    }, [fetchSelectedPlayerRoster]);


    if (!leagueLoading && !currentDraft) {
        return (
            <Layout variant="full">
                <div className="flex-grow flex items-center justify-center">
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
            <div className="flex flex-col md:flex-row gap-6 py-6">
                {/* Left Pane: Player List */}
                <div className="w-full md:w-1/3 bg-background-surface p-4 rounded-lg shadow-md">
                    <h2 className="text-2xl font-bold text-text-primary mb-4">Players</h2>
                    {loading && <p className="text-text-secondary">Loading...</p>}
                    {error && <p className="text-red-500">Error: {error}</p>}
                    <ul className="space-y-2">
                        {players.map(player => (
                            <li key={player.ID}>
                                <button
                                    onClick={() => setSelectedPlayer(player)}
                                    className={`w-full text-left p-3 rounded-lg transition-all duration-150 border shadow-sm ${selectedPlayer?.ID === player.ID ? 'border-accent-primary bg-accent-primary text-text-on-accent font-semibold shadow-inner' : 'border-gray-200 bg-white text-text-primary hover:bg-gray-50 hover:border-gray-300'}`}>
                                    {player.InLeagueName}
                                </button>
                            </li>
                        ))}
                    </ul>
                </div>

                {/* Right Pane: Roster Display */}
                <div className="w-full md:w-2/3 bg-background-surface p-6 rounded-lg shadow-md">
                    {selectedPlayer ? (
                        <RosterDisplay player={selectedPlayer} />
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
                    {rosterLoading && <p className="text-text-secondary">Loading defensive chart...</p>}
                    {rosterError && <p className="text-red-500">Error: {rosterError}</p>}
                    {!rosterLoading && !rosterError && selectedPlayerRoster.length > 0 && (
                        <DefensiveTypeChart roster={selectedPlayerRoster} />
                    )}
                    {!rosterLoading && !rosterError && selectedPlayerRoster.length === 0 && selectedPlayer && (
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
