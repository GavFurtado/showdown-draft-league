import { useState } from 'react';
import { DraftCardProps } from '../api/data_interfaces';

// Helper function to format names (replace hyphens with spaces and title case)
const formatName = (name: string): string => {
    return name
        .replace(/-/g, ' ')
        .split(' ')
        .map(word => word.charAt(0).toUpperCase() + word.slice(1))
        .join(' ');
};

// Helper function to get stat color based on value
const getStatColor = (value: number): string => {
    if (value <= 25) {
        return '#FF0000'; // Red
    } else if (value <= 60) {
        return '#FFA500'; // Orange
    } else if (value <= 89) {
        return '#FFFF00'; // Yellow
    } else if (value <= 120) {
        return '#A0F555'; // Lime Green
    } else if (value <= 199) {
        return '#23CD5E'; // Darker Green
    } else {
        return '#02FFFF'; // Cyan
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

export default function PokemonCard({ pokemon, cost, onImageError, leaguePokemonId, addPokemonToWishlist, removePokemonFromWishlist, isPokemonInWishlist }: DraftCardProps) {
    const [isFlipped, setIsFlipped] = useState(false);
    const handleFlip = () => {
        setIsFlipped(!isFlipped);
    };
    const types = pokemon.types.map(t => {
        if (typeof t === 'string' && t.length > 0) {
            return t.charAt(0).toUpperCase() + t.slice(1);
        }
        return t;
    }).join(', ');
    const formattedAbilities = pokemon.abilities.map(a => (
        <span key={a.name} className={a.is_hidden ? 'text-gray-400 italic' : ''}>
            {formatName(a.name)}
        </span>
    ));

    const isInWishlist = isPokemonInWishlist(leaguePokemonId);

    // console.log(`DraftCard: Rendering ${pokemon.name} (ID: ${leaguePokemonId}). isInWishlist: ${isPokemonInWishlist(leaguePokemonId)}`);
    const handleWishlistToggle = (e: React.MouseEvent) => {
        e.stopPropagation(); // Prevent card from flipping
        if (isInWishlist) {
            removePokemonFromWishlist(leaguePokemonId);
        } else {
            addPokemonToWishlist(leaguePokemonId);
        }
    };

    return (
        // Container of the whole thing : Sets perspective for effect and provides a clickable area
        <div
            className="group h-70 w-47 rounded-lg shadow-lg relative cursor-pointer [perspective:1000px]"
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
                            src={pokemon.sprites.front_default}
                            alt={pokemon.name}
                            onError={onImageError}
                            className="w-[100%] h-[100%] object-contain mb-4 bg-gray-100 p-2"
                        />
                        {/* Pokémon Cost */}
                        <p className="text-lg font-semibold absolute  bottom-2 right-2 ">
                            {cost}
                        </p>
                    </div>
                    <div className='flex w-[100%] justify-between'>
                        {/* Pokémon Name */}
                        <div>
                            <h3 className="text-lg pb-0 mb-0 font-bold text-gray-800 text-left">
                                {formatName(pokemon.name)}
                            </h3>
                            <p className='p-0 m-0 text-left text-sm text-gray-600'>{types}</p>
                        </div>
                        <button
                            onClick={handleWishlistToggle}
                            className={`relative flex items-center align-center justify-center mt-4 h-7.5 w-7.5 rounded-full p-0 transition-colors
                                ${isInWishlist ? 'bg-yellow-400 hover:bg-yellow-500' : 'bg-gray-200 hover:bg-gray-300'}
                            `}>
                            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
                                <path strokeLinecap="round" strokeLinejoin="round" d="M11.48 3.499a.562.562 0 0 1 1.04 0l2.125 5.111a.563.563 0 0 0 .475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 0 0-.182.557l1.285 5.385a.562.562 0 0 1-.84.61l-4.725-2.885a.562.562 0 0 0-.586 0L6.982 20.54a.562.562 0 0 1-.84-.61l1.285-5.386a.562.562 0 0 0-.182-.557L3.422 8.99a.562.562 0 0 1 .321-.989l5.518-.442a.563.563 0 0 0 .475-.345L11.48 3.5Z" />
                            </svg>

                        </button>
                    </div>

                </div>

                {/* Back Face of the Card */}
                <div className="absolute inset-0 bg-gray-700 text-white rounded-lg p-4 flex flex-col [backface-visibility:hidden] [transform:rotateY(180deg)]">
                    <h3 className="text-l font-bold mb-4 text-center">
                        {formatName(pokemon.name)}
                    </h3>
                    <div className="flex flex-col gap-1 w-full">
                        <StatBar label="HP" value={pokemon.stats.hp} />
                        <StatBar label="Att" value={pokemon.stats.attack} />
                        <StatBar label="Def" value={pokemon.stats.defense} />
                        <StatBar label="SpA" value={pokemon.stats["special_attack"]} />
                        <StatBar label="SpD" value={pokemon.stats["special_defense"]} />
                        <StatBar label="Spe" value={pokemon.stats.speed} />
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
    );
}
