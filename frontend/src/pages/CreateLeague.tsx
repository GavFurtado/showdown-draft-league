import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Layout from '../components/Layout';
import Modal from '../components/Modal';
import PlayerProfileModal from '../components/PlayerProfileModal';
import { createLeague, getPlayersByLeague } from '../api/api';
import { LeagueCreateRequest, UpdatePlayerInfoRequest } from '../api/request_interfaces';
import { League, LeagueFormat } from '../api/data_interfaces';
import { useUser } from '../context/UserContext';
import { ChevronRightIcon, ChevronLeftIcon, XMarkIcon } from '@heroicons/react/24/solid';
import { sanitizeInput, containsHtml, containsForbiddenChars } from '../utils/validationUtils';

// Default Initial State
const initialFormat: LeagueFormat = {
    IsSnakeRoundDraft: true,
    DraftOrderType: "RANDOM",
    SeasonType: "ROUND_ROBIN_ONLY",
    GroupCount: 1,
    PlayoffType: "NONE",
    PlayoffParticipantCount: 4,
    PlayoffByesCount: 0,
    PlayoffSeedingType: "STANDARD",
    AllowTransfers: true,
    TransfersCostCredits: true,
    TransferCreditsPerWindow: 2,
    TransferCreditCap: 6,
    TransferWindowFrequencyDays: 7,
    TransferWindowDuration: 48,
    DropCost: 1,
    PickupCost: 1,
};

const initialLeagueState: LeagueCreateRequest = {
    Name: "",
    RulesetDescription: "",
    MaxPokemonPerPlayer: 10,
    MinPokemonPerPlayer: 8,
    StartingDraftPoints: 100,
    Format: initialFormat
};

// Reusable "Card" Component for Selection
interface SelectionCardProps {
    title: string;
    description: string;
    isSelected: boolean;
    isDisabled?: boolean;
    onClick: () => void;
}
const SelectionCard: React.FC<SelectionCardProps> = ({ title, description, isSelected, isDisabled, onClick }) => (
    <div
        onClick={!isDisabled ? onClick : undefined}
        className={`border-2 rounded-xl p-4 transition-all duration-200 ${isDisabled
            ? 'border-gray-100 bg-gray-50 opacity-60 cursor-not-allowed'
            : 'cursor-pointer'
            } ${isSelected && !isDisabled
                ? 'border-accent-primary bg-indigo-50 shadow-md transform scale-[1.02]'
                : !isDisabled ? 'border-gray-200 bg-white hover:border-accent-primary/50 hover:bg-gray-50' : ''
            }`}
    >
        <div className="flex items-center justify-between">
            <h3 className={`font-bold ${isSelected && !isDisabled ? 'text-accent-primary' : 'text-text-primary'}`}>{title}</h3>
            {isSelected && !isDisabled && <div className="h-4 w-4 rounded-full bg-accent-primary" />}
        </div>
        <p className="text-sm text-text-secondary mt-1">{description}</p>
    </div>
);


