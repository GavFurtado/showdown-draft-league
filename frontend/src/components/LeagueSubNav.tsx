import { Disclosure } from '@headlessui/react';
import { Link } from 'react-router-dom';
import { League, Player } from '../api/data_interfaces';

interface NavigationItem {
    name: string;
    href: string;
    current: boolean;
}

const leagueNavigation: NavigationItem[] = [
    { name: 'Dashboard', href: 'dashboard', current: false },
    { name: 'Team Sheets', href: 'teamsheets', current: false },
    { name: 'Draftboard', href: 'draftboard', current: false },
    { name: 'Standings', href: 'standings', current: false },
];

function mergeClasses(...classes: (string | boolean | undefined | null)[]) {
    return classes.filter(Boolean).join(' ');
}

interface LeagueSubNavProps {
    currentLeague: League;
    userPlayer: Player | null;
    currentPage: string | undefined;
}

export default function LeagueSubNav({
    currentLeague,
    userPlayer,
    currentPage,
}: LeagueSubNavProps) {
    const isLeagueStaff = userPlayer && (userPlayer.Role === "moderator" || userPlayer.Role === "owner");

    return (
        <Disclosure as="nav" className="bg-[#4F5D75] shadow">
            <div className="mx-auto max-w-7xl px-2 sm:px-6 lg:px-8">
                <div className="flex h-10 items-center justify-start">
                    <div className="flex space-x-4">
                        {leagueNavigation.map((item) => (
                            <Link
                                key={item.name}
                                to={`/league/${currentLeague.ID}/${item.href}`}
                                aria-current={currentPage === item.name ? 'page' : undefined}
                                className={mergeClasses(
                                    currentPage === item.name ? 'bg-gray-700 text-white' : 'text-gray-200 hover:bg-gray-600 hover:text-white',
                                    'rounded-md px-3 py-1 text-sm font-medium',
                                )}
                            >
                                {item.name}
                            </Link>
                        ))}
                        {isLeagueStaff && (
                            <>

                                <Link to={`/league/${currentLeague.ID}/staff/edit-rules`} className={mergeClasses(
                                    currentPage === "Edit Rules" ? 'bg-gray-700 text-white' : 'text-gray-200 hover:bg-gray-600 hover:text-white',
                                    'rounded-md px-3 py-1 text-sm font-medium',
                                )}>
                                    Edit Rules
                                </Link>
                            </>
                        )}
                    </div>
                </div>
            </div>
        </Disclosure>
    );
}
