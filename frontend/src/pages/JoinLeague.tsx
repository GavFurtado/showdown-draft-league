import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { getLeagueById, getPlayerByUserIdAndLeagueId, joinLeague } from "../api/api";
import { League } from "../api/data_interfaces";
import { JoinLeagueRequest } from "../api/request_interfaces.ts"
import Layout from "../components/Layout";
import axios from "axios";
import Modal from "../components/Modal";
import { format } from "date-fns";
import { useUser } from "../context/UserContext"; // Import useUser
import { is } from "date-fns/locale";

interface LeagueDetailRowProps {
    label: string;
    value: React.ReactNode;
}

const LeagueDetailRow: React.FC<LeagueDetailRowProps> = ({ label, value }) => (
    <div className="flex justify-between items-center">
        <span className="font-bold">{label}:</span>
        <span className="font-mono font-medium">{value}</span>
    </div>
);

interface LeagueDetailsDisplayProps {
    league: League;
}

const LeagueDetailsDisplay: React.FC<LeagueDetailsDisplayProps> = ({ league }) => {
    return (
        <div className="space-y-2 text-m">
            <LeagueDetailRow
                label="Start Date"
                value={<span className="truncate font-mono">{format(new Date(league.StartDate), 'MMM dd, yyyy hh:mm')} (Local Time)</span>}
            />
            <LeagueDetailRow label="League Status" value={league.Status} />
            <LeagueDetailRow label="# of Players" value={league.Players?.length} />
            <LeagueDetailRow label="Draft Rounds" value={league.MaxPokemonPerPlayer} />
            <LeagueDetailRow label="Draft Order" value={league.Format.IsSnakeRoundDraft ? "Snake" : "Linear"} />
            <LeagueDetailRow
                label="Skips Allowed during Draft"
                value={`${league.MaxPokemonPerPlayer - league.MinPokemonPerPlayer} (Max.)`}
            />
            <LeagueDetailRow
                label="Roster Size"
                value={`${league.MinPokemonPerPlayer}-${league.MaxPokemonPerPlayer} Pokémon`}
            />
            <LeagueDetailRow label="Transfer Credits" value={league.Format.AllowTransferCredits ? "Enabled" : "Disabled"} />
            {/* TODO: Ruleset Description shouldn't use LeagueDetailRow since it could be pretty detailed  */}
            <LeagueDetailRow label="Rules" value={league.RulesetDescription ? league.RulesetDescription : "Not specified"} />
            {/* "More Details" section Groups*/}
            <details className="px-2 py-1 group bg-background-surface border border-solid border-background-surface rounded-xl">
                <summary className="flex justify-between items-center cursor-pointer text-text-primary">
                    <span className="font-bold text-accent-primary">Game Details</span>
                    <span className="transition-transform group-open:-rotate-180">
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth="1.5"
                            stroke="currentColor"
                            className="size-4"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                        </svg>
                    </span>
                </summary>

                <div className="mt-2 space-y-2 text-m">
                    <LeagueDetailRow label="Season Type" value={league.Format.SeasonType} />
                    {league.Format.SeasonType !== "PLAYOFFS_ONLY" &&
                        <LeagueDetailRow label="# of Groups" value={league.Format.GroupCount} />
                    }
                    {league.Format.SeasonType !== "PLAYOFFS_ONLY" && league.Format.GamesPerOpponent !== 1 &&
                        < LeagueDetailRow label="Games Per Opponent" value={league.Format.GamesPerOpponent} />
                    }

                    {league.Format.SeasonType !== "ROUND_ROBIN_ONLY" &&
                        <LeagueDetailRow label="# of Playoff Participants" value={league.Format.PlayoffParticipantCount} />
                    }
                    {league.Format.SeasonType !== "ROUND_ROBIN_ONLY" &&
                        <LeagueDetailRow label="Playoffs Type" value={league.Format.PlayoffType} />
                    }
                    {league.Format.SeasonType !== "ROUND_ROBIN_ONLY" &&
                        <LeagueDetailRow label="Playoff Seeding Type" value={league.Format.PlayoffSeedingType} />
                    }
                </div>
            </details>
            {league.Format.AllowTransferCredits &&
                <details className="px-2 py-1 group bg-background-surface border border-background-surface rounded-xl">
                    <summary className="flex justify-between items-center cursor-pointer text-text-primary">
                        <span className="font-bold text-accent-primary">Transfer Credit System Details</span>
                        <span className="transition-transform group-open:-rotate-180">
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                fill="none"
                                viewBox="0 0 24 24"
                                strokeWidth="1.5"
                                stroke="currentColor"
                                className="size-4"
                            >
                                <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" />
                            </svg>
                        </span>
                    </summary>
                    <div className="mt-2 space-y-2 text-m">
                        <LeagueDetailRow label="Transfer Credits / Window" value={league.Format.TransferCreditsPerWindow} />
                        <LeagueDetailRow label="Max. Transfer Credits" value={league.Format.TransferCreditCap} />
                        <LeagueDetailRow label="Pickup Cost" value={league.Format.PickupCost} />
                        <LeagueDetailRow label="Drop Cost" value={league.Format.DropCost} />
                        <LeagueDetailRow label="Transfer Window Frequency" value={`Every ${league.Format.TransferWindowFrequencyDays} days`} />
                    </div>
                </details>
            }
        </div>
    );
};

