
interface ModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: React.ReactNode;
    background?: string;
    titleStyle?: string;
    children: React.ReactNode;
    showDefaultCloseButton?: boolean;
}

const Modal: React.FC<ModalProps> = ({ isOpen, onClose, title, children, showDefaultCloseButton = false,
    background = 'bg-background-surface',
    titleStyle = "text-lg font-bold text-text-primary" }) => {
    if (!isOpen) return null;
    return (
        <div
            className="fixed inset-0 overflow-y-auto h-full w-full"
            style={{ backgroundColor: 'rgba(0, 0, 0, 0.5)' }}
            id="my-modal"
        >
            <div className={`relative top-20 mx-auto p-5 border max-w-xl shadow-lg rounded-md ${background}`}>
                <div className="flex justify-between items-center pb-3">
                    <h3 className={`${titleStyle}`}>{title}</h3>
                    <button className="text-gray-400 hover:text-gray-600" onClick={onClose}>
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
                    </button>
                </div>
                <div className="mt-2 mb-4">
                    {children}
                </div>
                {showDefaultCloseButton && (
                    <div className="flex justify-end pt-4">
                        <button className="px-4 py-2 bg-accent-primary text-text-on-accent text-base font-medium rounded-md shadow-sm hover:bg-accent-primary-hover focus:outline-none focus:ring-2 focus:ring-blue-500" onClick={onClose}>Close</button>
                    </div>
                )}
            </div>
        </div >
    );
};

export default Modal;
