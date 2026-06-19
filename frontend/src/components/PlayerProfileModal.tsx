import React, { useState, useEffect } from 'react';
import Modal from './Modal';
import { sanitizeInput, containsHtml, containsForbiddenChars } from '../utils/validationUtils';
import { updatePlayerProfile } from '../api/api';

interface PlayerProfileModalProps {
    isOpen: boolean;
    onClose: () => void;
    leagueId: string;
    playerId: string;
    leagueName: string;
    initialInLeagueName: string;
    initialTeamName: string;
    onSuccess: () => void;
    description?: React.ReactNode;
}

const PlayerProfileModal: React.FC<PlayerProfileModalProps> = ({
    isOpen,
    onClose,
    leagueId,
    playerId,
    leagueName,
    initialInLeagueName,
    initialTeamName,
    onSuccess,
    description
}) => {
    const [formData, setFormData] = useState({
        InLeagueName: initialInLeagueName,
        TeamName: initialTeamName
    });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Sync state when props change (especially useful if initial values arrive late)
    useEffect(() => {
        setFormData({
            InLeagueName: initialInLeagueName,
            TeamName: initialTeamName
        });
    }, [initialInLeagueName, initialTeamName, isOpen]);

    const handleSubmit = async () => {
        if (containsHtml(formData.InLeagueName) || containsHtml(formData.TeamName)) {
            setError("HTML tags (<, >) are not allowed.");
            return;
        }
        if (containsForbiddenChars(formData.InLeagueName) || containsForbiddenChars(formData.TeamName)) {
            setError("The '%' and '\' characters are not allowed.");
            return;
        }

        setLoading(true);
        setError(null);
        try {
            await updatePlayerProfile(leagueId, playerId, {
                InLeagueName: sanitizeInput(formData.InLeagueName),
                TeamName: sanitizeInput(formData.TeamName)
            });
            onSuccess();
            onClose();
        } catch (err: any) {
            console.error("Failed to update player info", err);
            setError(err.response?.data?.error || "Failed to update profile.");
        } finally {
            setLoading(false);
        }
    };

    return (
        <Modal
            isOpen={isOpen}
            onClose={onClose}
            title={<span className="text-xl font-bold text-text-primary">Player Profile</span>}
            showDefaultCloseButton={false}
        >
            <div className="space-y-6 pt-2">
                {description || (
                    <p className="text-text-secondary text-sm">
                        Update your identity for <strong>{leagueName}</strong>.
                    </p>
                )}
                
                {error && (
                    <div className="p-3 bg-red-50 border border-red-200 text-red-700 rounded-lg text-sm">
                        {error}
                    </div>
                )}

                <div>
                    <label className="block text-sm font-bold text-text-primary mb-1">Coach Name</label>
                    <input
                        type="text"
                        value={formData.InLeagueName}
                        onChange={(e) => setFormData({ ...formData, InLeagueName: e.target.value })}
                        className="w-full rounded-xl border-gray-300 bg-gray-50 focus:bg-white focus:border-accent-primary p-3 border shadow-sm"
                        placeholder="Your Name"
                    />
                </div>
                <div>
                    <label className="block text-sm font-bold text-text-primary mb-1">Team Name</label>
                    <input
                        type="text"
                        value={formData.TeamName}
                        onChange={(e) => setFormData({ ...formData, TeamName: e.target.value })}
                        className="w-full rounded-xl border-gray-300 bg-gray-50 focus:bg-white focus:border-accent-primary p-3 border shadow-sm"
                        placeholder="e.g. Pallet Town Pikachus"
                    />
                </div>
                <div className="flex justify-end gap-3 pt-4">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-sm font-bold text-text-secondary hover:text-text-primary transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSubmit}
                        disabled={loading}
                        className="bg-accent-primary text-white px-6 py-2 rounded-xl font-bold shadow-md hover:bg-accent-primary-hover transition-all"
                    >
                        {loading ? 'Saving...' : 'Save Profile'}
                    </button>
                </div>
            </div>
        </Modal>
    );
};

export default PlayerProfileModal;