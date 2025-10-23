import React from 'react';

interface ToggleSwitchProps {
    isOn: boolean;
    onToggle: () => void;
    label?: string;
}

const ToggleSwitch: React.FC<ToggleSwitchProps> = ({ isOn, onToggle, label }) => {
    return (
        <div className="flex items-center space-x-2">
            {label && <span className="text-text-primary text-sm font-medium">{label} </span>}
            <button
                onClick={onToggle}
                className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-accent-primary
                    ${isOn ? 'bg-accent-primary' : 'bg-gray-200'}`}
            >
                <span
                    className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform duration-200 ease-in-out
                        ${isOn ? 'translate-x-6' : 'translate-x-1'}`}
                />
            </button>
        </div>
    );
};

export default ToggleSwitch;
