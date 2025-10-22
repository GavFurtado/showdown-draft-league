import { createContext, useContext, useState, useEffect, ReactNode, useCallback } from 'react';
import { User, DiscordUser } from '../api/data_interfaces';
import { getUserMe, getMyDiscordDetails, logout as apiLogout } from '../api/api';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';

interface UserContextType {
    user: User | null;
    discordUser: DiscordUser | null;
    loading: boolean;
    error: string | null;
    logout: () => void;
}

export const UserContext = createContext<UserContextType | undefined>(undefined);

interface UserProviderProps {
    children: ReactNode;
}

export const UserProvider = ({ children }: UserProviderProps) => {
    const [user, setUser] = useState<User | null>(null);
    const [discordUser, setDiscordUser] = useState<DiscordUser | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);
    const navigate = useNavigate();

    useEffect(() => {
        const fetchUserData = async () => {
            try {
                setLoading(true);
                setError(null);

                const userResponse = await getUserMe();
                setUser(userResponse.data);

                const discordResponse = await getMyDiscordDetails();
                setDiscordUser(discordResponse.data);
            } catch (err) {
                if (axios.isAxiosError(err) && err.response?.status !== 401) {
                    setError(err.response?.data?.error || "Failed to load user data.");
                } else if (!axios.isAxiosError(err)) {
                    setError("A network or unknown error occurred while fetching user data.");
                }
                // We don't set an error for 401, as it simply means the user is not logged in.
            } finally {
                setLoading(false);
            }
        };
        fetchUserData();
    }, []);

    const handleLogout = useCallback(async () => {
        try {
            await apiLogout();
            setUser(null);
            setDiscordUser(null);
            navigate("/login");
        } catch (error) {
            console.error('Logout failed: ', error);
        }
    }, [navigate]);

    const contextValue: UserContextType = {
        user,
        discordUser,
        loading,
        error,
        logout: handleLogout,
    };

    return (
        <UserContext.Provider value={contextValue}>
            {children}
        </UserContext.Provider>
    );
};

export const useUser = () => {
    const context = useContext(UserContext);
    if (context === undefined) {
        throw new Error('useUser must be used within a UserProvider');
    }
    return context;
};