const JoinLeague: React.FC = () => {
    const { leagueId } = useParams<{ leagueId: string }>();
    const { user, loading: userLoading, error: userError } = useUser();
    const [league, setLeague] = useState<League | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [joinClicked, setJoinClicked] = useState<boolean>(false);
    const [error, setError] = useState<string | null>(null);
    const [inLeagueName, setInLeagueName] = useState<string>("")
    const [teamName, setTeamName] = useState<string>("")
    const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
    const [isPlayerInLeague, setIsPlayerInLeague] = useState<boolean | null>(null); // New state

    useEffect(() => {
        const fetchLeagueData = async () => {
            // Only fetch if user data is not loading and no user error
            if (userLoading || userError) {
                setLoading(userLoading);
                setError(userError);
                return;
            }

            setLoading(true);
            setError(null);
            setIsPlayerInLeague(null); // Reset player status

            try {
                if (leagueId) {
                    const resp = await getLeagueById(leagueId);
                    setLeague(resp.data);

                    if (user?.ID) {
                        try {
                            await getPlayerByUserIdAndLeagueId(leagueId, user.ID);
                            setIsPlayerInLeague(true);
                        } catch (playerErr) {
                            if (axios.isAxiosError(playerErr) && playerErr.response?.status === 404) {
                                setIsPlayerInLeague(false);
                            } else {
                                // other errors during player check are actual errors
                                console.error("Error checking player status:", playerErr);
                                let playerErrorMessage = "Failed to check player status.";
                                if (axios.isAxiosError(playerErr) && playerErr.response) {
                                    playerErrorMessage = playerErr.response.data.error || playerErrorMessage;
                                }
                                setError(playerErrorMessage);
                            }
                        }
                    }
                }
            } catch (err) {
                let errorMessage = "A network or unknown error occurred while fetching league data.";
                if (axios.isAxiosError(err) && err.response) {
                    errorMessage = err.response.data.error || errorMessage;
                }
                console.error("failed to fetch league:", err);
                setError(errorMessage);
            } finally {
                setLoading(false);
            }
        };

        fetchLeagueData();
    }, [leagueId, user, userLoading, userError]); // Add user and its loading/error to dependencies

    const navigate = useNavigate();
    const handleCloseModal = () => {
        navigate(`/dashboard`);
    };
    const handleJoinButtonClick = () => {
        setJoinClicked(true);
    };

    // New function to handle form submission
    const handleJoinLeagueSubmit = async (e: React.FormEvent) => {
        e.preventDefault(); if (!leagueId) return;
        if (!user) return;

        setIsSubmitting(true);
        setError(null); // Clear previous errors

        try {
            const requestData: JoinLeagueRequest = { LeagueID: leagueId, UserID: user?.ID };
            if (inLeagueName.trim() !== '') {
                requestData.InLeagueName = inLeagueName.trim();
            }
            if (teamName.trim() !== '') {
                requestData.TeamName = teamName.trim();
            }

            await joinLeague(leagueId, requestData);
            navigate(`/league/${leagueId}/dashboard`); // Navigate to league dashboard on success
        } catch (err) {
            let errorMessage = "Failed to join league.";
            if (axios.isAxiosError(err) && err.response) {
                errorMessage = err.response.data.error || errorMessage;
            }
            setError(errorMessage); // Set error message
        } finally {
            setIsSubmitting(false); // Always reset submitting state
        }
    };
    if (loading) {
        return (
            <div className="flex justify-center items-center h-screen">
                {/* Spinner */}
                <svg className="animate-spin h-10 w-10 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
            </div>
        );
    }

    if (error) {
        return (
            <div className="flex justify-center items-center h-screen">
                <div className="text-red-500 text-center">{error}</div>
            </div>
        );
    }

    if (!league) {
        return (
            <div className="flex justify-center items-center h-screen">
                <div>League not found.</div>
            </div>
        );
    }

    return (
        <Layout variant="container">
            <div>
                {!joinClicked &&
                    <Modal isOpen={true} onClose={handleCloseModal}
                        title={<span>Join '<span className="text-yellow-500">{league.Name}</span>'?</span>}
                        background="bg-background-tertiary"
                        titleStyle="font-extrabold text-xl text-black"
                    >
                        <hr className="py-2 border-solid border-gray-950 opacity-90" />
                        <LeagueDetailsDisplay league={league} />
                        <div className="mt-6 flex justify-end gap-4">
                            <button
                                onClick={handleCloseModal}
                                disabled={joinClicked}
                                className="px-4 py-2 text-white bg-error-500 rounded hover:bg-background-secondary  hover:text-error-500 cursor-pointer"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleJoinButtonClick}
                                disabled={joinClicked || isPlayerInLeague as boolean}
                                className={`px-4 py-2 cursor-pointer rounded text-text-on-accent border-accent-primary bg-accent-primary ${!isPlayerInLeague ? "hover:bg-accent-primary-hover" : "bg-gray-300 cursor-not-allowed"}`}
                                title={isPlayerInLeague ? "You are already a part of this league" : "Join"}
                            >
                                Join
                            </button>
                        </div>
                    </Modal>
                }
                {joinClicked &&
                    <Modal
                        isOpen={true}
                        onClose={handleCloseModal}
                        title={<span>Set a name for '<span className="text-yellow-500">{league.Name}</span>'</span>}
                        showDefaultCloseButton={false}
                    >
                        <div className="p-4 mt-4 bg-blue-100 text-blue-800 rounded-lg shadow-md flex items-center space-x-2" role="alert">
                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" className="w-5 h-5 flex-shrink-0">
                                <path strokeLinecap="round" strokeLinejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
                            </svg>
                            <p className="text-sm font-medium">Your username will be used if the textboxes are left empty.</p>
                        </div>

                        <form className="flex flex-col justify-evenly mt-2" onSubmit={handleJoinLeagueSubmit}>
                            <div className="bg-background-tertiary p-4 rounded-lg">
                                <div className="mb-4">
                                    <label htmlFor="inLeagueName"
                                        className="block text-text-primary text-sm font-semibold mb-2">Your Name in League</label>
                                    <input
                                        type="text"
                                        id="inLeagueName"
                                        value={inLeagueName as string}
                                        onChange={(e) => setInLeagueName(e.target.value)}
                                        placeholder="e.g., Clay"
                                        className="w-full px-3 py-2 bg-background-input border border-border-primary rounded-md text-accent-alt placeholder-text-placeholder focus:outline-none focus:ring-2 focus:ring-accent-primary focus:border-transparent"
                                    />
                                </div>

                                <div>
                                    <label htmlFor="teamName"
                                        className="block text-text-primary text-sm font-semibold mb-2">Your Team Name</label>
                                    <input
                                        type="text"
                                        id="teamName"
                                        value={teamName as string}
                                        onChange={(e) => setTeamName(e.target.value)}
                                        placeholder="e.g., Driftveil City Capybaras"
                                        className="w-full px-3 py-2 border border-border-primary rounded-md text-accent-alt placeholder-text-placeholder focus:outline-none focus:ring-2 focus:ring-accent-primary focus:border-transparent"
                                    />
                                </div>
                            </div>
                            <div className="mt-6 flex justify-end gap-4">
                                <button
                                    type="button"
                                    onClick={handleCloseModal}
                                    disabled={isSubmitting}
                                    className="px-4 py-2 text-white bg-error-500 rounded hover:bg-background-secondary hover:text-error-500 cursor-pointer"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="submit"
                                    disabled={isSubmitting}
                                    className={`px-4 py-2 cursor-pointer rounded text-text-on-accent border-accent-primary bg-accent-primary hover:bg-accent-primary-hover ${isSubmitting ? 'opacity-50 cursor-not-allowed' : ''}`}
                                >
                                    {isSubmitting ? (
                                        <svg className="animate-spin h-5 w-5 text-white inline-block mr-2" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                                        </svg>
                                    ) : (
                                        'Join League'
                                    )}
                                </button>
                            </div>
                        </form>

                    </Modal>
                }
            </div>
        </Layout >
    );
};

export default JoinLeague;
