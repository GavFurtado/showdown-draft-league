import { useState } from 'react'; // Import useState
import Modal from "../components/Modal";

export default function TeamSheets() {
    const [isModalOpen, setIsModalOpen] = useState(false); // Use useState for modal visibility

    const closeModal = () => {
        setIsModalOpen(false);
    };

    const title: string = "My Team Sheet Modal";

    return (
        <>
            <div>
                <button onClick={() => setIsModalOpen(true)}>Open Team Sheet Modal</button> {/* Button to open modal */}
                <Modal isOpen={isModalOpen} onClose={closeModal} title={title}>
                    <p>This is the content for your team sheet.</p>
                    <p>You can display player rosters, stats, etc., here.</p>
                </Modal>
            </div>
        </>
    );
}
