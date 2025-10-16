import { createContext, useContext, useState, useEffect, ReactNode, useCallback, useMemo } from 'react';
import { League, Draft, Player, DiscordUser } from '../api/data_interfaces';
import { getLeague, getDraftByLeagueID, getMyDiscordDetails, getPlayersByUserId, getPlayerById } from '../api/api';
import axios from 'axios'; // Import axios for error handling
import { useParams } from 'react-router-dom'; // Import useParams

interface LeagueContextType {
    currentLeague: League | null;
    setCurrentLeague: (league: League | null) => void;
    currentDraft: Draft | null;
    currentPlayer: Player | null;
    myDiscordUser: DiscordUser | null;
    loading: boolean;
    error: string | null;
    refetch: () => void;
}

export const LeagueContext = createContext<LeagueContextType | undefined>(undefined);

interface LeagueProviderProps {
    children: ReactNode;
}

interface LeagueState {
    currentLeague: League | null;
    currentDraft: Draft | null;
    currentPlayer: Player | null;
    myDiscordUser: DiscordUser | null;
    loading: boolean;
    error: string | null;
}

export const LeagueProvider = ({ children }: LeagueProviderProps) => {
    const { leagueId } = useParams<{ leagueId: string }>();
    const [state, setState] = useState<LeagueState>({
        currentLeague: null,
        currentDraft: null,
        currentPlayer: null,
        myDiscordUser: null,
        loading: true,
        error: null,
    });

    const fetchData = useCallback(async () => {
        if (!leagueId) {
            setState(prevState => ({
                ...prevState,
                currentLeague: null,
                currentDraft: null,
                currentPlayer: null,
                myDiscordUser: null,
                loading: false,
            }));
            return;
        }

        setState(prevState => ({ ...prevState, loading: true, error: null }));

        try {
            const leagueResponse = await getLeague(leagueId);
            const draftResponse = await getDraftByLeagueID(leagueId);
            const discordUserResponse = await getMyDiscordDetails();

            let playerInCurrentLeague: Player | null = null;
            if (discordUserResponse.data?.ID) {
                const playersResponse = await getPlayersByUserId(discordUserResponse.data.ID);
                playerInCurrentLeague = playersResponse.data.find((p: Player) => p.LeagueID === leagueId) || null;
            }

            let currentDraftWithPlayer = draftResponse.data;
            if (currentDraftWithPlayer && currentDraftWithPlayer.CurrentTurnPlayerID) {
                try {
                    const turnPlayerResponse = await getPlayerById(leagueId, currentDraftWithPlayer.CurrentTurnPlayerID);
                    currentDraftWithPlayer = { ...currentDraftWithPlayer, CurrentTurnPlayer: turnPlayerResponse.data };
                } catch (playerErr) {
                    console.error("Failed to fetch CurrentTurnPlayer:", playerErr);
                    // Optionally, handle this error more gracefully, e.g., set CurrentTurnPlayer to null
                }
            }

            setState(prevState => ({
                ...prevState,
                currentLeague: leagueResponse.data,
                currentDraft: currentDraftWithPlayer,
                myDiscordUser: discordUserResponse.data,
                currentPlayer: playerInCurrentLeague,
                loading: false,
                error: null,
            }));

        } catch (err) {
            let errorMessage = "A network or unknown error occurred while fetching league data.";
            if (axios.isAxiosError(err) && err.response) {
                errorMessage = err.response.data.error || errorMessage;
            }
            setState(prevState => ({
                ...prevState,
                currentLeague: null,
                currentDraft: null,
                currentPlayer: null,
                myDiscordUser: null,
                loading: false,
                error: errorMessage,
            }));
            console.error("LeagueProvider: Error fetching league data:", err);
        }
    }, [leagueId]);

    useEffect(() => {
        fetchData();
    }, [fetchData]);

    const contextValue: LeagueContextType = {
        currentLeague: state.currentLeague,
        setCurrentLeague: (league) => setState(prevState => ({ ...prevState, currentLeague: league })),
        currentDraft: state.currentDraft,
        currentPlayer: state.currentPlayer,
        myDiscordUser: state.myDiscordUser,
        loading: state.loading,
        error: state.error,
        refetch: fetchData,
    };

    return (
        <LeagueContext.Provider value={contextValue}>
            {children}
        </LeagueContext.Provider>
    );
};


// Custom hook for easy consumption of the LeagueContext
export const useLeague = () => {
    const context = useContext(LeagueContext);
    if (context === undefined) {
        throw new Error('useLeague must be used within a LeagueProvider');
    }
    return context;
};
