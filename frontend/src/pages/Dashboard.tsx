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
import { Cog6ToothIcon } from '@heroicons/react/24/outline';

const Dashboard: React.FC = () => {
    const { currentLeague, currentPlayer, loading: leagueLoading, error: leagueError } = useLeague();

    const [players, setPlayers] = useState<Player[]>([]);
    const [roster, setRoster] = useState<DraftedPokemon[]>([]);
    const [games, setGames] = useState<Game[]>([]);

    const [loadingData, setLoadingData] = useState(true);
    const [showSettings, setShowSettings] = useState(false);

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
                            <p className="mt-1 ml-0.5 text-sm text-text-secondary">
                                Season Status: <span className="font-medium text-text-primary">{currentLeague.Status.replace('_', ' ')}</span>
                            </p>
                            <div className="mt-2">
                                <button
                                    type="button"
                                    onClick={() => setShowSettings(!showSettings)}
                                    className="inline-flex items-center rounded-md bg-white px-2 py-1 text-xs font-semibold text-text-primary shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                >
                                    <Cog6ToothIcon className="-ml-0.5 mr-1.5 h-4 w-4 text-gray-400" aria-hidden="true" />
                                    View League Details
                                </button>
                            </div>
                        </div>
                        <div className="mt-4 flex flex-col items-end md:ml-4 md:mt-0">
                            {currentPlayer && (
                                <div className="text-right">
                                    <div className="text-lg font-bold text-text-primary">{currentPlayer.TeamName}</div>
                                    <div className="text-sm text-text-secondary">{currentPlayer.InLeagueName}</div>
                                </div>
                            )}
                        </div>
                    </div>

                    <div className={`grid transition-[grid-template-rows] duration-100 ease-out ${showSettings ? 'grid-rows-[1fr]' : 'grid-rows-[0fr]'}`}>
                        <div className="overflow-hidden">
                            <div className="mt-4 border-t border-gray-200 pt-4">
                                <h3 className="text-sm font-medium text-text-primary mb-3">League Format & Rules</h3>
                                <dl className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2 lg:grid-cols-3">
                                    <div className="sm:col-span-1">
                                        <dt className="text-sm font-medium text-text-secondary">Season Type</dt>
                                        <dd className="mt-1 text-sm text-text-primary">{currentLeague.Format.SeasonType.replace(/_/g, ' ')}</dd>
                                    </div>
                                    <div className="sm:col-span-1">
                                        <dt className="text-sm font-medium text-text-secondary">Draft Type</dt>
                                        <dd className="mt-1 text-sm text-text-primary">
                                            {currentLeague.Format.IsSnakeRoundDraft ? "Snake Draft" : "Linear Draft"} ({currentLeague.Format.DraftOrderType})
                                        </dd>
                                    </div>
                                    <div className="sm:col-span-1">
                                        <dt className="text-sm font-medium text-text-secondary">Roster Size Limits</dt>
                                        <dd className="mt-1 text-sm text-text-primary">Min: {currentLeague.MinPokemonPerPlayer}, Max: {currentLeague.MaxPokemonPerPlayer}</dd>
                                    </div>
                                    <div className="sm:col-span-1">
                                        <dt className="text-sm font-medium text-text-secondary">Starting Points</dt>
                                        <dd className="mt-1 text-sm text-text-primary">{currentLeague.StartingDraftPoints}</dd>
                                    </div>
                                    <div className="sm:col-span-1">
                                        <dt className="text-sm font-medium text-text-secondary">Number of Groups</dt>
                                        <dd className="mt-1 text-sm text-text-primary">{currentLeague.Format.GroupCount}</dd>
                                    </div>
                                    <div className="sm:col-span-1">
                                        <dt className="text-sm font-medium text-text-secondary">Playoff Format</dt>
                                        <dd className="mt-1 text-sm text-text-primary">
                                            {currentLeague.Format.PlayoffType !== "NONE"
                                                ? `${currentLeague.Format.PlayoffType.replace('_', ' ')} (${currentLeague.Format.PlayoffParticipantCount} teams)`
                                                : "No Playoffs"}
                                        </dd>
                                    </div>
                                    <div className="sm:col-span-1">
                                        <dt className="text-sm font-medium text-text-secondary">Transfers</dt>
                                        <dd className="mt-1 text-sm text-text-primary">
                                            {currentLeague.Format.AllowTransfers ? "Allowed" : "Disabled"}
                                        </dd>
                                    </div>
                                    {currentLeague.Format.AllowTransfers && (
                                        <>
                                            <div className="sm:col-span-1">
                                                <dt className="text-sm font-medium text-text-secondary">Transfer Window Frequency</dt>
                                                <dd className="mt-1 text-sm text-text-primary">
                                                    Every {currentLeague.Format.TransferWindowFrequencyDays} days
                                                </dd>
                                            </div>
                                            <div className="sm:col-span-1">
                                                <dt className="text-sm font-medium text-text-secondary">Transfer Window Duration</dt>
                                                <dd className="mt-1 text-sm text-text-primary">
                                                    {currentLeague.Format.TransferWindowDuration} hours
                                                </dd>
                                            </div>
                                            <div className="sm:col-span-1">
                                                <dt className="text-sm font-medium text-text-secondary">Transfer Credits</dt>
                                                <dd className="mt-1 text-sm text-text-primary">
                                                    {currentLeague.Format.TransfersCostCredits
                                                        ? `${currentLeague.Format.TransferCreditsPerWindow} per window (Cap: ${currentLeague.Format.TransferCreditCap})`
                                                        : "Unlimited"}
                                                </dd>
                                            </div>
                                            {currentLeague.Format.TransfersCostCredits && (
                                                <div className="sm:col-span-1">
                                                    <dt className="text-sm font-medium text-text-secondary">Transfer Costs</dt>
                                                    <dd className="mt-1 text-sm text-text-primary">
                                                        Drop: {currentLeague.Format.DropCost}, Pickup: {currentLeague.Format.PickupCost}
                                                    </dd>
                                                </div>
                                            )}
                                        </>
                                    )}
                                </dl>
                                <div className="mt-4">
                                    <dt className="text-sm font-medium text-text-secondary">Rules Description</dt>
                                    <dd className="mt-1 text-sm text-text-primary whitespace-pre-wrap bg-gray-50 p-3 rounded-md border border-gray-100">
                                        {currentLeague.RulesetDescription || "No specific rules description provided."}
                                    </dd>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">

                    {/* My Roster Widget */}
                    <div className="bg-white overflow-hidden shadow rounded-lg border border-gray-200 flex flex-col h-full">
                        <div className="px-4 py-5 sm:px-6 flex flex-col border-b border-gray-200">
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
