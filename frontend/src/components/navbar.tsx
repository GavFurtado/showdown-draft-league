import { Disclosure } from '@headlessui/react';
import { useNavigate } from 'react-router-dom';
import { useLeague } from '../context/LeagueContext';
import { useEffect, useState } from 'react';
import { League, Player, User } from '../api/data_interfaces';
import { DiscordUser } from '../api/request_interfaces';
import { getMyLeagues, getUserMe, getMyDiscordDetails, getPlayersByLeague, logout } from '../api/api';
import axios from 'axios'; // Import axios for error handling

// Import the new sub-components
import LeagueDropdown from './LeagueDropdown';
import GlobalNavLinks from './GlobalNavLinks';
import UserAuthSection from './UserAuthSection';
import LeagueSubNav from './LeagueSubNav';

interface NavBarProps {
    page?: string;
}

export default function NavBar({ page }: NavBarProps) {
    const { currentLeague } = useLeague();
    const navigate = useNavigate();

    // User-related states
    const [user, setUser] = useState<User | null>(null);
    const [discordUser, setDiscordUser] = useState<DiscordUser | null>(null);
    const [userLoading, setUserLoading] = useState<boolean>(true);
    const [userError, setUserError] = useState<string | null>(null);

    // User's league states (for dropdown)
    const [userLeagues, setUserLeagues] = useState<League[]>([]);
    const [userLeaguesLoading, setUserLeaguesLoading] = useState<boolean>(true);
    const [userLeaguesError, setUserLeaguesError] = useState<string | null>(null);

    // Current user's Player object for the active league (for league staff roles)
    const [userPlayer, setUserPlayer] = useState<Player | null>(null);
    const [userPlayerLoading, setUserPlayerLoading] = useState<boolean>(true);
    const [userPlayerError, setUserPlayerError] = useState<string | null>(null);

    const logoPic = "https://www.elitefourum.com/uploads/default/original/3X/4/b/4bbe5270ed2b07d84730959af8819f255a922ea0.png";

    // Handle Logout
    const handleLogout = async () => {
        try {
            await logout();
            setUser(null)
            setDiscordUser(null);
            navigate("/login")
        } catch (error) {
            console.error('Logout failed: ', error);
        }
    };

    // Handle League Selection from Dropdown
    const handleLeagueSelect = (selectedLeagueId: string) => {
        console.log(`Navigating to league: ${selectedLeagueId}`);
        navigate(`/league/${selectedLeagueId}/dashboard`);
    };

    // Effect to fetch user data (User and DiscordUser)
    useEffect(() => {
        // console.log("NavBar: useEffect for user data running.");
        const fetchUserData = async () => {
            // console.log("NavBar: fetchUserData called.");
            try {
                setUserLoading(true);
                setUserError(null);

                const userResponse = await getUserMe();
                // console.log("NavBar:: getUserMe response:", userResponse.data);
                setUser(userResponse.data);

                const discordResponse = await getMyDiscordDetails();
                // console.log("NavBar:: getMyDiscordDetails response:", discordResponse.data);
                setDiscordUser(discordResponse.data);
            } catch (err) {
                if (axios.isAxiosError(err) && err.response) {
                    setUserError(err.response.data.error || "Failed to load user data.");
                } else {
                    setUserError("A network or unknown error occurred while fetching user data.");
                }
                console.error("NavBar:: Error fetching user data:", err);
            } finally {
                setUserLoading(false);
                console.log("NavBar:: fetchUserData finished. User loading:", userLoading);
            }
        };
        fetchUserData();
    }, []); // Run once on component mount

    // Effect to fetch user's leagues
    useEffect(() => {
        // console.log("NavBar: useEffect for user leagues running. Current user:", user);
        const fetchUserLeagues = async () => {
            // console.log("NavBar: fetchUserLeagues called.");
            if (!user?.id) {
                // console.log("NavBar: No user ID, skipping fetchUserLeagues.");
                setUserLeagues([]);
                setUserLeaguesLoading(false);
                return;
            }

            try {
                setUserLeaguesLoading(true);
                setUserLeaguesError(null);
                const response = await getMyLeagues();
                // console.log("NavBar: getMyLeagues response:", response.data);
                setUserLeagues(response.data);
            } catch (err) {
                if (axios.isAxiosError(err) && err.response) {
                    setUserLeaguesError(err.response.data.error || "Failed to load user's leagues.");
                } else {
                    setUserLeaguesError("A network or unknown error occurred while fetching user's leagues.");
                }
                // console.error("NavBar: Error fetching user leagues:", err);
            } finally {
                setUserLeaguesLoading(false);
                // console.log("NavBar: fetchUserLeagues finished. User leagues loading:", false);
            }
        };

        if (user) {
            fetchUserLeagues();
        } else {
            // console.log("NavBar: No user, setting userLeagues to empty.");
            setUserLeagues([]);
            setUserLeaguesLoading(false);
        }
    }, [user]); // Re-fetch when user changes (i.e., logs in/out)

    // Effect to fetch current user's Player object for the active league
    useEffect(() => {
        // console.log("NavBar: useEffect for user player running. User ID:", user?.id, "League ID:", currentLeague?.id);
        const fetchUserPlayer = async () => {
            // console.log("NavBar: fetchUserPlayer called.");
            if (!user?.id || !currentLeague?.id) {
                // console.log("NavBar: Missing user ID or current league ID, skipping fetchUserPlayer.");
                setUserPlayer(null);
                setUserPlayerLoading(false);
                return;
            }

            try {
                setUserPlayerLoading(true);
                setUserPlayerError(null);

                const response = await getPlayersByLeague(currentLeague.id);
                // console.log("NavBar: getPlayersByLeague response:", response.data);
                const allPlayersInLeague: Player[] = response.data;
                const foundPlayer = allPlayersInLeague.find(
                    (player) => player.userId === user.id
                );

                setUserPlayer(foundPlayer || null);
            } catch (err) {
                if (axios.isAxiosError(err) && err.response) {
                    setUserPlayerError(err.response.data.error || "Failed to load player data for this league.");
                } else {
                    setUserPlayerError("A network or unknown error occurred while fetching player data.");
                }
                // console.error("NavBar: Error fetching user player data:", err);
            } finally {
                setUserPlayerLoading(false);
                // console.log("NavBar: fetchUserPlayer finished. User player loading:", false);
            }
        };
        fetchUserPlayer();
    }, [user?.id, currentLeague?.id]); // Re-fetch if user or currentLeague changes

    return (
        <>
            {/* Top Level Navbar */}
            <Disclosure as="nav" className="bg-[#2D3142]">
                <div className="mx-auto max-w-7xl px-2 sm:px-6 lg:px-8">
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
                    userPlayer={userPlayer}
                    currentPage={page}
                />
            )}
        </>
    );
}
