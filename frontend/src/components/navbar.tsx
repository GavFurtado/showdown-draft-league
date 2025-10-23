import { Disclosure } from '@headlessui/react';
import { useNavigate } from 'react-router-dom';
import { useLeague } from '../context/LeagueContext';
import { useUser } from '../context/UserContext'; // Import useUser
import { useEffect, useState } from 'react';
import { League } from '../api/data_interfaces';
import { getMyLeagues } from '../api/api';
import axios from 'axios';

// Import the new sub-components
import LeagueDropdown from './LeagueDropdown';
import GlobalNavLinks from './GlobalNavLinks';
import UserAuthSection from './UserAuthSection';
import LeagueSubNav from './LeagueSubNav';

interface NavBarProps {
    page?: string;
}

export default function NavBar({ page }: NavBarProps) {
    const { currentLeague, currentPlayer } = useLeague(); // Get currentPlayer from LeagueContext
    const { user, discordUser, loading: userLoading, error: userError, logout } = useUser(); // Consume UserContext
    const navigate = useNavigate();

    // User's league states (for dropdown)
    const [userLeagues, setUserLeagues] = useState<League[]>([]);
    const [userLeaguesLoading, setUserLeaguesLoading] = useState<boolean>(true);
    const [userLeaguesError, setUserLeaguesError] = useState<string | null>(null);

    const logoPic = "https://www.elitefourum.com/uploads/default/original/3X/4/b/4bbe5270ed2b07d84730959af8819f255a922ea0.png";

    // Handle Logout
    const handleLogout = async () => {
        logout(); // Use logout from UserContext
    };

    // Handle League Selection from Dropdown
    const handleLeagueSelect = (selectedLeagueId: string) => {
        console.log(`Navigating to league: ${selectedLeagueId}`);
        navigate(`/league/${selectedLeagueId}/dashboard`);
    };

    // Effect to fetch user's leagues
    useEffect(() => {
        const fetchUserLeagues = async () => {
            if (!user?.ID) {
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
            setUserLeagues([]);
            setUserLeaguesLoading(false);
        }
    }, [user]); // Re-fetch when user changes (i.e., logs in/out)

    return (
        <>
            {/* Top Level Navbar */}
            <Disclosure as="nav" className="bg-[#2D3142]">
                {/* <div className="mx-auto max-w-7xl px-2 sm:px-6 lg:px-8"> */}
                <div className="mx-auto px-2 sm:px-6 lg:px-8">
                    <div className="relative flex h-16 items-center justify-between">
                        {/* Left Section: Logo & League Dropdown */}
                        <div className="flex items-center">
                            <img alt="Logo" src={logoPic} className="h-8 w-auto" />
                            <LeagueDropdown
                                currentLeague={currentLeague}
                                userLeagues={userLeagues}
                                loading={userLeaguesLoading}
                                error={userLeaguesError}
                                onSelectLeague={handleLeagueSelect}
                            />
                        </div>

                        {/* Center Section: Global Navigation Links */}
                        <div className="flex flex-1 justify-start">
                            <GlobalNavLinks user={user} currentPage={page} />
                        </div>

                        {/* Right Section: User Info & Logout */}
                        <div className="absolute inset-y-0 right-0 flex items-center pr-2 sm:static sm:inset-auto sm:ml-6 sm:pr-0">
                            <UserAuthSection
                                user={user}
                                discordUser={discordUser}
                                loading={userLoading}
                                error={userError}
                                onLogout={handleLogout}
                            />
                        </div>
                    </div>
                </div>
            </Disclosure>

            {/* Second Level Navbar (League-Specific Navigation) */}
            {currentLeague && (
                <LeagueSubNav
                    currentLeague={currentLeague}
                    userPlayer={currentPlayer} // Use currentPlayer from LeagueContext
                    currentPage={page}
                />
            )}
        </>
    );
}
