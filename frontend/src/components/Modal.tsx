import React from 'react';

interface ModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: string;
    children: React.ReactNode;
}

const Modal: React.FC<ModalProps> = ({ isOpen, onClose, title, children }) => {
    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full" id="my-modal">
            <div className="relative top-20 mx-auto p-5 border max-w-xl shadow-lg rounded-md bg-white">
                <div className="flex justify-between items-center pb-3">
                    <h3 className="text-lg font-bold">{title}</h3>
                    <button className="text-gray-400 hover:text-gray-600" onClick={onClose}>&times;</button>
                </div>
                <div className="mt-2 mb-4">
                    {children}
                </div>
                <div className="flex justify-end pt-4">
                    <button className="px-4 py-2 bg-blue-500 text-white text-base font-medium rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500" onClick={onClose}>Close</button>
                </div>
            </div>
        </div>
    );
};

export default Modal;
