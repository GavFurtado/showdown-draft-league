import placeholder from '../assets/placeholder-roster.png'

export default function Roster() {
    return (
        <table className="pokemon-table">
            <tr>
                <div className='first-five'>
                    <td>
                        <img src={placeholder} alt="Pokemon 1" />
                        <span className="name">Pokemon 1</span>
                    </td>
                    <td>
                        <img src={placeholder} alt="Pokemon 2" />
                        <span className="name">Pokemon 2</span>
                    </td>
                    <td>
                        <img src={placeholder} alt="Pokemon 3" />
                        <span className="name">Pokemon 3</span>
                    </td>
                    <td>
                        <img src={placeholder} alt="Pokemon 4" />
                        <span className="name">Pokemon 4</span>
                    </td>
                    <td>
                        <img src={placeholder} alt="Pokemon 5" />
                        <span className="name">Pokemon 5</span>
                    </td>
                </div>
                <div className='second-five'>
                    <td>
                        <img src={placeholder} alt="Pokemon 6" />
                        <span className="name">Pokemon 6</span>
                    </td>
                    <td>
                        <img src={placeholder} alt="Pokemon 7" />
                        <span className="name">Pokemon 7</span>
                    </td>
                    <td>
                        <img src={placeholder} alt="Pokemon 8" />
                        <span className="name">Pokemon 8</span>
                    </td>
                    <td>
                        <img src={placeholder} alt="Pokemon 9" />
                        <span className="name">Pokemon 9</span>
                    </td>
                    <td>
                        <img src={placeholder} alt="Pokemon 10" />
                        <span className="name">Pokemon 10</span>
                    </td>
                </div>
            </tr>
        </table>
    )
}
