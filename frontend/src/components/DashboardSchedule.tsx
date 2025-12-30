import React from 'react';
import { Game } from '../api/data_interfaces';
import { getUpcomingGames } from '../utils/scheduleUtils';

interface DashboardScheduleProps {
    games: Game[];
    currentPlayerId: string;
}

export const DashboardSchedule: React.FC<DashboardScheduleProps> = ({ games, currentPlayerId }) => {
    // Filter for active or upcoming games using utility
    const displayGames = getUpcomingGames(games, 3);
    // console.log(displayGames);

    if (displayGames.length === 0) {
        return <div className="text-text-secondary text-sm p-4">No upcoming games scheduled.</div>;
    }

    return (
        <div className="flex flex-col h-full overflow-hidden">
            <div className="grow overflow-y-auto">
                {displayGames.map((game) => {
                    const isPlayer1 = game.Player1ID === currentPlayerId;
                    const opponentTeamName = isPlayer1 ? game.Player2?.TeamName || "Opponent" : game.Player1?.TeamName || "Opponent";
                    const opponentPlayerName = isPlayer1 ? game.Player2?.InLeagueName : game.Player1?.InLeagueName;
                    {/* console.log(`Game (ID: ${game.ID}) Status: ${game.Status}`) */ }

                    // Determine status badge
                    let statusColor = "bg-gray-100 text-gray-800";
                    if (game.Status === "SCHEDULED") statusColor = "bg-blue-100 text-blue-800";
                    if (game.Status === "COMPLETED") statusColor = "bg-green-100 text-green-800";
                    if (game.Status === "DISPUTED") statusColor = "bg-red-100 text-red-800";

                    return (
                        <div key={game.ID} className="p-4 hover:bg-gray-50 transition-colors">
                            <div className="flex justify-between items-start mb-1">
                                <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
                                    Week {game.RoundNumber}
                                </span>
                                <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${statusColor}`}>
                                    {game.Status}
                                </span>
                            </div>
                            <div className="flex items-center justify-between mt-2">
                                <div>
                                    <p className="text-sm font-medium text-text-primary">vs. {opponentTeamName}</p>
                                    <p className="text-xs text-text-secondary">{opponentPlayerName}</p>
                                </div>
                                {game.Status === 'COMPLETED' ? (
                                    <div className="text-sm font-bold text-text-primary">
                                        {isPlayer1 ? game.Player1Wins : game.Player2Wins} - {isPlayer1 ? game.Player2Wins : game.Player1Wins}
                                    </div>
                                ) : (
                                    <div className="text-xs text-text-secondary italic">
                                        TBD
                                    </div>
                                )}
                            </div>
                        </div>
                    );
                })}
            </div>
            <div className="bg-gray-50 px-4 py-3 text-right">
                <span className="text-xs font-medium text-accent-primary hover:text-accent-primary-hover cursor-pointer">
                    Full Schedule &rarr;
                </span>
            </div>
        </div>
    );
}
