import React from 'react';

interface FullScreenModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: string;
    children: React.ReactNode;
}

const FullScreenModal: React.FC<FullScreenModalProps> = ({ isOpen, onClose, title, children }) => {
    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-background-primary overflow-y-auto h-full w-full text-text-primary">
            <div className="relative p-5">
                <div className="flex justify-between items-center pb-3">
                    <h3 className="text-lg font-bold">{title}</h3>
                    <button className="text-gray-400 hover:text-gray-600" onClick={onClose}>
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
                <div className="mt-2 mb-4">
                    {children}
                </div>
            </div>
        </div>
    );
};

export default FullScreenModal;
