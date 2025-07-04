import placeholder from '../assets/placeholder-roster.png'

export default function Roster(){
    return(
        <table class="pokemon-table">
            <tr>
                <div className='first-five'>
                <td>
                    <img src={placeholder}alt="Pokemon 1"/>
                    <span class="name">Pokemon 1</span>
                </td>
                <td>
                    <img src={placeholder}alt="Pokemon 2"/>
                    <span class="name">Pokemon 2</span>
                </td>
                <td>
                    <img src={placeholder}alt="Pokemon 3"/>
                    <span class="name">Pokemon 3</span>
                </td>
                <td>
                    <img src={placeholder}alt="Pokemon 4"/>
                    <span class="name">Pokemon 4</span>
                </td>
                <td>
                    <img src={placeholder}alt="Pokemon 5"/>
                    <span class="name">Pokemon 5</span>
                </td>
                </div>
                <div className='second-five'>
                <td>
                    <img src={placeholder}alt="Pokemon 6"/>
                    <span class="name">Pokemon 6</span>
                </td>
                <td>
                    <img src={placeholder}alt="Pokemon 7"/>
                    <span class="name">Pokemon 7</span>
                </td>
                <td>
                    <img src={placeholder}alt="Pokemon 8"/>
                    <span class="name">Pokemon 8</span>
                </td>
                <td>
                    <img src={placeholder}alt="Pokemon 9"/>
                    <span class="name">Pokemon 9</span>
                </td>
                <td>
                    <img src={placeholder}alt="Pokemon 10"/>
                    <span class="name">Pokemon 10</span>
                </td>
                </div>
            </tr>
        </table>
    )
}