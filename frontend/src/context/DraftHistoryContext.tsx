import { createContext, useContext, useState, useEffect, ReactNode, useCallback } from 'react';
import { DraftedPokemon } from '../api/data_interfaces';
import { getDraftHistory } from '../api/api';
import { useLeague } from './LeagueContext';
import axios from 'axios';

interface DraftHistoryContextType {
    draftHistory: DraftedPokemon[];
    loading: boolean;
    error: string | null;
    refetch: () => void;
}

export const DraftHistoryContext = createContext<DraftHistoryContextType | undefined>(undefined);

interface DraftHistoryProviderProps {
    children: ReactNode;
}

export const DraftHistoryProvider = ({ children }: DraftHistoryProviderProps) => {
    const { currentLeague } = useLeague();
    const [draftHistory, setDraftHistory] = useState<DraftedPokemon[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);

    const fetchDraftHistory = useCallback(async () => {
        if (!currentLeague?.ID) {
            setDraftHistory([]);
            setLoading(false);
            return;
        }

        try {
            setLoading(true);
            setError(null);
            const response = await getDraftHistory(currentLeague.ID);
            setDraftHistory(response.data);
        } catch (err) {
            if (axios.isAxiosError(err) && err.response) {
                setError(err.response.data.error || "Failed to load draft history.");
            } else {
                setError("A network or unknown error occurred while fetching draft history.");
            }
            console.error("Error fetching draft history:", err);
        } finally {
            setLoading(false);
        }
    }, [currentLeague?.ID]);

    useEffect(() => {
        fetchDraftHistory();
    }, [fetchDraftHistory]);

    const contextValue: DraftHistoryContextType = {
        draftHistory,
        loading,
        error,
        refetch: fetchDraftHistory,
    };

    return (
        <DraftHistoryContext.Provider value={contextValue}>
            {children}
        </DraftHistoryContext.Provider>
    );
};

export const useDraftHistory = () => {
    const context = useContext(DraftHistoryContext);
    if (context === undefined) {
        throw new Error('useDraftHistory must be used within a DraftHistoryProvider');
    }
    return context;
};
