import React from 'react';
import { DraftedPokemon } from '../api/data_interfaces';
import Layout from '../components/Layout';
import { format } from 'date-fns';
import { formatPokemonName } from '../utils/nameFormatter';
import { useLeague } from '../context/LeagueContext';
import { useDraftHistory } from '../context/DraftHistoryContext';

const DraftHistory: React.FC = () => {
    const { currentDraft, loading: leagueLoading } = useLeague();
    const { draftHistory, loading: historyLoading, error: historyError, refetch } = useDraftHistory();

    if (leagueLoading) {
        return <Layout variant="container"><div className="text-center text-xl text-text-primary">Loading...</div></Layout>;
    }

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
            <div className='flex-col bg-background-surface rounded-lg shadow-md px-4 py-4'>
                <div className="flex justify-between items-center mb-4">
                    <h1 className="text-3xl font-bold text-text-primary">Draft History</h1>
                    <button
                        onClick={refetch}
                        disabled={historyLoading}
                        className="px-3 py-1.5 border text-text-primary rounded-md hover:bg-accent-primary hover:text-text-on-accent disabled:bg-gray-500"
                    >
                        {historyLoading ? 'Refreshing...' : 'Refresh'}
                    </button>
                </div>

                {historyLoading && <div className="text-center text-xl text-text-primary">Loading draft history...</div>}
                {historyError && <div className="text-red-500">{historyError}</div>}

                {!historyLoading && !historyError && (
                    <>
                        {draftHistory.length === 0 ? (
                            <p className="text-white">No draft history available for this league.</p>
                        ) : (
                            <div className="overflow-x-auto bg-background-surface-alt border-black/50 rounded-lg shadow-md">
                                <table className="min-w-full divide-y divide-gray-700">
                                    <thead className="bg-gray-700 border rouded-lg shadow-md">
                                        <tr>
                                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-text-on-nav uppercase tracking-wider">
                                                Pokémon
                                            </th>
                                            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-text-on-nav uppercase tracking-wider">
                                                Player
                                            </th>
                                            <th scope="col" className="w-28 px-3 py-3 text-center text-xs font-medium text-text-on-nav uppercase tracking-wider">
                                                Pick #
                                            </th>
                                            <th scope="col" className="w-28 px-3 py-3 text-center text-xs font-medium text-text-on-nav uppercase tracking-wider">
                                                Round
                                            </th>
                                            <th scope="col" className="w-28 px-3 py-3 text-center text-xs font-medium text-text-on-nav uppercase tracking-wider">
                                                Cost
                                            </th>
                                            <th scope="col" className="px-6 py-3 text-center text-xs font-medium text-text-on-nav uppercase tracking-wider">
                                                Status
                                            </th>
                                            <th scope="col" className="px-6 py-3 text-center text-xs font-medium text-text-on-nav uppercase tracking-wider">
                                                Action
                                            </th>
                                            <th scope="col" className="px-6 py-3 text-center text-xs font-medium text-text-on-nav uppercase tracking-wider">
                                                Timestamp
                                            </th>
                                        </tr>
                                    </thead>
                                    <tbody className="bg-background-surface divide-y divide-gray-700/20">
                                        {draftHistory.map((item) => (
                                            <tr key={item.ID}>
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <div className="flex items-center">
                                                        <div className="flex-shrink-0 h-10 w-10">
                                                            <img className="h-10 w-10 rounded-full" src={item.PokemonSpecies.Sprites.FrontDefault || item.PokemonSpecies.Sprites.OfficialArtwork || '/placeholder-roster.png'} alt={item.PokemonSpecies.Name} />
                                                        </div>
                                                        <div className="ml-4">
                                                            <div className="text-sm font-medium text-text-primary">{formatPokemonName(item.PokemonSpecies.Name)}</div>
                                                        </div>
                                                    </div>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap">
                                                    <div className="text-left text-sm text-text-primary">{item.Player.InLeagueName}</div>
                                                </td>
                                                <td className="w-28 px-3 py-4 whitespace-nowrap text-sm text-text-primary text-center">
                                                    {item.DraftPickNumber}
                                                </td>
                                                <td className="w-28 px-3 py-4 whitespace-nowrap text-sm text-text-primary text-center">
                                                    {item.DraftRoundNumber}
                                                </td>
                                                <td className="w-28 px-3 py-4 whitespace-nowrap text-sm text-text-primary text-center">
                                                    {item.LeaguePokemon.Cost ?? 'N/A'}
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap text-center">
                                                    <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${item.IsReleased ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800'}`}>
                                                        {item.IsReleased ? 'Released' : 'Active'}
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap text-center">
                                                    <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${item.IsReleased ? 'bg-orange-100 text-orange-800' : 'bg-blue-100 text-blue-800'}`}>
                                                        {item.IsReleased ? 'Dropped' : 'Draft Pick'}
                                                    </span>
                                                </td>
                                                <td className="px-6 py-4 whitespace-nowrap text-sm text-text-secondary text-center">
                                                    {format(new Date(item.CreatedAt), 'MMM dd, yyyy HH:mm')}
                                                </td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                        )}
                    </>
                )}
            </div>
        </Layout>
    );
};

export default DraftHistory;
