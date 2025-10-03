// frontend/src/context/LeagueContext.tsx

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { League } from '../api/data_interfaces';
import { getLeague } from '../api/api';
import axios from 'axios'; // Import axios for error handling
import { useParams } from 'react-router-dom'; // Import useParams

interface LeagueContextType {
  currentLeague: League | null;
  setCurrentLeague: (league: League | null) => void;
  loading: boolean;
  error: string | null;
}

export const LeagueContext = createContext<LeagueContextType | undefined>(undefined);

interface LeagueProviderProps {
  children: ReactNode;
  // leagueId?: string; // Removed as useParams will be used internally
}

export const LeagueProvider = ({ children }: LeagueProviderProps) => {
  const { leagueId } = useParams<{ leagueId: string }>(); // Get leagueId from URL here
  const [currentLeague, setCurrentLeague] = useState<League | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  console.log("LeagueProvider: Component rendered. leagueId from URL:", leagueId);

  useEffect(() => {
    console.log("LeagueProvider: useEffect for fetchLeague running. leagueId:", leagueId);
    const fetchLeague = async () => {
      console.log("LeagueProvider: fetchLeague called.");
      if (!leagueId) {
        console.log("LeagueProvider: No leagueId, setting currentLeague to null.");
        setCurrentLeague(null);
        setLoading(false);
        return;
      }

      try {
        setLoading(true);
        setError(null);
        console.log(`LeagueProvider: Attempting to fetch league with ID: ${leagueId}`);
        const response = await getLeague(leagueId);
        console.log("LeagueProvider: getLeague response:", response.data);
        setCurrentLeague(response.data);
      } catch (err) {
        if (axios.isAxiosError(err) && err.response) {
          setError(err.response.data.error || "Failed to load league data.");
        } else {
          setError("A network or unknown error occurred while fetching league data.");
        }
        console.error("LeagueProvider: Error fetching league data:", err);
        setCurrentLeague(null); // Ensure no partial or incorrect data is set
      } finally {
        setLoading(false);
        console.log("LeagueProvider: fetchLeague finished. Loading:", false);
      }
    };

    fetchLeague();
  }, [leagueId]); // leagueId dependency; Re-fetch if leagueId changes

  const contextValue: LeagueContextType = {
    currentLeague,
    setCurrentLeague,
    loading,
    error,
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
