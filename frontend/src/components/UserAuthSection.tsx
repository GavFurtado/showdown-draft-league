import { Link } from 'react-router-dom';
import { User } from '../api/data_interfaces';
import { DiscordUser } from '../api/request_interfaces';

function mergeClasses(...classes: (string | boolean | undefined | null)[]) {
    return classes.filter(Boolean).join(' ');
}

interface UserAuthSectionProps {
    user: User | null;
    discordUser: DiscordUser | null;
    loading: boolean;
    error: string | null;
    onLogout: () => void;
}

export default function UserAuthSection({
    user,
    discordUser,
    loading,
    error,
    onLogout,
}: UserAuthSectionProps) {
    if (loading) {
        return <span className="text-gray-400">Loading user...</span>;
    }

    if (error) {
        return <span className="text-red-400">User Error</span>;
    }

    return (
        <div className="absolute inset-y-0 right-0 flex items-center pr-2 sm:static sm:inset-auto sm:ml-6 sm:pr-0">
            {user && discordUser ? (
                <div className="flex items-center space-x-3">
                    <img
                        className="h-8 w-8 rounded-full"
                        src={discordUser.avatar}
                        alt="User avatar"
                    />
                    <span className="text-white text-sm font-medium hidden md:block">
                        {discordUser.username}#{discordUser.discriminator}
                    </span>
                    <button
                        onClick={onLogout}
                        className={mergeClasses('text-gray-300 hover:bg-gray-700 hover:text-white rounded-md px-3 py-2 text-sm font-medium')}
                    >
                        Logout
                    </button>
                </div>
            ) : (
                <Link
                    to="/login"
                    className={mergeClasses('text-gray-300 hover:bg-gray-700 hover:text-white rounded-md px-3 py-2 text-sm font-medium')}
                >
                    Login
                </Link>
            )}
        </div>
    );
}
