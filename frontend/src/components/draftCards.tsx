import React from 'react';
import { formatPokemonName, formatAbilityName } from "../utils/nameFormatter";
import { DraftCardProps } from '../api/data_interfaces';
import pokeballIcon from '../assets/pokeball-icon.png';


// Helper function to get stat color based on value
const getStatColor = (value: number): string => {
    if (value <= 25) {
        return '#FF0000';
    } else if (value <= 60) {
        return '#FFA500';
    } else if (value <= 89) {
        return '#FFFF00';
    } else if (value <= 120) {
        return '#A0F555';
    } else if (value <= 199) {
        return '#23CD5E';
    } else {
        return '#02FFFF';
    }
};

// StatBar Component
interface StatBarProps {
    label: string;
    value: number;
}

const StatBar: React.FC<StatBarProps> = ({ label, value }) => {
    const maxWidth = 255; // Max possible stat value
    const barWidth = (value / maxWidth) * 100; // Calculate width as percentage
    const color = getStatColor(value);

    return (
        <div className="flex items-center gap-1">
            <span className="w-10 text-right text-xs font-medium">{label}:</span>
            <div className="flex-1 bg-gray-600 h-4">
                <div
                    className="h-4"
                    style={{ width: `${barWidth}%`, backgroundColor: color }}
                ></div>
            </div>
            <span className="w-6 text-left text-xs">{value}</span>
        </div>
    );
};

export default function PokemonCard({ pokemon, cost, onImageError, leaguePokemonId, addPokemonToWishlist, removePokemonFromWishlist, isPokemonInWishlist, isFlipped, onFlip, isDraftable, onDraft, isAvailable, isMyTurn }: DraftCardProps) {
    console.log(`DraftCard for ${pokemon.Name}: isMyTurn=${isMyTurn}, isAvailable=${isAvailable}, isDraftable=${isDraftable}`);
    const handleFlip = () => {
        onFlip(leaguePokemonId);
    };
    const types = pokemon.Types.map(t => {
        if (typeof t === 'string' && t.length > 0) {
            return t.charAt(0).toUpperCase() + t.slice(1);
        }
        return t;
    }).join(', ');
    const formattedAbilities = pokemon.Abilities.map(a => (
        <span key={a.Name} className={a.IsHidden ? 'text-gray-400 italic' : ''}>
            {formatAbilityName(a.Name)}
        </span>
    ));

    const isInWishlist = isPokemonInWishlist(leaguePokemonId);

    const handleWishlistToggle = (e: React.MouseEvent) => {
        e.stopPropagation(); // Prevent card from flipping
        if (isInWishlist) {
            removePokemonFromWishlist(leaguePokemonId);
        } else {
            addPokemonToWishlist(leaguePokemonId);
        }
    };

    return (
        <div
            className="group h-70 w-47 rounded-lg shadow-lg relative cursor-pointer [perspective:1000px]"
            onClick={handleFlip}
        >
            <div
                className={`
            relative
            w-full h-full
            transition-transform duration-700 ease-in-out
            [transform-style:preserve-3d]
            ${isFlipped ? '[transform:rotateY(180deg)]' : ''}
            `}
            >
                <div className={`absolute inset-0 bg-background-surface rounded-lg p-4 flex flex-col items-center justify-center [backface-visibility:hidden] ${!isAvailable ? 'opacity-50 grayscale' : ''}`}>
                    <div className="relative w-full h-[100%]">
                        <img
                            src={pokemon.Sprites.FrontDefault}
                            alt={pokemon.Name}
                            onError={onImageError}
                            className="w-[100%] h-[100%] object-contain mb-4 bg-background-tertiary p-2"
                        />
                        <p className="text-lg font-semibold absolute bottom-2 right-2 ">
                            {cost}
                        </p>
                    </div>
                    <div className='flex w-[100%] justify-between'>
                        <div>
                            <h3 className={`pb-0 mb-0 font-bold text-gray-800 text-left ${formatPokemonName(pokemon.Name).length > 12 ? 'text-base' : 'text-lg'}`}>
                                {formatPokemonName(pokemon.Name)}
                            </h3>
                            <p className='p-0 m-0 text-left text-sm text-gray-600'>{types}</p>
                        </div>
                        {isAvailable && ( // Only render buttons if available
                            isMyTurn ? ( // If it's my turn, show Pokeball (draft button)
                                <button
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        onDraft(leaguePokemonId);
                                    }}
                                    className="relative flex items-center align-center justify-center mt-4 h-7.5 w-7.5 rounded-full p-0 transition-all duration-150 hover:bg-gray-200 hover:shadow-lg"
                                >
                                    <img src={pokeballIcon} alt="draft mon" className="w-5 h-5 rounded-full transition-transform duration-150 transform hover:scale-125" />
                                </button>
                            ) : ( // Else, show Star (wishlist button)
                                <button
                                    onClick={handleWishlistToggle}
                                    className={`relative flex items-center align-center justify-center mt-4 h-7.5 w-7.5 rounded-full p-0 transition-all duration-150 focus:outline-none focus:ring-0
                                                ${isInWishlist ? 'bg-yellow-400 hover:bg-yellow-500' : 'bg-gray-200 hover:bg-gray-300'}`}
                                >
                                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5}
                                        className={`w-5 h-5 transition-colors duration-150
                                                     ${isInWishlist ? 'stroke-white' : 'stroke-gray-600'}`}
                                    >
                                        <path strokeLinecap="round" strokeLinejoin="round" d="M11.48 3.499a.562.562 0 0 1 1.04 0l2.125 5.111a.563.563 0 0 0 .475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 0 0-.182.557l1.285 5.385a.562.562 0 0 1-.84.61l-4.725-2.885a.562.562 0 0 0-.586 0L6.982 20.54a.562.562 0 0 1-.84-.61l1.285-5.386a.562.562 0 0 0-.182-.557L3.422 8.99a.562.562 0 0 1 .321-.989l5.518-.442a.563.563 0 0 0 .475-.345L11.48 3.5Z" />
                                    </svg>
                                </button>
                            )
                        )}
                    </div>

                </div>

                {/* Back Face of the Card */}
                <div className="text-l absolute inset-0 bg-gray-700 text-white rounded-lg p-4 flex flex-col [backface-visibility:hidden] [transform:rotateY(180deg)]">
                    <h3 className={`font-bold mb-4 text-center ${formatPokemonName(pokemon.Name).length > 12 ? 'text-base' : 'text-lg'}`}>
                        {formatPokemonName(pokemon.Name)}
                    </h3>
                    <div className="flex flex-col gap-1 w-full">
                        <StatBar label="HP" value={pokemon.Stats.Hp} />
                        <StatBar label="Att" value={pokemon.Stats.Attack} />
                        <StatBar label="Def" value={pokemon.Stats.Defense} />
                        <StatBar label="SpA" value={pokemon.Stats.SpecialAttack} />
                        <StatBar label="SpD" value={pokemon.Stats.SpecialAttack} />
                        <StatBar label="Spe" value={pokemon.Stats.Speed} />
                    </div>
                    <div className="text-left text-xs mt-4">
                        <p className="font-xs">
                            <span className='font-bold'>Abilities:</span>{' '}
                            {formattedAbilities.map((el, i) => (
                                <span key={i}>
                                    {i > 0 && ', '}
                                    {el}
                                </span>
                            ))}
                        </p>
                    </div>
                    <p className="text-base text-center mb-auto">

                    </p>
                </div>
            </div>
        </div>
    )
}
