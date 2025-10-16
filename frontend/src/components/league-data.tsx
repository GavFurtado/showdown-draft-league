import { useState } from 'react';
import Score from './score';
import PlayerScedhule from './playerScedhule';
import Roster from './roster'
export default function LeagueDropdown() {
    const [isOpen, setIsOpen] = useState(false);

    const toggleDropdown = () => {
        setIsOpen(!isOpen);
    };

    return (
        <div className="viewLeagueButton">
            <button className="showLeagueButton" onClick={toggleDropdown}> League 1</button>
            <div className={`dropdown-content ${isOpen ? 'show' : ''}`}>
                {/* content to be displayed after dropdown */}
                <h1> Team - Big Booty Boys</h1>
                <PlayerScedhule />
                <Score />
                <Roster />
            </div>
        </div>
    );
};


