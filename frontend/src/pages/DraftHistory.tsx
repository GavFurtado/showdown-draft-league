import React, { useState, useEffect, useCallback } from 'react';
import { useLeague } from '../context/LeagueContext';
import { getDraftHistory } from '../api/api';
import { formatPokemonName } from "../utils/nameFormatter";
import { DraftedPokemon } from '../api/data_interfaces';
import Layout from '../components/Layout';
import { PokemonListItem } from '../components/PokemonListItem';
import axios from 'axios';

const formatTimestamp = (isoString: string) => {
    const date = new Date(isoString);
    const localTime = date.toLocaleString(); // Client's local time
    const utcTime = date.toUTCString(); // UTC time
    return (
        <span title={`UTC: ${utcTime}`}>
            {localTime}
        </span>
    );
};

const getActionStatus = (draftedPokemon: DraftedPokemon): string => {
    if (draftedPokemon.IsReleased) {
        return "Released";
    } else if (draftedPokemon.DraftPickNumber !== null) {
        return "Draft Pick";
    } else {
        return "Free Agent Pickup";
    }
};

export default function DraftHistory() {
    const { currentLeague, loading: leagueLoading, error: leagueError } = useLeague();
    const [draftHistory, setDraftHistory] = useState<DraftedPokemon[]>([]);
    const [historyLoading, setHistoryLoading] = useState<boolean>(true);
    const [historyError, setHistoryError] = useState<string | null>(null);

    const fetchDraftHistory = useCallback(async () => {
        if (!currentLeague?.ID) {
            setDraftHistory([]);
            setHistoryLoading(false);
            return;
        }
        setHistoryLoading(true);
        setHistoryError(null);
        try {
            const response = await getDraftHistory(currentLeague.ID);
            setDraftHistory(response.data);
        } catch (err) {
            console.error("Failed to fetch draft history:", err);
            if (axios.isAxiosError(err) && err.response) {
                setHistoryError(err.response.data.error || "Failed to load draft history.");
            } else {
                setHistoryError("A network or unknown error occurred while fetching draft history.");
            }
        } finally {
            setHistoryLoading(false);
        }
    }, [currentLeague?.ID]);

    useEffect(() => {
        fetchDraftHistory();
    }, [fetchDraftHistory]);

    if (leagueLoading || historyLoading) {
        return <Layout variant="full"><div className="p-4 text-center text-text-body">Loading draft history...</div></Layout>;
    }

    if (leagueError || historyError) {
        return (
            <Layout variant="full">
                <div className="p-4 text-center text-danger-primary">
                    Error: {leagueError || historyError}
                </div>
            </Layout>
        );
    }

    return (
        <Layout variant="full">
            <div className="p-4 sm:p-6">
                <h1 className="text-3xl font-bold text-text-heading mb-6">Draft & Transaction History</h1>

                <div className="bg-card-bg shadow-md rounded-lg overflow-x-auto">
                    {draftHistory.length > 0 ? (
                        <table className="min-w-full divide-y divide-border-light">
                            <thead className="bg-secondary-bg">
                                <tr>
                                    <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-text-heading">Draft Position</th>
                                    <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-text-heading">Draft Round</th>
                                    <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-text-heading">Pick in Round</th>
                                    <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-text-heading">Pokémon</th>
                                    <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-text-heading">Player</th>
                                    <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-text-heading">Cost</th>
                                    <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-text-heading">Timestamp</th>
                                    <th scope="col" className="px-3 py-3.5 text-left text-sm font-semibold text-text-heading">Status</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-border-light bg-card-bg">
                                {draftHistory.map(dp => {
                                    const pickInRound = dp.DraftPickNumber !== null && currentLeague?.Players?.length
                                        ? ((dp.DraftPickNumber - 1) % currentLeague.Players.length) + 1
                                        : 'N/A';
                                    return (
                                        <tr key={dp.ID}>
                                            <td className="whitespace-nowrap px-3 py-4 text-sm text-text-body">{dp.DraftPickNumber ?? 'N/A'}</td>
                                            <td className="whitespace-nowrap px-3 py-4 text-sm text-text-body">{dp.DraftRoundNumber ?? 'N/A'}</td>
                                            <td className="whitespace-nowrap px-3 py-4 text-sm text-text-body">{pickInRound}</td>
                                            <td className="whitespace-nowrap px-3 py-4 text-sm text-text-body">
                                                <div className="flex items-center">
                                                    <img src={dp.PokemonSpecies.Sprites.FrontDefault} alt={dp.PokemonSpecies.Name} className="h-10 w-10 flex-shrink-0" />
                                                    <div className="ml-4">
                                                        <div className="font-medium text-text-default">{formatPokemonName(dp.PokemonSpecies.Name)}</div>
                                                    </div>
                                                </div>
                                            </td>
                                            <td className="whitespace-nowrap px-3 py-4 text-sm text-text-body">{dp.Player?.TeamName || dp.Player?.InLeagueName || 'N/A'}</td>
                                            <td className="whitespace-nowrap px-3 py-4 text-sm text-text-body">{dp.LeaguePokemon?.Cost ?? 'N/A'}</td>
                                            <td className="whitespace-nowrap px-3 py-4 text-sm text-text-body">{formatTimestamp(dp.CreatedAt)}</td>
                                            <td className="whitespace-nowrap px-3 py-4 text-sm text-text-body">{getActionStatus(dp)}</td>
                                        </tr>
                                    );
                                })}
                            </tbody>
                        </table>
                    ) : (
                        <p className="p-4 text-text-secondary">No draft or transaction history available for this league.</p>
                    )}
                </div>
            </div>
        </Layout>
    );
}