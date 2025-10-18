import { useState, useEffect } from 'react';
import Layout from '../components/Layout';
import { useLeague } from '../context/LeagueContext';
import { getPlayersByLeague, getDraftedPokemonByPlayer } from '../api/api';
import { Player, DraftedPokemon } from '../api/data_interfaces';
import axios from 'axios';
import { PokemonRosterList } from '../components/PokemonRosterList';

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
            <div className="flex flex-col md:flex-col gap-5 py-5">
                <div className="w-full bg-background-surface p-4 rounded-lg shadow-md">
                    <h2 className="text-xl font-bold text-text-primary mb-4">Roster's Defensive Type Chart</h2>
                    <table className="min-w-full divide-y divide-gray-700 table-fixed">
                        <thead className="bg-gray-700">
                            <tr>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-left text-[10px] uppercase tracking-wider w-40">
                                    Pokémon
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Normal
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Fire
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Water
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Electric
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Grass
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Ice
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Fighting
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Poison
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Ground
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Flying
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Psychic
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Bug
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Rock
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Ghost
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Dragon
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Dark
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Steel
                                </th>
                                <th scope="col" className="px-2 py-3 text-text-on-nav text-center text-[10px] uppercase tracking-wider w-[calc((100%-160px)/18)]">
                                    Fairy
                                </th>

                            </tr>
                        </thead>
                    </table>
                </div>
            </div>

        </Layout>
    );
}
