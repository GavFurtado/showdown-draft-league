import React, { useState } from 'react';

export default function PokemonCard(pokemon) {
    const [isFlipped, setIsFlipped] = useState(false); 
    const handleFlip = () => {
        setIsFlipped(!isFlipped);
    };
    const handleImageError = (e) => {
        e.target.onerror = null;
        e.target.src = `https://placehold.co/150x150/cccccc/333333?text=No+Image`;
    };
    const types = pokemon.type.map(t => {
        if (typeof t === 'string' && t.length > 0) {
            return t.charAt(0).toUpperCase() + t.slice(1);
        }
        return t;
    }).join(', ');
    const abilities=pokemon.ability.map(a => a.name).join(', ')
    return (
        // Container of the whole thing : Sets perspective for effect and provides a clickable area
        <div
        className=" group h-70 w-47 rounded-lg shadow-lg relative cursor-pointer [perspective:1000px]" // Fixed size for the card
        onClick={handleFlip}
        >
        {/* Inner container: This is the part that actually flips */}
        <div
            className={`
            relative
            w-full h-full
            transition-transform duration-700 ease-in-out
            [transform-style:preserve-3d]
            ${isFlipped ? '[transform:rotateY(180deg)]' : ''}
            `}
        >
            {/* Front Face of the Card */}
            <div className="absolute inset-0 bg-white rounded-lg p-4 flex flex-col items-center justify-center [backface-visibility:hidden]">
                {/* Pokémon Image */}
                <div className="relative w-full h-[100%]">
                    <img
                        src={pokemon.pic}
                        alt={pokemon.name}
                        onError={handleImageError}
                        className="w-[100%] h-[100%] object-contain mb-4 bg-gray-100 p-2"
                    />
                    {/* Pokémon Cost */}
                    <p className="text-lg font-semibold absolute  bottom-2 right-2 ">
                            {pokemon.cost}
                    </p>
                </div>
                    <div className='flex w-[100%] justify-between'>
                        {/* Pokémon Name */}
                        <div>
                            <h3 className="text-lg pb-0 mb-0 font-bold text-gray-800 text-left">
                                {pokemon.name}
                            </h3>
                            <p className='p-0 m-0 text-left text-sm text-gray-600'>{types}</p>
                        </div>
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                            }}
                            className="relative flex items-center align-center justify-center mt-4 h-7.5 w-7.5 rounded-full p-0 hover:border-2 hover: transition-colors">
                            <img
                                src="https://upload.wikimedia.org/wikipedia/commons/5/53/Poké_Ball_icon.svg"
                                className='h-7.5 w-7.5'
                            />
                        </button>
                    </div>
                    
            </div>

            {/* Back Face of the Card */}
            <div className="absolute inset-0 bg-gray-700 text-white rounded-lg p-4 flex flex-col [backface-visibility:hidden] [transform:rotateY(180deg)]">
                <h3 className="text-l font-bold mb-4 text-center">
                    {pokemon.name}
                </h3>
                <ul className="text-left text-s">
                    <li>Hp: {pokemon.hp}</li>
                    <li>Attack: {pokemon.attack}</li>
                    <li>Defense: {pokemon.defense}</li>
                    <li>Special Attack: {pokemon.specialAtk}</li>
                    <li>Special Defense: {pokemon.specialDef}</li>
                    <li>Speed: {pokemon.speed}</li>
                    <li>Abilities: {abilities}</li>

                </ul>
                <p className="text-base text-center mb-auto">
                    
                </p>
            </div>
        </div>
        </div>
    );
}