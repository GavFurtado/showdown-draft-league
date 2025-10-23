import { createContext, useContext, useState, useEffect, ReactNode, useCallback, useMemo } from 'react';
import { League, Draft, Player } from '../api/data_interfaces';
import { getLeagueById, getDraftByLeagueID, getPlayerById, getPlayerByUserIdAndLeagueId } from '../api/api';
import axios from 'axios';
import { useParams } from 'react-router-dom';
import { useUser } from './UserContext'; // Import useUser

interface LeagueContextType {
    currentLeague: League | null;
    setCurrentLeague: (league: League | null) => void;
    currentDraft: Draft | null;
    currentPlayer: Player | null;
    // myDiscordUser: DiscordUser | null; // Removed as it comes from UserContext
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
    // myDiscordUser: DiscordUser | null; // Removed
    loading: boolean;
    error: string | null;
}

export const LeagueProvider = ({ children }: LeagueProviderProps) => {
    const { leagueId } = useParams<{ leagueId: string }>();
    const { user, discordUser, loading: userLoading, error: userError } = useUser(); // Consume UserContext

    const [state, setState] = useState<LeagueState>({
        currentLeague: null,
        currentDraft: null,
        currentPlayer: null,
        // myDiscordUser: null, // Removed
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
                // myDiscordUser: null, // Removed
                loading: false,
            }));
            return;
        }

        setState(prevState => ({ ...prevState, loading: true, error: null }));

        try {
            const leagueData = await getLeagueById(leagueId);
            const draftData = await getDraftByLeagueID(leagueId);
            // const discordUserData = await getMyDiscordDetails(); // Removed

            let playerInCurrentLeague: Player | null = null;
            if (discordUser?.ID) { // Use discordUser from UserContext
                const playerResponse = await getPlayerByUserIdAndLeagueId(leagueId, discordUser.ID);
                playerInCurrentLeague = playerResponse.data;
            }

            let currentDraftWithPlayer = draftData.data;
            if (currentDraftWithPlayer && currentDraftWithPlayer.CurrentTurnPlayerID) {
                try {
                    const turnPlayerResponse = await getPlayerById(leagueId, currentDraftWithPlayer.CurrentTurnPlayerID);
                    currentDraftWithPlayer = { ...currentDraftWithPlayer, CurrentTurnPlayer: turnPlayerResponse.data };
                } catch (playerErr) {
                    console.error("Failed to fetch CurrentTurnPlayer:", playerErr);
                }
            }

            setState(prevState => ({
                ...prevState,
                currentLeague: leagueData.data,
                currentDraft: currentDraftWithPlayer,
                // myDiscordUser: discordUserData.data, // Removed
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
                // myDiscordUser: null, // Removed
                loading: false,
                error: errorMessage,
            }));
            console.error("LeagueProvider: Error fetching league data:", err);
        }
    }, [leagueId, discordUser?.ID]); // Add discordUser.ID to dependencies

    useEffect(() => {
        // Only fetch if user data is not loading and no user error
        if (!userLoading && !userError) {
            fetchData();
        }
    }, [fetchData, userLoading, userError]);

    const contextValue: LeagueContextType = {
        currentLeague: state.currentLeague,
        setCurrentLeague: (league) => setState(prevState => ({ ...prevState, currentLeague: league })),
        currentDraft: state.currentDraft,
        currentPlayer: state.currentPlayer,
        // myDiscordUser: state.myDiscordUser, // Removed
        loading: state.loading || userLoading, // Combine loading states
        error: state.error || userError, // Combine error states
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
        // Return a default context when not within a LeagueProvider
        return {
            currentLeague: null,
            setCurrentLeague: () => {},
            currentDraft: null,
            currentPlayer: null,
            loading: false,
            error: null,
            refetch: () => {},
        };
    }
    return context;
};
