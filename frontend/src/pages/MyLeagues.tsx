import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { format } from "date-fns";
import axios from "axios";
import { PlusIcon, UserGroupIcon, CalendarIcon, TrophyIcon } from "@heroicons/react/24/outline";

import Layout from "../components/Layout";
import { League } from "../api/data_interfaces";
import { useUser } from "../context/UserContext";
import { getMyLeagues } from "../api/api";

const LeagueCard: React.FC<{ league: League }> = ({ league }) => {
    const statusColors = {
        PENDING: "bg-yellow-100 text-yellow-800",
        SETUP: "bg-blue-100 text-blue-800",
        DRAFTING: "bg-purple-100 text-purple-800",
        POST_DRAFT: "bg-indigo-100 text-indigo-800",
        REGULAR_SEASON: "bg-green-100 text-green-800",
        PLAYOFFS: "bg-orange-100 text-orange-800",
        COMPLETED: "bg-gray-100 text-gray-800",
        CANCELLED: "bg-red-100 text-red-800",
        TRANSFER_WINDOW: "bg-teal-100 text-teal-800",
    };

    const statusColor = statusColors[league.Status] || "bg-gray-100 text-gray-800";

    return (
        <Link
            to={`/league/${league.ID}/dashboard`}
            className="block bg-background-surface overflow-hidden rounded-lg shadow hover:shadow-md transition-shadow duration-200 border border-gray-200"
        >
            <div className="px-4 py-5 sm:p-6">
                <div className="flex justify-between items-start mb-4">
                    <h3 className="text-lg font-semibold text-text-primary truncate" title={league.Name}>
                        {league.Name}
                    </h3>
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColor}`}>
                        {league.Status.replace('_', ' ')}
                    </span>
                </div>

                <div className="space-y-2 text-sm text-text-secondary">
                    <div className="flex items-center">
                        <UserGroupIcon className="mr-2 h-4 w-4" />
                        <span>{league.PlayerCount} Players</span>
                    </div>
                    <div className="flex items-center">
                        <CalendarIcon className="mr-2 h-4 w-4" />
                        <span>Started: {format(new Date(league.StartDate), "MMM d, yyyy")}</span>
                    </div>
                    <div className="flex items-center">
                        <TrophyIcon className="mr-2 h-4 w-4" />
                        <span>{league.Format.SeasonType.replace('_', ' ')}</span>
                    </div>
                </div>
            </div>
            <div className="bg-gray-50 px-4 py-4 sm:px-6">
                <div className="text-sm text-accent-primary hover:text-accent-primary-hover font-medium">
                    Go to Dashboard &rarr;
                </div>
            </div>
        </Link>
    );
};

const MyLeagues: React.FC = () => {
    const [userLeagues, setUserLeagues] = useState<League[] | null>(null);
    const [userLeaguesLoading, setUserLeaguesLoading] = useState<boolean>(true);
    const [userLeaguesError, setUserLeaguesError] = useState<string | null>(null);
    const { user } = useUser();

    useEffect(() => {
        const fetchUserLeagues = async () => {
            if (!user?.ID) {
                // If user is not logged in or context not ready, handled by loading state or auth redirect usually
                // but if we are here and no user, strictly empty
                setUserLeagues([]);
                setUserLeaguesLoading(false);
                return;
            }

            try {
                setUserLeaguesLoading(true);
                setUserLeaguesError(null);
                const response = await getMyLeagues();
                setUserLeagues(response.data);
            } catch (err) {
                if (axios.isAxiosError(err) && err.response) {
                    setUserLeaguesError(err.response.data.error || "Failed to load user's leagues.");
                } else {
                    setUserLeaguesError("A network or unknown error occurred while fetching user's leagues.");
                }
            } finally {
                setUserLeaguesLoading(false);
            }
        };

        if (user) {
            fetchUserLeagues();
        } else {
            // Wait for user to be loaded or determined null
            // userLoading from context would be better to check, but existing code relied on user object check
            // We can assume if user is nullish initially, we wait. 
            // However, to keep it simple and safe:
            if (user === null) {
                // Context might still be loading, but if explicitly null (logged out), we stop
                setUserLeaguesLoading(false);
            }
        }
    }, [user]);

    const handleCreateLeague = () => {
        alert("Create League feature is coming soon!");
    };

    if (userLeaguesError) {
        return (
            <Layout variant="container">
                <div className="rounded-md bg-red-50 p-4 mt-8">
                    <div className="flex">
                        <div className="ml-3">
                            <h3 className="text-sm font-medium text-red-800">Error</h3>
                            <div className="mt-2 text-sm text-red-700">
                                <p>{userLeaguesError}</p>
                            </div>
                        </div>
                    </div>
                </div>
            </Layout>
        );
    }

    return (
        <Layout variant="container">
            <div className="sm:flex sm:items-center sm:justify-between mb-8">
                <div>
                    <h1 className="text-3xl font-bold text-text-primary">My Leagues</h1>
                    <p className="mt-2 text-sm text-text-secondary">
                        Manage and view all the leagues you are participating in.
                    </p>
                </div>
                <div className="mt-4 sm:mt-0">
                    <button
                        type="button"
                        onClick={handleCreateLeague}
                        className="inline-flex items-center rounded-md bg-accent-primary px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-accent-primary-hover focus-visible:outline-2 focus-visible:outline-accent-primary"
                    >
                        <PlusIcon className="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
                        Create League
                    </button>
                </div>
            </div>

            {userLeaguesLoading ? (
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    {[1, 2, 3].map((i) => (
                        <div key={i} className="animate-pulse bg-white rounded-lg shadow p-6 h-48 border border-gray-200">
                            <div className="h-4 bg-gray-200 rounded w-3/4 mb-4"></div>
                            <div className="space-y-3">
                                <div className="h-3 bg-gray-200 rounded"></div>
                                <div className="h-3 bg-gray-200 rounded w-5/6"></div>
                            </div>
                        </div>
                    ))}
                </div>
            ) : userLeagues && userLeagues.length > 0 ? (
                <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
                    {userLeagues.map((league) => (
                        <LeagueCard key={league.ID} league={league} />
                    ))}
                </div>
            ) : (
                <div className="text-center py-12 bg-white rounded-lg shadow border border-gray-200">
                    <TrophyIcon className="mx-auto h-12 w-12 text-gray-400" />
                    <h3 className="mt-2 text-sm font-semibold text-text-primary">No leagues found</h3>
                    <p className="mt-1 text-sm text-text-secondary">Get started by creating a new league or joining an existing one.</p>
                    <div className="mt-6">
                        <button
                            type="button"
                            onClick={handleCreateLeague}
                            className="inline-flex items-center rounded-md bg-accent-primary px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-accent-primary-hover focus-visible:outline-2 focus-visible:outline-accent-primary"
                        >
                            <PlusIcon className="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
                            Create League
                        </button>
                    </div>
                </div>
            )}
        </Layout>
    );
};

export default MyLeagues;
