import { Dialog, DialogPanel } from '@headlessui/react';
import React from 'react';

interface FullScreenModalProps {
    isOpen: boolean;
    onClose: () => void;
    children: React.ReactNode;
}

const FullScreenModal: React.FC<FullScreenModalProps> = ({ isOpen, onClose, children }) => {
    return (
        <Dialog as="div" className="relative z-50" open={isOpen} onClose={onClose}>
            <div className="fixed inset-0 bg-black/80 backdrop-blur-sm" />

            <div className="fixed inset-0 overflow-y-auto">
                <div className="flex min-h-full items-center justify-center p-4 text-center">
                    <DialogPanel className="w-full h-full transform transition-all">
                        <div className="relative p-5">
                            <div className="flex justify-end items-center pb-3">
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
                    </DialogPanel>
                </div>
            </div>
        </Dialog>
    );
};

export default FullScreenModal;
