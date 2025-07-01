import React, { useState } from 'react';

export default function PokemonCard() {
    const [isFlipped, setIsFlipped] = useState(false);
    const pokemon={
        pic:"https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/257.png",
        name:"Blaziken",
        cost:10,
    }   
    const handleFlip = () => {
        setIsFlipped(!isFlipped);
    };

    const handleImageError = (e) => {
        e.target.onerror = null;
        e.target.src = `https://placehold.co/150x150/cccccc/333333?text=No+Image`;
    };

    return (
        // Container of the whole thing : Sets perspective for effect and provides a clickable area
        <div
        className=" group h-80 w-60 rounded-lg shadow-lg relative cursor-pointer [perspective:1000px]" // Fixed size for the card
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
                <img
                    src={pokemon.pic}
                    alt={pokemon.name}
                    onError={handleImageError}
                    className="w-[100%] h-[100%] object-contain mb-4 bg-gray-100 p-2"
                />
                <div className='flex justify-between w-[100%]'>
                    {/* Pokémon Name */}
                    <h3 className="text-xl font-bold text-gray-800 mb-2 text-center">
                        {pokemon.name}
                    </h3>

                    {/* Pokémon Cost */}
                    <p className="text-lg font-semibold">
                        {pokemon.cost}
                    </p>
                </div>
                
            </div>

            {/* Back Face of the Card */}
            <div className="absolute inset-0 bg-gray-700 text-white rounded-lg p-4 flex flex-col [backface-visibility:hidden] [transform:rotateY(180deg)]">
                <h3 className="text-2xl font-bold mb-4 text-center">
                    {pokemon.name}
                </h3>
                <p className="text-base text-center mb-auto">
                    Super Cool Pokemon bruh
                </p>
                <ul className="text-left mt-4 ">
                    <li>Type: Fire</li>
                    <li>Ability: kicking or sum</li>
                    <li>HP: 69420</li>
                </ul>
            </div>
        </div>
        </div>
    );
}