import React from 'react';
import { Player } from '../api/data_interfaces';
import { getTopPlayers } from '../utils/standingsUtils';

interface DashboardStandingsProps {
    players: Player[];
}

export const DashboardStandings: React.FC<DashboardStandingsProps> = ({ players }) => {
    // Take top 5 for dashboard using utility
    const topPlayers = getTopPlayers(players, 5);

    if (players.length === 0) {
        return <div className="text-text-secondary text-sm">No standings available.</div>;
    }

    return (
        <div className="flex flex-col h-full overflow-hidden">
            <div className="grow overflow-y-auto">
                <table className="min-w-full text-left text-sm whitespace-nowrap">
                    <thead className="uppercase tracking-wider border-b border-gray-200 bg-gray-50">
                        <tr>
                            <th scope="col" className="px-4 py-3 font-semibold text-text-secondary">Pos</th>
                            <th scope="col" className="px-4 py-3 font-semibold text-text-secondary">Team</th>
                            <th scope="col" className="px-4 py-3 font-semibold text-text-secondary text-right">W-L</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100">
                        {topPlayers.map((player, index) => (
                            <tr key={player.ID} className="hover:bg-gray-50 transition-colors">
                                <td className="px-4 py-3 font-medium text-text-primary">{index + 1}</td>
                                <td className="px-4 py-3 text-text-secondary truncate max-w-[150px]" title={player.TeamName}>
                                    {player.TeamName}
                                    <span className="block text-xs text-gray-400">{player.InLeagueName}</span>
                                </td>
                                <td className="px-4 py-3 text-text-primary font-medium text-right">
                                    {player.Wins}-{player.Losses}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
            <div className="bg-gray-50 px-4 py-3 text-right">
                <span className="text-xs font-medium text-accent-primary hover:text-accent-primary-hover cursor-pointer">
                    Full Standings &rarr;
                </span>
            </div>
        </div>
    );
};
