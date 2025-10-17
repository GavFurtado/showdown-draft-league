import { useState, useEffect } from 'react';
import Layout from '../components/Layout';
import { useLeague } from '../context/LeagueContext';
import { getPlayersByLeague, getDraftedPokemonByPlayer } from '../api/api';
import { Player, DraftedPokemon } from '../api/data_interfaces';
import { PokemonListItem } from '../components/PokemonListItem';
import axios from 'axios';

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
                    setRosterError(err.response.data.error || "Failed to load roster.");
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
                <div className="flex flex-col space-y-2">
                    {roster.map(dp => (
                        <PokemonListItem
                            key={dp.ID}
                            pokemon={dp.PokemonSpecies} // Pass the PokemonSpecies object
                            cost={dp.LeaguePokemon.Cost} // Pass the cost from LeaguePokemon
                            leaguePokemonId={dp.LeaguePokemonID} // Pass the LeaguePokemonID
                            pickNumber={dp.DraftPickNumber} // Pass the draft pick number
                            bgColor="bg-gray-50"
                        />
                    ))}
                </div>
            ) : (
                <p className="text-text-secondary">This player has no Pokémon on their roster.</p>
            )}
        </div>
    );
};

// Main Teamsheets Page Component
export default function Teamsheets() {
    const { currentLeague } = useLeague();
    const [players, setPlayers] = useState<Player[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [selectedPlayer, setSelectedPlayer] = useState<Player | null>(null);

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

    return (
        <Layout variant="container">
            <div className="flex flex-col md:flex-row gap-6 py-6">
                {/* Left Pane: Player List */}
                <div className="w-full md:w-1/3 bg-background-surface p-4 rounded-lg shadow-md">
                    <h2 className="text-xl font-bold text-text-primary mb-4">Players</h2>
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
        </Layout>
    );
}