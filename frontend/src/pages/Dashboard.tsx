import React, { useEffect, useState } from 'react';
import Layout from '../components/Layout';
import { useLeague } from '../context/LeagueContext';
import { PokemonRosterList } from '../components/PokemonRosterList';
import { DashboardStandings } from '../components/DashboardStandings';
import { DashboardSchedule } from '../components/DashboardSchedule';
import { Player, DraftedPokemon, Game } from '../api/data_interfaces';
import { getPlayersByLeague, getDraftedPokemonByPlayer, getGamesByPlayer } from '../api/api';
import { Link } from 'react-router-dom';
import { getPlayerSlug } from '../utils/nameFormatter';

const Dashboard: React.FC = () => {
    const { currentLeague, currentPlayer, loading: leagueLoading, error: leagueError } = useLeague();

    const [players, setPlayers] = useState<Player[]>([]);
    const [roster, setRoster] = useState<DraftedPokemon[]>([]);
    const [games, setGames] = useState<Game[]>([]);

    const [loadingData, setLoadingData] = useState(true);

    useEffect(() => {
        const fetchData = async () => {
            if (!currentLeague?.ID) return;

            setLoadingData(true);
            try {
                // Fetch Players (Standings)
                const playersRes = await getPlayersByLeague(currentLeague.ID);
                setPlayers(playersRes.data);

                if (currentPlayer?.ID) {
                    // Fetch My Roster
                    const rosterRes = await getDraftedPokemonByPlayer(currentLeague.ID, currentPlayer.ID);
                    setRoster(rosterRes.data);

                    // Fetch My Schedule
                    const gamesRes = await getGamesByPlayer(currentLeague.ID, currentPlayer.ID);
                    setGames(gamesRes.data.games || []);
                }
            } catch (err) {
                console.error("Failed to fetch dashboard data", err);
            } finally {
                setLoadingData(false);
            }
        };

        if (currentLeague) {
            fetchData();
        }
    }, [currentLeague, currentPlayer]);


    if (leagueLoading) {
        return (
            <Layout variant="container">
                <div className="flex justify-center items-center h-64">
                    <span className="text-xl text-text-secondary">Loading League Dashboard...</span>
                </div>
            </Layout>
        );
    }

    if (leagueError) {
        return (
            <Layout variant="container">
                <div className="rounded-md bg-red-50 p-4 mt-8">
                    <h3 className="text-sm font-medium text-red-800">Error loading league</h3>
                    <div className="mt-2 text-sm text-red-700">{leagueError}</div>
                </div>
            </Layout>
        );
    }

    if (!currentLeague) {
        return (
            <Layout variant="container">
                <div className="text-center py-12">
                    <h3 className="text-lg font-medium text-text-primary">League not found</h3>
                </div>
            </Layout>
        );
    }

    return (
        <Layout variant="container">
            <div className="space-y-6">
                <div className="bg-white shadow px-4 py-5 sm:rounded-lg sm:p-6 border border-gray-200">
                    <div className="md:flex md:items-center md:justify-between">
                        <div className="min-w-0 flex-1">
                            <h2 className="text-2xl font-bold leading-7 text-text-primary sm:truncate sm:text-3xl sm:tracking-tight">
                                {currentLeague.Name}
                            </h2>
                            <p className="mt-1 text-sm text-text-secondary">
                                Season Status: <span className="font-medium text-text-primary">{currentLeague.Status.replace('_', ' ')}</span>
                            </p>
                        </div>
                    </div>
                </div>

                <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">

                    {/* My Roster Widget */}
                    <div className="bg-white overflow-hidden shadow rounded-lg border border-gray-200 flex flex-col h-full">
                        <div className="px-4 py-5 sm:px-6 border-b border-gray-200">
                            <h3 className="text-lg font-medium text-text-primary">My Roster</h3>
                        </div>
                        <div className="p-4 grow overflow-y-auto">
                            {loadingData ? (
                                <p className="text-text-secondary">Loading roster...</p>
                            ) : currentPlayer ? (
                                <PokemonRosterList roster={roster} rosterType="drafted" />
                            ) : (
                                <p className="text-text-secondary">You are not a player in this league.</p>
                            )}
                        </div>
                        <div className="bg-gray-50 px-4 py-3 text-right">
                            <Link to={`/league/${currentLeague.ID}/teamsheets#${currentPlayer ? getPlayerSlug(currentPlayer.InLeagueName) : ''}`} className="text-xs font-medium text-accent-primary hover:text-accent-primary-hover">
                                View Teamsheets &rarr;
                            </Link>
                        </div>
                    </div>

                    {/* Standings Widget */}
                    <div className="bg-white overflow-hidden shadow rounded-lg border border-gray-200 flex flex-col h-full">
                        <div className="px-4 py-5 sm:px-6 border-b border-gray-200">
                            <h3 className="text-lg font-medium text-text-primary">Standings</h3>
                        </div>
                        <div className="grow overflow-hidden">
                            {loadingData ? (
                                <div className="p-4 text-text-secondary">Loading standings...</div>
                            ) : (
                                <DashboardStandings players={players} />
                            )}
                        </div>
                    </div>

                    {/* Schedule Widget */}
                    <div className="bg-white overflow-hidden shadow rounded-lg border border-gray-200 flex flex-col h-full">
                        <div className="px-4 py-5 sm:px-6 border-b border-gray-200">
                            <h3 className="text-lg font-medium text-text-primary">My Schedule</h3>
                        </div>
                        <div className="grow overflow-hidden">
                            {loadingData ? (
                                <div className="p-4 text-text-secondary">Loading schedule...</div>
                            ) : currentPlayer ? (
                                <DashboardSchedule games={games} currentPlayerId={currentPlayer.ID} />
                            ) : (
                                <p className="p-4 text-text-secondary">You are not a player in this league.</p>
                            )}
                        </div>
                    </div>
                </div>
            </div>
        </Layout>
    );
};

export default Dashboard;
