import NavBar from "../componenets/navbar"
import DraftCard from "../componenets/draftCards"
import pokemonData from "../assets/pokemons.json"
import Filter from "../componenets/filter"
import {useState, useEffect} from "react"

const defaultFilters = {
        selectedTypes: [],
        selectedCost: '',
        sortByStat: '',
        sortOrder: 'asc',
    };

export default function Draftboard(){
    const [cards,setCards] = useState(pokemonData);
    const [filters, setFilters] = useState(defaultFilters);
    const [searchTerm, setSearchTerm] = useState('');

    function applyFilter(){
        let update = [...pokemonData]

        if (searchTerm.trim() !== '') {
            update = update.filter(card =>
                card.name.toLowerCase().includes(searchTerm.toLowerCase())
            );
        }

        if(filters.selectedTypes.length>0){
            update=update.filter(card=>
                filters.selectedTypes.some(type =>
                card.types.includes(type)
                )
            );
        }
        
        if(filters.selectedCost){
            update=update.filter(card=>card.cost === filters.selectedCost)
        }
        if (filters.sortByStat) {
            update=update.sort((a, b) => {
                const statA = a.stats[filters.sortByStat];
                const statB = b.stats[filters.sortByStat];

                return filters.sortOrder === 'asc' ? statA - statB : statB - statA;
            });
        }

        setCards(update) 
    }
    function updateFilter(key,value){
        setFilters(prev=>({...prev, [key]:value}))           
    }
    const resetAllFilters = () => {
        setFilters(defaultFilters)
    };
    useEffect(() => {
        applyFilter();
    }, [filters]);
    useEffect(() => {
        applyFilter();
    }, [filters, searchTerm]);
    const cardsToDisplay = cards.map(pokemon=>{
        const name = pokemon.name.charAt(0).toUpperCase() + pokemon.name.slice(1)
        return <DraftCard key={pokemon.id}
            name={name} 
            pic={pokemon.sprites.front_default}
            type={pokemon.types}
            hp={pokemon.stats.hp}
            ability={pokemon.abilities}
            attack={pokemon.stats.attack}
            defense={pokemon.stats.defense}
            specialAtk={pokemon.stats["special-attack"]}
            specialDef={pokemon.stats["special-defense"]}
            speed={pokemon.stats.speed}
            cost={pokemon.cost || 10}
        
        />
    })
    return(
        <>
        <div className="min-h-screen bg-[#BFC0C0] ">
            <NavBar page="Draftboard"/>
            
            <div className="flex flex-row">
                <div className="flex flex-col w-[70%]">
                    <div className="flex flex-row m-4 p-8 pb-0 mb-2 justify-between">
                        {/* search button */}
                        <div className="relative flex text-black">
                            <input
                                type="search"
                                className="placeholder:text-black relative m-0 block flex-auto rounded border border-solid border-black bg-transparent bg-clip-padding px-3 py-[0.25rem] text-base font-normal leading-[1.6] text-surface outline-none transition duration-200 ease-in-out focus:z-[3] focus:border-primary focus:shadow-inset focus:outline-none motion-reduce:transition-none  dark:autofill:shadow-autofill dark:focus:border-primary"
                                placeholder="Search"
                                aria-label="Search"
                                id="exampleFormControlInput2"
                                aria-describedby="button-addon2" 
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                                />
                            <span
                                className="flex items-center whitespace-nowrap px-3 py-[0.25rem] text-surface dark:border-neutral-400 dark:text-white [&>svg]:h-5 [&>svg]:w-5"
                                id="button-addon2">
                                <svg
                                xmlns="http://www.w3.org/2000/svg"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke-width="2"
                                stroke="black">
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
                                </svg>
                            </span>
                        </div>
                        <Filter updateFilter={updateFilter} filters={filters} resetAllFilters={resetAllFilters} />
                    </div>
                    <div className="grid grid-cols-5 gap-4 m-4 p-6 mt-0 pt-0 pr-8 overflow-scroll h-auto max-h-screen rounded-2xl">
                        {cardsToDisplay}
                    </div>
                </div>
            
                <div className="bg-white shadow-md rounded-md overflow-hidden w-[25%] mx-auto mt-16 ml-2 h-[100%]">
                    <div className="bg-gray-100 py-2 px-4">
                        <h2 className="text-l font-semibold text-gray-800">Your Team</h2>
                    </div>
                    <ul className="divide-y divide-gray-200">
                        <li className="flex items-center py-4 px-6">
                            
                            <img className="w-12 h-12 object-cover mr-4" src="https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/9.png" alt="User avatar"></img>
                            <div className="flex-1">
                                <h3 className="text-lg font-medium text-gray-800">BIG MAN BLASTOISE</h3>
                                {/* <p className="text-gray-600 text-base">1234 points</p> */}
                            </div>
                        </li>   
                    </ul>
                </div>
            </div>
        </div>
        </>
    )
}