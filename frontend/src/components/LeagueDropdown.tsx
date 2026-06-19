import { Menu, MenuButton, MenuItem, MenuItems, Transition } from '@headlessui/react';
import { Link } from 'react-router-dom';
import { Fragment } from 'react';
import { League, Player } from '../api/data_interfaces';

interface LeagueDropdownProps {
    currentLeague: League | null;
    currentPlayer: Player | null;
    userLeagues: League[];
    loading: boolean;
    error: string | null;
    onSelectLeague: (leagueId: string) => void;
}

function mergeClasses(...classes: (string | boolean | undefined | null)[]) {
    return classes.filter(Boolean).join(' ');
}

export default function LeagueDropdown({
    currentLeague,
    currentPlayer,
    userLeagues,
    loading,
    error,
    onSelectLeague,
}: LeagueDropdownProps) {
    if (loading) {
        return <span className="ml-4 text-gray-400 text-lg">Loading Leagues...</span>;
    }

    if (error) {
        return <span className="ml-4 text-red-400 text-lg">Error loading leagues</span>;
    }

    return (
        <Menu as="div" className="relative ml-3">
            <div>
                <MenuButton className="flex items-center rounded-md bg-gray-800 px-3 py-2 text-sm font-medium text-white hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-white focus:ring-offset-2 focus:ring-offset-gray-800 ml-1">
                    <span className="sr-only">Open league menu</span>
                    <div className="flex flex-col items-start">
                        {currentLeague ? (
                            <>
                                <span>{currentLeague.Name}</span>
                                {currentPlayer && (
                                    <span className="text-xs text-gray-400">{currentPlayer.TeamName} ({currentPlayer.InLeagueName})</span>
                                )}
                            </>
                        ) : (
                            "Select League"
                        )}
                    </div>
                    <svg className="h-5 w-5 ml-2 text-gray-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                        <path fillRule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 10.94l3.71-3.71a.75.75 0 111.06 1.06l-4.25 4.25a.75.75 0 01-1.06 0L5.23 8.29a.75.75 0 01.02-1.06z" clipRule="evenodd" />
                    </svg>
                </MenuButton>
            </div>
            <Transition
                as={Fragment}
                enter="transition ease-out duration-100"
                enterFrom="transform opacity-0 scale-95"
                enterTo="transform opacity-100 scale-100"
                leave="transition ease-in duration-75"
                leaveFrom="transform opacity-100 scale-100"
                leaveTo="transform opacity-0 scale-95"
            >
                <MenuItems className="absolute left-0 z-10 mt-2 w-48 origin-top-right rounded-md bg-white py-1 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
                    {userLeagues.length > 0 ? (
                        userLeagues.map((league) => (
                            <MenuItem key={league.ID}>
                                {({ active }) => (
                                    <Link
                                        to={`/league/${league.ID}/dashboard`}
                                        onClick={() => onSelectLeague(league.ID)}
                                        className={mergeClasses(
                                            active ? 'bg-gray-100' : '',
                                            'block px-4 py-2 text-sm text-gray-700'
                                        )}
                                    >
                                        {league.Name}
                                    </Link>
                                )}
                            </MenuItem>
                        ))
                    ) : (
                        <MenuItem disabled>
                            <span className="block px-4 py-2 text-sm text-gray-500">No leagues found</span>
                        </MenuItem>
                    )}
                    <MenuItem>
                        {({ active }) => (
                            <Link
                                to="/my-leagues"
                                className={mergeClasses(
                                    active ? 'bg-gray-100' : '',
                                    'block px-4 py-2 text-sm text-gray-700 border-t border-gray-200 mt-1'
                                )}
                            >
                                View All My Leagues
                            </Link>
                        )}
                    </MenuItem>
                </MenuItems>
            </Transition>
        </Menu>
    );
}