const CreateLeague: React.FC = () => {
    const navigate = useNavigate();
    const { user } = useUser();
    const [currentStep, setCurrentStep] = useState(1);
    const [formData, setFormData] = useState<LeagueCreateRequest>(initialLeagueState);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Post-creation Modal State
    const [createdLeague, setCreatedLeague] = useState<League | null>(null);
    const [showPlayerModal, setShowPlayerModal] = useState(false);
    const [newPlayerId, setNewPlayerId] = useState<string>('');
    const [initialPlayerData, setInitialPlayerData] = useState({
        InLeagueName: "",
        TeamName: ""
    });

    // Enforce constraints between fields
    React.useEffect(() => {
        if (formData.Format.PlayoffByesCount >= formData.Format.PlayoffParticipantCount) {
            updateFormat("PlayoffByesCount", Math.max(0, formData.Format.PlayoffParticipantCount - 1));
        }
    }, [formData.Format.PlayoffParticipantCount]);

    React.useEffect(() => {
        if (formData.MinPokemonPerPlayer > formData.MaxPokemonPerPlayer && formData.MinPokemonPerPlayer !== 0) {
            setFormData(prev => ({ ...prev, MinPokemonPerPlayer: prev.MaxPokemonPerPlayer }));
        }
    }, [formData.MaxPokemonPerPlayer]);

    React.useEffect(() => {
        if (formData.Format.PlayoffType === "SINGLE_ELIM" && formData.Format.PlayoffSeedingType === "FULLY_SEEDED") {
            updateFormat("PlayoffSeedingType", "STANDARD");
        }
    }, [formData.Format.PlayoffType]);

    const updateFormat = (field: keyof LeagueFormat, value: any) => {
        setFormData(prev => ({
            ...prev,
            Format: {
                ...prev.Format,
                [field]: value
            }
        }));
    };

    const updateField = (field: keyof LeagueCreateRequest, value: any) => {
        setFormData(prev => ({
            ...prev,
            [field]: value
        }));
    };

    const handleNext = () => {
        setError(null);

        // Validation & Sanitization per step
        if (currentStep === 1) {
            if (!formData.Name) {
                setError("League Name is required.");
                return;
            }
            if (containsHtml(formData.Name) || containsHtml(formData.RulesetDescription)) {
                setError("HTML tags (<, >) are not allowed in the name or description.");
                return;
            }
            if (containsForbiddenChars(formData.Name) || containsForbiddenChars(formData.RulesetDescription)) {
                setError("The '%' and '\\' characters are not allowed.");
                return;
            }

            // Final sanitize (as a fallback)
            const cleanName = sanitizeInput(formData.Name);
            const cleanRules = sanitizeInput(formData.RulesetDescription);
            if (cleanName !== formData.Name || cleanRules !== formData.RulesetDescription) {
                setFormData(prev => ({ ...prev, Name: cleanName, RulesetDescription: cleanRules }));
            }
        }

        if (currentStep === 2) {
            if (formData.MinPokemonPerPlayer > 0 && formData.MinPokemonPerPlayer > formData.MaxPokemonPerPlayer) {
                setError("Minimum roster size cannot exceed maximum roster size.");
                return;
            }
        }

        // Logic for skipping Step 4 (Playoffs) if Round Robin Only is selected
        if (currentStep === 3) {
            if (formData.Format.SeasonType === "ROUND_ROBIN_ONLY") {
                setCurrentStep(5);
                return;
            }
        }

        if (currentStep === 4) {
            if (formData.Format.PlayoffType !== "NONE") {
                if (formData.Format.PlayoffByesCount >= formData.Format.PlayoffParticipantCount) {
                    setError("Byes cannot be greater than or equal to the number of participants.");
                    return;
                }
            }
        }

        if (currentStep === 5) {
            if (formData.Format.AllowTransfers) {
                if (formData.Format.TransferWindowFrequencyDays % 7 !== 0 || formData.Format.TransferWindowFrequencyDays === 0) {
                    setError("Transfer window frequency must be a multiple of 7 (weekly intervals).");
                    return;
                }
            }
        }

        setCurrentStep(prev => prev + 1);
    };

    const handleBack = () => {
        setError(null);
        // Logic for skipping Step 4 (Playoffs) if Round Robin Only is selected
        if (currentStep === 5 && formData.Format.SeasonType === "ROUND_ROBIN_ONLY") {
            setCurrentStep(3);
            return;
        }
        setCurrentStep(prev => prev - 1);
    };

    const handleSubmit = async () => {
        if (!user) return;
        setLoading(true);
        setError(null);

        try {
            const payload = {
                ...formData,
                Name: sanitizeInput(formData.Name),
                RulesetDescription: sanitizeInput(formData.RulesetDescription),
                StartDate: new Date().toISOString()
            };

            const response = await createLeague(payload);
            const newLeague = response.data as League;
            setCreatedLeague(newLeague);

            // Fetch the automatically created player profile
            const playersResponse = await getPlayersByLeague(newLeague.ID);
            const players = playersResponse.data as any[];
            const myPlayer = players.find((p: any) => p.UserID === user.ID);

            if (myPlayer) {
                setNewPlayerId(myPlayer.ID);
                setInitialPlayerData({
                    InLeagueName: user.ShowdownUsername || user.DiscordUsername,
                    TeamName: `${user.DiscordUsername}'s Team`
                });
                setShowPlayerModal(true);
            } else {
                // Fallback if player not found immediately
                navigate(`/league/${newLeague.ID}/dashboard`);
            }
        } catch (err: any) {
            console.error("Failed to create league", err);
            setError(err.response?.data?.error || "Failed to create league. Please try again.");
        } finally {
            setLoading(false);
        }
    };

    // --- Render Steps ---
    // (Rest of the render functions remain same...)


    // --- Render Steps ---

    const renderStep1 = () => (
        <div className="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
            <h2 className="text-2xl font-bold text-text-primary">Let's get started.</h2>
            <div className="bg-yellow-50 border border-yellow-200 rounded-xl p-4 flex items-start">
                <div className="shrink-0">
                    <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                        <path fillRule="evenodd" d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 5zm0 9a1 1 0 100-2 1 1 0 000 2z" clipRule="evenodd" />
                    </svg>
                </div>
                <div className="ml-3">
                    <h3 className="text-sm font-medium text-yellow-800">Limit: 2 Leagues</h3>
                    <div className="mt-1 text-sm text-yellow-700">
                        <p>You can currently manage up to 2 leagues per user.</p>
                    </div>
                </div>
            </div>
            <p className="text-text-secondary">Give your league a name and describe the rules.</p>

            <div className="space-y-4">
                <div>
                    <label className="block text-sm font-bold text-text-primary mb-1">League Name</label>
                    <input
                        type="text"
                        value={formData.Name}
                        onChange={(e) => updateField("Name", e.target.value)}
                        className="w-full rounded-xl border-gray-300 bg-gray-50 focus:bg-white focus:border-accent-primary focus:ring-accent-primary transition-colors p-3 border shadow-sm"
                        placeholder="e.g. Indigo Plateau Season 1"
                    />
                </div>
                <div>
                    <label className="block text-sm font-bold text-text-primary mb-1">Ruleset & Description</label>
                    <textarea
                        rows={6}
                        value={formData.RulesetDescription}
                        onChange={(e) => updateField("RulesetDescription", e.target.value)}
                        className="w-full rounded-xl border-gray-300 bg-gray-50 focus:bg-white focus:border-accent-primary focus:ring-accent-primary transition-colors p-3 border shadow-sm"
                        placeholder="Describe tiers, battle rules, scheduling policies..."
                    />
                </div>
            </div>
        </div>
    );

    const renderStep2 = () => (
        <div className="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
            <h2 className="text-2xl font-bold text-text-primary">Draft Configuration</h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <SelectionCard
                    title="Snake Draft"
                    description="Order reverses every round (1-10, 10-1)."
                    isSelected={formData.Format.IsSnakeRoundDraft}
                    onClick={() => updateFormat("IsSnakeRoundDraft", true)}
                />
                <SelectionCard
                    title="Linear Draft"
                    description="Order stays the same every round (1-10, 1-10)."
                    isSelected={!formData.Format.IsSnakeRoundDraft}
                    onClick={() => updateFormat("IsSnakeRoundDraft", false)}
                />
            </div>

            <div className="pt-4 space-y-4">
                <h3 className="font-bold text-text-primary">Draft Order</h3>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <SelectionCard
                        title="Randomized"
                        description="System generates order."
                        isSelected={formData.Format.DraftOrderType === "RANDOM"}
                        onClick={() => updateFormat("DraftOrderType", "RANDOM")}
                    />
                    <SelectionCard
                        title="Manual"
                        description="Admin sets order manually."
                        isSelected={formData.Format.DraftOrderType === "MANUAL"}
                        onClick={() => updateFormat("DraftOrderType", "MANUAL")}
                    />
                    <SelectionCard
                        title="Pending"
                        description="Decide later."
                        isSelected={formData.Format.DraftOrderType === "PENDING"}
                        onClick={() => updateFormat("DraftOrderType", "PENDING")}
                    />
                </div>
            </div>

            <div className="pt-4 space-y-4">
                <h3 className="font-bold text-text-primary">Roster Size & Points</h3>
                <div className="grid grid-cols-3 gap-4">
                    <div>
                        <label className="block text-xs font-bold text-text-secondary uppercase mb-1">Min. Pokémon</label>
                        <input
                            type="number"
                            min={0}
                            max={formData.MaxPokemonPerPlayer}
                            value={formData.MinPokemonPerPlayer}
                            onChange={(e) => updateField("MinPokemonPerPlayer", parseInt(e.target.value))}
                            className="w-full rounded-xl border-gray-300 bg-gray-50 p-3 border shadow-sm"
                        />
                        <p className="text-[10px] text-text-secondary mt-1">Set to 0 to disable minimum restriction.</p>
                    </div>
                    <div>
                        <label className="block text-xs font-bold text-text-secondary uppercase mb-1">Max. Pokémon</label>
                        <input
                            type="number"
                            min={1}
                            max={18}
                            value={formData.MaxPokemonPerPlayer}
                            onChange={(e) => updateField("MaxPokemonPerPlayer", parseInt(e.target.value))}
                            className="w-full rounded-xl border-gray-300 bg-gray-50 p-3 border shadow-sm"
                        />
                    </div>
                    <div>
                        <label className="block text-xs font-bold text-text-secondary uppercase mb-1">Draft Points</label>
                        <input
                            type="number"
                            min={0}
                            value={formData.StartingDraftPoints}
                            onChange={(e) => updateField("StartingDraftPoints", parseInt(e.target.value))}
                            className="w-full rounded-xl border-gray-300 bg-gray-50 p-3 border shadow-sm"
                        />
                    </div>
                </div>
            </div>
        </div>
    );

    const renderStep3 = () => (
        <div className="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
            <h2 className="text-2xl font-bold text-text-primary">Season Format</h2>

            <div className="grid grid-cols-1 gap-4">
                <SelectionCard
                    title="Round Robin Only"
                    description="Classic league format. Everyone plays everyone."
                    isSelected={formData.Format.SeasonType === "ROUND_ROBIN_ONLY"}
                    onClick={() => {
                        setFormData(prev => ({
                            ...prev,
                            Format: {
                                ...prev.Format,
                                SeasonType: "ROUND_ROBIN_ONLY",
                                PlayoffType: "NONE"
                            }
                        }));
                    }}
                />
                <SelectionCard
                    title="Playoffs Only"
                    description="Skip straight to a bracket tournament. (Coming Soon)"
                    isSelected={formData.Format.SeasonType === "PLAYOFFS_ONLY"}
                    isDisabled={true}
                    onClick={() => { }} // Disabled
                />
                <SelectionCard
                    title="Hybrid (Round Robin + Playoffs)"
                    description="Regular season followed by a top-cut bracket."
                    isSelected={formData.Format.SeasonType === "HYBRID"}
                    onClick={() => {
                        setFormData(prev => ({
                            ...prev,
                            Format: {
                                ...prev.Format,
                                SeasonType: "HYBRID",
                                PlayoffType: "SINGLE_ELIM"
                            }
                        }));
                    }}
                />
            </div>

            <div className="pt-4 space-y-4">
                <h3 className="font-bold text-text-primary">Divisions (Groups)</h3>
                <div className="grid grid-cols-2 gap-4">
                    <SelectionCard
                        title="Single Group"
                        description="All players in one division."
                        isSelected={formData.Format.GroupCount === 1}
                        onClick={() => updateFormat("GroupCount", 1)}
                    />
                    <SelectionCard
                        title="Two Groups"
                        description="Split players into two divisions."
                        isSelected={formData.Format.GroupCount === 2}
                        onClick={() => updateFormat("GroupCount", 2)}
                    />
                </div>
            </div>
        </div>
    );

    const renderStep4 = () => (
        <div className="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
            <h2 className="text-2xl font-bold text-text-primary">Playoff Structure</h2>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <SelectionCard
                    title="No Playoffs"
                    description="Season ends after round robin."
                    isSelected={formData.Format.PlayoffType === "NONE"}
                    onClick={() => updateFormat("PlayoffType", "NONE")}
                />
                <SelectionCard
                    title="Single Elim"
                    description="One loss and you're out."
                    isSelected={formData.Format.PlayoffType === "SINGLE_ELIM"}
                    onClick={() => updateFormat("PlayoffType", "SINGLE_ELIM")}
                />
                <SelectionCard
                    title="Double Elim"
                    description="Losers bracket opportunity."
                    isSelected={formData.Format.PlayoffType === "DOUBLE_ELIM"}
                    onClick={() => updateFormat("PlayoffType", "DOUBLE_ELIM")}
                />
            </div>

            {formData.Format.PlayoffType !== "NONE" && (
                <div className="space-y-4 pt-4 border-t border-gray-100">
                    <h3 className="font-bold text-text-primary">Seeding</h3>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                        <SelectionCard
                            title="Standard"
                            description="No byes, all players start in Round 1."
                            isSelected={formData.Format.PlayoffSeedingType === "STANDARD"}
                            onClick={() => updateFormat("PlayoffSeedingType", "STANDARD")}
                        />
                        <SelectionCard
                            title="Byes Only"
                            description="Top seeds earn a first-round bye."
                            isSelected={formData.Format.PlayoffSeedingType === "BYES_ONLY"}
                            onClick={() => updateFormat("PlayoffSeedingType", "BYES_ONLY")}
                        />
                        {/* Seeded option is not allowed for Single Elim, but allowed for Double Elim */}
                        {formData.Format.PlayoffType !== "SINGLE_ELIM" && (
                            <SelectionCard
                                title="Fully Seeded"
                                description="Top: Byes. Mid: Upper R1. Low: Lower Bracket."
                                isSelected={formData.Format.PlayoffSeedingType === "FULLY_SEEDED"}
                                onClick={() => updateFormat("PlayoffSeedingType", "FULLY_SEEDED")}
                            />
                        )}
                    </div>

                    <div className="grid grid-cols-2 gap-6">
                        <div>
                            <label className="block text-sm font-bold text-text-primary mb-1">Participants</label>
                            <input
                                type="number"
                                min={2}
                                value={formData.Format.PlayoffParticipantCount}
                                onChange={(e) => updateFormat("PlayoffParticipantCount", parseInt(e.target.value))}
                                className="w-full rounded-xl border-gray-300 bg-gray-50 p-3 border shadow-sm"
                            />
                        </div>
                        <div>
                            <label className="block text-sm font-bold text-text-primary mb-1">Byes</label>
                            <input
                                type="number"
                                min={0}
                                max={formData.Format.PlayoffParticipantCount - 1}
                                value={formData.Format.PlayoffSeedingType === "STANDARD" ? 0 : formData.Format.PlayoffByesCount}
                                disabled={formData.Format.PlayoffSeedingType === "STANDARD"}
                                onChange={(e) => updateFormat("PlayoffByesCount", parseInt(e.target.value))}
                                className={`w-full rounded-xl border-gray-300 p-3 border shadow-sm ${formData.Format.PlayoffSeedingType === "STANDARD" ? 'bg-gray-100 text-gray-400 opacity-50' : 'bg-gray-50'}`}
                            />
                            {formData.Format.PlayoffSeedingType === "STANDARD" && (
                                <p className="text-[10px] text-text-secondary mt-1 italic">Standard seeding uses no byes.</p>
                            )}
                        </div>
                    </div>

                    <div className="mt-4 p-3 bg-blue-50 border border-blue-100 rounded-xl text-[11px] text-blue-700">
                        <p className="font-bold mb-1 flex items-center gap-1">
                            <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20"><path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd"></path></svg>
                            Configuration Note
                        </p>
                        Some playoff configurations (like too many byes for the participant count) may be invalid depending on the bracket size. These can be adjusted later in the league settings.
                    </div>
                </div>
            )}
        </div>
    );

    const renderStep5 = () => (
        <div className="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
            <h2 className="text-2xl font-bold text-text-primary">Transfer Market</h2>

            <SelectionCard
                title="Enable Transfers"
                description="Allow coaches to drop/add Pokemon during the season."
                isSelected={formData.Format.AllowTransfers}
                onClick={() => updateFormat("AllowTransfers", !formData.Format.AllowTransfers)}
            />

            {formData.Format.AllowTransfers && (
                <div className="pt-4 space-y-6 animate-in fade-in slide-in-from-top-2">
                    <div className="grid grid-cols-2 gap-6">
                        <div>
                            <label className="block text-sm font-bold text-text-primary mb-1">Window Frequency (Days)</label>
                            <input
                                type="number"
                                min={7}
                                step={7}
                                value={formData.Format.TransferWindowFrequencyDays}
                                onChange={(e) => updateFormat("TransferWindowFrequencyDays", parseInt(e.target.value))}
                                className="w-full rounded-xl border-gray-300 bg-gray-50 p-3 border shadow-sm"
                            />
                            <p className="text-[10px] text-text-secondary mt-1 italic">Must be a multiple of 7 (e.g. 7, 14, 21).</p>
                        </div>
                        <div>
                            <label className="block text-sm font-bold text-text-primary mb-1">Window Duration (Hours)</label>
                            <input
                                type="number"
                                min={1}
                                value={formData.Format.TransferWindowDuration}
                                onChange={(e) => updateFormat("TransferWindowDuration", parseInt(e.target.value))}
                                className="w-full rounded-xl border-gray-300 bg-gray-50 p-3 border shadow-sm"
                            />
                            <p className="text-[10px] text-text-secondary mt-1 italic">How long the transfer window stays open.</p>
                        </div>
                    </div>

                    <SelectionCard
                        title="Transfers Cost Credits"
                        description="Limit activity with a credit system."
                        isSelected={formData.Format.TransfersCostCredits}
                        onClick={() => updateFormat("TransfersCostCredits", !formData.Format.TransfersCostCredits)}
                    />

                    {formData.Format.TransfersCostCredits && (
                        <div className="animate-in fade-in slide-in-from-top-2 space-y-6">
                            <div className="grid grid-cols-2 gap-6">
                                <div>
                                    <label className="block text-sm font-bold text-text-primary mb-1">Credits / Window</label>
                                    <input
                                        type="number"
                                        min={0}
                                        value={formData.Format.TransferCreditsPerWindow}
                                        onChange={(e) => updateFormat("TransferCreditsPerWindow", parseInt(e.target.value))}
                                        className="w-full rounded-xl border-gray-300 bg-gray-50 p-3 border shadow-sm"
                                    />
                                    <p className="text-[10px] text-text-secondary mt-1 italic">Credits granted at the start of each window.</p>
                                </div>
                                <div>
                                    <label className="block text-sm font-bold text-text-primary mb-1">Max. Credit Cap</label>
                                    <input
                                        type="number"
                                        min={0}
                                        value={formData.Format.TransferCreditCap}
                                        onChange={(e) => updateFormat("TransferCreditCap", parseInt(e.target.value))}
                                        className="w-full rounded-xl border-gray-300 bg-gray-50 p-3 border shadow-sm"
                                    />
                                    <p className="text-[10px] text-text-secondary mt-1 italic">Maximum credits a coach can hold.</p>
                                </div>
                            </div>

                            <div className="grid grid-cols-2 gap-6">
                                <div>
                                    <label className="block text-sm font-bold text-text-primary mb-1">Drop Cost (Points)</label>
                                    <input
                                        type="number"
                                        min={0}
                                        value={formData.Format.DropCost}
                                        onChange={(e) => updateFormat("DropCost", parseInt(e.target.value))}
                                        className="w-full rounded-xl border-gray-300 bg-gray-50 p-3 border shadow-sm"
                                    />
                                    <p className="text-[10px] text-text-secondary mt-1 italic">Draft points lost when dropping a Pokemon.</p>
                                </div>
                                <div>
                                    <label className="block text-sm font-bold text-text-primary mb-1">Pickup Cost (Points)</label>
                                    <input
                                        type="number"
                                        min={0}
                                        value={formData.Format.PickupCost}
                                        onChange={(e) => updateFormat("PickupCost", parseInt(e.target.value))}
                                        className="w-full rounded-xl border-gray-300 bg-gray-50 p-3 border shadow-sm"
                                    />
                                    <p className="text-[10px] text-text-secondary mt-1 italic">Draft points lost when picking up a Pokemon.</p>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            )}
        </div>
    );

    const renderReview = () => (
        <div className="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300">
            <h2 className="text-2xl font-bold text-text-primary">Ready to Create?</h2>
            <p className="text-text-secondary">Double check your settings. Some cannot be changed once the draft begins.</p>

            <div className="bg-indigo-50 p-6 rounded-xl border border-indigo-100 space-y-3 shadow-sm">
                <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                        <span className="block text-text-secondary">League Name</span>
                        <span className="font-bold text-text-primary text-lg">{formData.Name}</span>
                    </div>
                    <div>
                        <span className="block text-text-secondary">Draft Style</span>
                        <span className="font-bold text-text-primary">{formData.Format.IsSnakeRoundDraft ? "Snake" : "Linear"} ({formData.Format.DraftOrderType})</span>
                    </div>
                    <div>
                        <span className="block text-text-secondary">Roster Size</span>
                        <span className="font-bold text-text-primary">
                            {formData.MinPokemonPerPlayer > 0 ? formData.MinPokemonPerPlayer : 'No Min'} - {formData.MaxPokemonPerPlayer} Pkmn
                        </span>
                    </div>
                    <div>
                        <span className="block text-text-secondary">Season</span>
                        <span className="font-bold text-text-primary">{formData.Format.SeasonType.replace(/_/g, ' ')}</span>
                    </div>
                    <div>
                        <span className="block text-text-secondary">Playoffs</span>
                        <span className="font-bold text-text-primary">{formData.Format.PlayoffType.replace(/_/g, ' ')}</span>
                    </div>
                    <div>
                        <span className="block text-text-secondary">Transfers</span>
                        <span className="font-bold text-text-primary">{formData.Format.AllowTransfers ? "Enabled" : "Disabled"}</span>
                    </div>
                </div>
            </div>
        </div>
    );

    const steps = 6;
    const progress = (currentStep / steps) * 100;

    return (
        <Layout variant="container">
            <div className="max-w-2xl mx-auto py-10">
                {/* Header & Progress */}
                <div className="mb-10">
                    <div className="flex justify-between items-end mb-2">
                        <span className="text-accent-primary font-bold tracking-wide text-sm uppercase">Draft League Creator</span>
                        <span className="text-gray-400 font-medium text-sm">Step {currentStep} of {steps}</span>
                    </div>
                    <div className="h-2 w-full bg-gray-200 rounded-full overflow-hidden">
                        <div
                            className="h-full bg-accent-primary transition-all duration-500 ease-out rounded-full"
                            style={{ width: `${progress}%` }}
                        />
                    </div>
                </div>

                {/* Main Card */}
                <div className="bg-white shadow-xl shadow-indigo-100/50 rounded-2xl p-8 min-h-[500px] border border-gray-100 relative flex flex-col justify-between">
                    {error && (
                        <div className="mb-6 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg flex items-center gap-2" role="alert">
                            <span className="font-bold">Error:</span>
                            <span>{error}</span>
                        </div>
                    )}

                    <div className="grow">
                        {currentStep === 1 && renderStep1()}
                        {currentStep === 2 && renderStep2()}
                        {currentStep === 3 && renderStep3()}
                        {currentStep === 4 && renderStep4()}
                        {currentStep === 5 && renderStep5()}
                        {currentStep === 6 && renderReview()}
                    </div>

                    {/* Navigation */}
                    <div className="mt-10 flex justify-between pt-6 border-t border-gray-100">
                        <button
                            onClick={currentStep === 1 ? () => navigate('/my-leagues') : handleBack}
                            disabled={loading}
                            className="px-6 py-3 rounded-xl font-bold text-text-secondary hover:bg-gray-50 hover:text-text-primary transition-colors flex items-center gap-2"
                        >
                            {currentStep === 1 ? <XMarkIcon className="h-5 w-5" /> : <ChevronLeftIcon className="h-5 w-5" />}
                            {currentStep === 1 ? 'Cancel' : 'Back'}
                        </button>

                        {currentStep < 6 ? (
                            <button
                                onClick={handleNext}
                                className="bg-accent-primary text-white px-8 py-3 rounded-xl font-bold shadow-lg shadow-indigo-500/30 hover:bg-accent-primary-hover hover:shadow-indigo-500/40 hover:-translate-y-0.5 transition-all flex items-center gap-2"
                            >
                                Continue <ChevronRightIcon className="h-5 w-5" />
                            </button>
                        ) : (
                            <button
                                onClick={handleSubmit}
                                disabled={loading}
                                className="bg-green-500 text-white px-8 py-3 rounded-xl font-bold shadow-lg shadow-green-500/30 hover:bg-green-600 hover:shadow-green-500/40 hover:-translate-y-0.5 transition-all"
                            >
                                {loading ? 'Creating...' : 'Create League'}
                            </button>
                        )}
                    </div>
                </div>
            </div>

            {/* Player Setup Modal */}
            {createdLeague && (
                <PlayerProfileModal
                    isOpen={showPlayerModal}
                    onClose={() => navigate(`/league/${createdLeague.ID}/dashboard`)}
                    leagueId={createdLeague.ID}
                    playerId={newPlayerId}
                    leagueName={createdLeague.Name}
                    initialInLeagueName={initialPlayerData.InLeagueName}
                    initialTeamName={initialPlayerData.TeamName}
                    onSuccess={() => navigate(`/league/${createdLeague.ID}/dashboard`)}
                    description={
                        <p className="text-text-secondary text-sm">
                            League created successfully! Now, let's set up your identity for this league.
                        </p>
                    }
                />
            )}

        </Layout>
    );
};

export default CreateLeague;
