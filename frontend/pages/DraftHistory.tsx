import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { getDraftHistory } from '../src/api/api';
import { DraftedPokemon } from '../src/api/data_interfaces';
import Layout from '../src/components/Layout';
import { format } from 'date-fns';

const DraftHistory: React.FC = () => {
    const { leagueId } = useParams<{ leagueId: string }>();
    const [draftHistory, setDraftHistory] = useState<DraftedPokemon[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (leagueId) {
            const fetchDraftHistory = async () => {
                try {
                    setLoading(true);
                    const response = await getDraftHistory(leagueId);
                    setDraftHistory(response.data);
                } catch (err) {
                    setError('Failed to fetch draft history.');
                    console.error(err);
                } finally {
                    setLoading(false);
                }
            };
            fetchDraftHistory();
        }
    }, [leagueId]);

    if (loading) {
        return <Layout variant="container"><div className="text-white">Loading draft history...</div></Layout>;
    }

    if (error) {
        return <Layout variant="container"><div className="text-red-500">{error}</div></Layout>;
    }

    return (
        <Layout variant="container">
            {draftHistory.length === 0 ? (
                <p className="text-white">No draft history available for this league.</p>
            ) : (
                <div className="overflow-x-auto bg-background-surface-alt rounded-lg shadow">
                    <table className="min-w-full divide-y divide-gray-700">
                        <thead className="bg-gray-700">
                            <tr>
                                <th scope="col" className="text-left px-6 py-3 text-center text-xs font-medium text-text-on-nav uppercase tracking-wider">
                                    Pokemon
                                </th>
                                <th scope="col" className="text-left px-6 py-3 text-center text-xs font-medium text-text-on-nav uppercase tracking-wider">
                                    Player
                                </th>
                                <th scope="col" className=" w-28 px-3 py-3 text-center text-xs font-medium text-text-on-nav uppercase tracking-wider">
                                    Pick#
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
                                <th scope="col" className="text-left px-6 py-3 text-center text-xs font-medium text-text-on-nav uppercase tracking-wider">
                                    Timestamp
                                </th>
                            </tr>
                        </thead>
                        <tbody className="bg-background-surface divide-y divide-gray-700">
                            {draftHistory.map((item) => (
                                <tr key={item.ID}>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <div className="flex items-center">
                                            <div className="flex-shrink-0 h-10 w-10">
                                                <img className="h-10 w-10 rounded-full" src={item.PokemonSpecies.Sprites.FrontDefault || item.PokemonSpecies.Sprites.OfficialArtwork || '/placeholder-roster.png'} alt={item.PokemonSpecies.Name} />
                                            </div>
                                            <div className="ml-4">
                                                <div className="text-sm font-medium text-text-primary">{item.PokemonSpecies.Name}</div>
                                            </div>
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <div className="text-sm text-text-primary">{item.Player.InLeagueName}</div>
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
        </Layout>
    );
};

export default DraftHistory;
