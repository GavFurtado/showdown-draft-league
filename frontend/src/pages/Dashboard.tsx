import React from 'react';
import Layout from '../components/Layout';
import { useLeague } from '../context/LeagueContext';

const Dashboard: React.FC = () => {
    const { currentLeague, loading, error } = useLeague();

    if (loading) {
        return (
            <Layout variant="container">
                <div className="flex justify-center items-center h-64">
                    <span className="text-xl text-text-secondary">Loading League Dashboard...</span>
                </div>
            </Layout>
        );
    }

    if (error) {
        return (
            <Layout variant="container">
                <div className="rounded-md bg-red-50 p-4 mt-8">
                    <h3 className="text-sm font-medium text-red-800">Error loading league</h3>
                    <div className="mt-2 text-sm text-red-700">{error}</div>
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
                    {/* Placeholder content for dashboard widgets */}
                    <div className="bg-white overflow-hidden shadow rounded-lg border border-gray-200">
                        <div className="px-4 py-5 sm:p-6">
                            <h3 className="text-lg font-medium text-text-primary">My Team</h3>
                            <p className="mt-1 text-sm text-text-secondary">View and manage your roster.</p>
                            <div className="mt-4">
                                <span className="text-accent-primary hover:text-accent-primary-hover text-sm font-medium cursor-pointer">
                                    View Roster &rarr;
                                </span>
                            </div>
                        </div>
                    </div>

                    <div className="bg-white overflow-hidden shadow rounded-lg border border-gray-200">
                        <div className="px-4 py-5 sm:p-6">
                            <h3 className="text-lg font-medium text-text-primary">Standings</h3>
                            <p className="mt-1 text-sm text-text-secondary">Check the current league standings.</p>
                            <div className="mt-4">
                                <span className="text-accent-primary hover:text-accent-primary-hover text-sm font-medium cursor-pointer">
                                    View Standings &rarr;
                                </span>
                            </div>
                        </div>
                    </div>

                    <div className="bg-white overflow-hidden shadow rounded-lg border border-gray-200">
                        <div className="px-4 py-5 sm:p-6">
                            <h3 className="text-lg font-medium text-text-primary">Schedule</h3>
                            <p className="mt-1 text-sm text-text-secondary">Upcoming matches and results.</p>
                            <div className="mt-4">
                                <span className="text-accent-primary hover:text-accent-primary-hover text-sm font-medium cursor-pointer">
                                    View Schedule &rarr;
                                </span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </Layout>
    );
};

export default Dashboard;