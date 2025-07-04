import NavBar from "../componenets/navbar"
import DraftCard from "../componenets/draftCards"
import pokemonData from "../assets/pokemons"
import Filter from "../componenets/filter"

export default function draftboard(){
    const cards = pokemonData.map(pokemon=>{
        return <DraftCard key={pokemon.id}
            name={pokemon.name} 
            pic={pokemon.sprites.front_default}
            type={pokemon.types[0]}
            hp={pokemon.stats.hp}
            ability={pokemon.abilities[0].name}
            attack={pokemon.stats.attack}
            defense={pokemon.stats.defense}
            specialAtk={pokemon.stats["special-attack"]}
            specialDef={pokemon.stats["special-defense"]}
            speed={pokemon.stats.speed}
        
        />
    })
    //for filter: stats, types, cost, name
    return(
        <>
        <div className="h-full bg-[#BFC0C0]" flex flex-row>
            <NavBar page="Draftboard"/>
            
            <div className="flex flex-row">
                <div className="flex flex-col w-[70%]">
                    <div className="flex flex-row m-4 p-8 pb-0 mb-2 justify-between">
                        <div class="max-w-screen-md leading-6"> 
                            <form class="relative mx-auto flex max-w-2xl items-center justify-between rounded-md border shadow-lg"> 
                                <svg class="absolute left-2 block h-3 w-5 text-gray-400" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <circle cx="11" cy="11" r="8" class=""></circle>
                                <line x1="21" y1="21" x2="16.65" y2="16.65" class=""></line>
                                </svg>
                                <input type="name" name="search" class="h-10 w-full rounded-md py-0 pr-30 pl-12 outline-none focus:ring-2" placeholder="Search :" />
                                <button type="submit" class="absolute right-0 mr-1 inline-flex h-8 items-center justify-center rounded-lg bg-gray-900 px-10 font-medium text-white focus:ring-4 hover:bg-gray-700">Search</button>
                            </form>
                        </div>
                        <Filter/>
                    </div>
                    <div className="grid grid-cols-5 gap-4 m-4 p-6 mt-0 pt-0 pr-8 overflow-scroll h-screen rounded-2xl">
                        {cards}
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