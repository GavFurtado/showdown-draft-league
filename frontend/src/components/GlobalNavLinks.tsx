import { Link } from 'react-router-dom';
import { User } from '../api/data_interfaces';

interface NavigationItem {
    name: string;
    href: string;
    current: boolean;
}

const globalNavigation: NavigationItem[] = [
    { name: 'My Leagues', href: '/my-leagues', current: false },
    { name: 'FAQ', href: '/faq', current: false },
];

function mergeClasses(...classes: (string | boolean | undefined | null)[]) {
    return classes.filter(Boolean).join(' ');
}

interface GlobalNavLinksProps {
    user: User | null;
    currentPage: string | undefined;
}

export default function GlobalNavLinks({ user, currentPage }: GlobalNavLinksProps) {
    return (
        <div className="flex space-x-2">
            {globalNavigation.map((item) => (
                <Link
                    key={item.name}
                    to={item.href}
                    aria-current={currentPage === item.name ? 'page' : undefined}
                    className={mergeClasses(
                        currentPage === item.name ? 'bg-gray-900 text-white' : 'text-gray-300 hover:bg-gray-700 hover:text-white',
                        'rounded-md px-3 py-2 text-sm font-medium',
                    )}
                >
                    {item.name}
                </Link>
            ))}
            {user && user.Role === "admin" && (
                <Link to="/admin/dashboard" className={mergeClasses(
                    currentPage === "Admin Dashboard" ? 'bg-gray-900 text-white' : 'text-gray-300 hover:bg-gray-700 hover:text-white',
                    'rounded-md px-3 py-2 text-sm font-medium',
                )}>
                    Admin Dashboard
                </Link>
            )}
        </div>
    );
}
